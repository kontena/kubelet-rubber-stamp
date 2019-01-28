package certificatesigningrequest

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"reflect"
	"strings"

	authorization "k8s.io/api/authorization/v1beta1"
	capi "k8s.io/api/certificates/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type csrRecognizer struct {
	recognize      func(csr *capi.CertificateSigningRequest, x509cr *x509.CertificateRequest) bool
	permission     authorization.ResourceAttributes
	successMessage string
}

func recognizers() []csrRecognizer {
	recognizers := []csrRecognizer{
		{
			recognize:      isNodeServingCert,
			permission:     authorization.ResourceAttributes{Group: "certificates.k8s.io", Resource: "certificatesigningrequests", Verb: "create"},
			successMessage: "Auto approving kubelet serving certificate after SubjectAccessReview.",
		},
	}
	return recognizers
}

func isNodeServingCert(csr *capi.CertificateSigningRequest, x509cr *x509.CertificateRequest) bool {
	if !reflect.DeepEqual([]string{"system:nodes"}, x509cr.Subject.Organization) {
		log.Printf("Org does not match: %s\n", x509cr.Subject.Organization)
		return false
	}
	if (len(x509cr.DNSNames) < 1) || (len(x509cr.IPAddresses) < 1) {
		return false
	}
	if !hasExactUsages(csr, kubeletClientUsages) {
		log.Println("Usage does not match")
		return false
	}
	if !strings.HasPrefix(x509cr.Subject.CommonName, "system:node:") {
		log.Printf("CN does not match: %s\n", x509cr.Subject.CommonName)
		return false
	}
	return true
}

func hasExactUsages(csr *capi.CertificateSigningRequest, usages []capi.KeyUsage) bool {
	if len(usages) != len(csr.Spec.Usages) {
		return false
	}

	usageMap := map[capi.KeyUsage]struct{}{}
	for _, u := range usages {
		usageMap[u] = struct{}{}
	}

	for _, u := range csr.Spec.Usages {
		if _, ok := usageMap[u]; !ok {
			return false
		}
	}

	return true
}

var kubeletClientUsages = []capi.KeyUsage{
	capi.UsageKeyEncipherment,
	capi.UsageDigitalSignature,
	capi.UsageServerAuth,
}

// Add creates a new CertificateSigningRequest Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCertificateSigningRequest{client: mgr.GetClient(), scheme: mgr.GetScheme(), clientset: clientset.NewForConfigOrDie(mgr.GetConfig())}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("certificatesigningrequest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CertificateSigningRequest
	err = c.Watch(&source.Kind{Type: &capi.CertificateSigningRequest{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCertificateSigningRequest{}

// ReconcileCertificateSigningRequest reconciles a CertificateSigningRequest object
type ReconcileCertificateSigningRequest struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	// TODO: Need to refactor to use only single client
	clientset clientset.Interface
}

// Reconcile reads that state of the cluster for a CertificateSigningRequest object and makes changes based on the state read
// and what is in the CertificateSigningRequest.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCertificateSigningRequest) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling CertificateSigningRequest %s/%s\n", request.Namespace, request.Name)

	// Fetch the CertificateSigningRequest instance
	csr := &capi.CertificateSigningRequest{}
	err := r.client.Get(context.TODO(), request.NamespacedName, csr)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if len(csr.Status.Certificate) != 0 {
		log.Println("CSR already has a certificate, ignoring")
		return reconcile.Result{}, nil
	}

	if approved, denied := getCertApprovalCondition(&csr.Status); approved || denied {
		log.Printf("CSR already has a approval status: %v\n", csr.Status)
		return reconcile.Result{}, nil
	}

	x509cr, err := parseCSR(csr)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("unable to parse csr %q: %v", csr.Name, err)
	}

	tried := []string{}

	for _, recognizer := range recognizers() {
		if !recognizer.recognize(csr, x509cr) {
			continue
		}

		tried = append(tried, recognizer.permission.Subresource)

		approved, err := r.authorize(csr, recognizer.permission)
		if err != nil {
			log.Printf("SubjectAccessReview failed:%s\n", err)
			return reconcile.Result{}, err
		}

		if approved {
			appendApprovalCondition(csr, recognizer.successMessage)
			_, err = r.clientset.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(csr)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("error updating approval for csr: %v", err)
			}
		} else {
			log.Printf("SubjectAccessReview not succesfull for CSR %s\n", request.NamespacedName)
			return reconcile.Result{}, fmt.Errorf("SubjectAccessReview not succesfull")
		}

		return reconcile.Result{}, nil

	}

	if len(tried) != 0 {
		return reconcile.Result{}, fmt.Errorf("recognized csr %q as %v but subject access review was not approved", csr.Name, tried)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileCertificateSigningRequest) authorize(csr *capi.CertificateSigningRequest, rattrs authorization.ResourceAttributes) (bool, error) {
	log.Printf("Authorizing CSR %s\n", csr.ObjectMeta.Name)
	extra := make(map[string]authorization.ExtraValue)
	for k, v := range csr.Spec.Extra {
		extra[k] = authorization.ExtraValue(v)
	}

	sar := &authorization.SubjectAccessReview{
		Spec: authorization.SubjectAccessReviewSpec{
			User:               csr.Spec.Username,
			UID:                csr.Spec.UID,
			Groups:             csr.Spec.Groups,
			Extra:              extra,
			ResourceAttributes: &rattrs,
		},
	}
	sar, err := r.clientset.AuthorizationV1beta1().SubjectAccessReviews().Create(sar)
	if err != nil {
		return false, err
	}
	log.Printf("SAR status: %v", sar)
	return sar.Status.Allowed, nil
}

func appendApprovalCondition(csr *capi.CertificateSigningRequest, message string) {
	csr.Status.Conditions = append(csr.Status.Conditions, capi.CertificateSigningRequestCondition{
		Type:    capi.CertificateApproved,
		Reason:  "AutoApproved",
		Message: message,
	})
}
