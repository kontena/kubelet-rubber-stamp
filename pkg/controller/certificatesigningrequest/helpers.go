package certificatesigningrequest

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"k8s.io/klog"
	"reflect"
	"strings"

	capi "k8s.io/api/certificates/v1"
)

func getCertApprovalCondition(status *capi.CertificateSigningRequestStatus) (approved bool, denied bool) {
	for _, c := range status.Conditions {
		if c.Type == capi.CertificateApproved {
			approved = true
		}
		if c.Type == capi.CertificateDenied {
			denied = true
		}
	}
	return
}

func isApproved(csr *capi.CertificateSigningRequest) bool {
	approved, denied := getCertApprovalCondition(&csr.Status)
	return approved && !denied
}

// parseCSR extracts the CSR from the API object and decodes it.
func parseCSR(obj *capi.CertificateSigningRequest) (*x509.CertificateRequest, error) {
	// extract PEM from request object
	pemBytes := obj.Spec.Request
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, errors.New("PEM block type must be CERTIFICATE REQUEST")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}
	return csr, nil
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

var kubeletServerUsages = []capi.KeyUsage{
	capi.UsageDigitalSignature,
	capi.UsageServerAuth,
}

func isNodeServingCert(csr *capi.CertificateSigningRequest, x509cr *x509.CertificateRequest) bool {
	if !reflect.DeepEqual([]string{"system:nodes"}, x509cr.Subject.Organization) {
		klog.Warningf("Org does not match: %s", x509cr.Subject.Organization)
		return false
	}
	if (len(x509cr.DNSNames) < 1) || (len(x509cr.IPAddresses) < 1) {
		return false
	}
	if !hasExactUsages(csr, kubeletServerUsages) {
		klog.V(2).Info("Usage does not match")
		return false
	}
	if !strings.HasPrefix(x509cr.Subject.CommonName, "system:node:") {
		klog.Warningf("CN does not start with 'system:node': %s", x509cr.Subject.CommonName)
		return false
	}
	if csr.Spec.Username != x509cr.Subject.CommonName {
		klog.Warningf("x509 CN %q doesn't match CSR username %q", x509cr.Subject.CommonName, csr.Spec.Username)
		return false
	}
	return true
}
