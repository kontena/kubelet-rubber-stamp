package certificatesigningrequest

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"testing"

	capi "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeServingCert_org(t *testing.T) {

	var csr = capi.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "test-csr"},
		Spec: capi.CertificateSigningRequestSpec{
			Usages: kubeletServerUsages,
		},
	}

	var x509cr = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"foobar"},
			CommonName:   "system:node:node-01",
		},
		DNSNames:    []string{"foobar"},
		IPAddresses: []net.IP{net.ParseIP("1.2.3.4")},
	}

	v := isNodeServingCert(&csr, &x509cr)

	if v != false {
		t.Error("Only 'system:nodes' accepted as org")
	}
}

func TestNodeServingCert_NoDNSOrIP(t *testing.T) {
	var csr = capi.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "test-csr"},
		Spec: capi.CertificateSigningRequestSpec{
			Usages: kubeletServerUsages,
		},
	}

	var x509cr = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"foobar"},
			CommonName:   "system:node:node-01",
		},
	}

	v := isNodeServingCert(&csr, &x509cr)

	if v != false {
		t.Error("Need at least one DNS name and IPAddress")
	}
}

func TestNodeServingCert_OnlyIP(t *testing.T) {
	var csr = capi.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "test-csr"},
		Spec: capi.CertificateSigningRequestSpec{
			Usages: kubeletServerUsages,
		},
	}

	var x509cr = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"foobar"},
			CommonName:   "system:node:node-01",
		},
		IPAddresses: []net.IP{net.ParseIP("1.2.3.4")},
	}

	v := isNodeServingCert(&csr, &x509cr)

	if v != false {
		t.Error("Need at least one DNS name and IPAddress")
	}
}

func TestNodeServingCert_OnlyDNS(t *testing.T) {
	var csr = capi.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "test-csr"},
		Spec: capi.CertificateSigningRequestSpec{
			Usages: kubeletServerUsages,
		},
	}

	var x509cr = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"foobar"},
			CommonName:   "system:node:node-01",
		},
		DNSNames: []string{"foobar"},
	}

	v := isNodeServingCert(&csr, &x509cr)

	if v != false {
		t.Error("Need at least one DNS name and IPAddress")
	}
}

func TestNodeServingCert_unmatchingUsages(t *testing.T) {
	var csr = capi.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "test-csr"},
		Spec: capi.CertificateSigningRequestSpec{
			Usages: []capi.KeyUsage{"foo", "bar"},
		},
	}

	var x509cr = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"foobar"},
			CommonName:   "system:node:node-01",
		},
		DNSNames: []string{"foobar"},
	}

	v := isNodeServingCert(&csr, &x509cr)

	if v != false {
		t.Errorf("Usages need to match %v\n", kubeletServerUsages)
	}
}

func TestNodeServingCert_unmatchingCN(t *testing.T) {
	var csr = capi.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "test-csr"},
		Spec: capi.CertificateSigningRequestSpec{
			Usages: []capi.KeyUsage{"foo", "bar"},
		},
	}

	var x509cr = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"foobar"},
			CommonName:   "system:foo:node-01",
		},
		DNSNames: []string{"foobar"},
	}

	v := isNodeServingCert(&csr, &x509cr)

	if v != false {
		t.Errorf("CN need to match 'system:node:*'")
	}
}
