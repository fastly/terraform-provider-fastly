package fastly

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	rnd "math/rand"
	"strings"
	"time"
)

const (
	emptyString = ""
	formatDigit = "%d"
	bitSize     = 2048
)

func generateKey() (string, error) {
	reader := rand.Reader

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return emptyString, err
	}
	privateKey := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	bytes := pem.EncodeToMemory(privateKey)

	return strings.TrimSpace(string(bytes)), nil
}

func generateKeyAndCert(SANs ...string) (string, string, error) {
	privKey, key, err := buildPrivateKey()
	if err != nil {
		return emptyString, emptyString, err
	}

	cert, err := buildCertificate(privKey, SANs...)
	if err != nil {
		return emptyString, emptyString, err
	}

	return key, cert, nil
}

func generateKeyAndMultipleCerts(SANs ...string) (string, string, string, error) {
	privKey, key, err := buildPrivateKey()
	if err != nil {
		return emptyString, emptyString, emptyString, err
	}

	cert, err := buildCertificate(privKey, SANs...)
	if err != nil {
		return emptyString, emptyString, emptyString, err
	}

	cert2, err := buildCertificate(privKey, SANs...)
	if err != nil {
		return emptyString, emptyString, emptyString, err
	}

	return key, cert, cert2, nil
}

func generateKeyAndCertWithCA(domains ...string) (string, string, string, error) {
	caCert, caPEM, caKey, err := buildCACertificate(domains...)
	if err != nil {
		return emptyString, emptyString, emptyString, err
	}

	privateKey, key, err := buildPrivateKey()
	if err != nil {
		return emptyString, emptyString, emptyString, err
	}

	cert, err := buildCertificateFromCA(caCert, privateKey, caKey, domains...)
	if err != nil {
		return emptyString, emptyString, emptyString, err
	}

	return key, cert, caPEM, nil
}

func buildPrivateKey() (*rsa.PrivateKey, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, emptyString, fmt.Errorf("failed to generate private key: %s", err)
	}
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, emptyString, fmt.Errorf("unable to marshal private key: %v", err)
	}
	keyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	}
	return key, strings.TrimSpace(string(pem.EncodeToMemory(keyBlock))), nil
}

func buildCertificate(privateKey *rsa.PrivateKey, domains ...string) (string, error) {
	now := time.Now()
	serialNumber := new(big.Int).SetInt64(rnd.Int63())
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			SerialNumber: fmt.Sprintf(formatDigit, serialNumber),
		},
		NotBefore:             now,
		NotAfter:              now.Add(24 * 90 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		IsCA:                  true,
		DNSNames:              domains,
	}

	if len(domains) != 0 {
		template.Subject.CommonName = domains[0]
	}

	c, err := formatCertificate(template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return emptyString, err
	}
	return c, nil
}

func buildCertificateFromCA(ca *x509.Certificate, privateKey *rsa.PrivateKey, caKey *rsa.PrivateKey, domains ...string) (string, error) {
	now := time.Now()
	serialNumber := new(big.Int).SetInt64(rnd.Int63())
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			SerialNumber: fmt.Sprintf(formatDigit, serialNumber),
		},
		NotBefore:             now,
		NotAfter:              now.Add(24 * 90 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		DNSNames:              domains,
	}

	if len(domains) != 0 {
		template.Subject.CommonName = domains[0]
	}

	c, err := formatCertificate(template, ca, &privateKey.PublicKey, caKey)
	if err != nil {
		return emptyString, err
	}
	return c, nil
}

func buildCACertificate(domains ...string) (*x509.Certificate, string, *rsa.PrivateKey, error) {
	now := time.Now()
	serialNumber := new(big.Int).SetInt64(rnd.Int63())
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			SerialNumber: fmt.Sprintf(formatDigit, serialNumber),
		},
		NotBefore:             now,
		NotAfter:              now.Add(24 * 90 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		IsCA:                  true,
		DNSNames:              domains,
	}

	privateKey, _, err := buildPrivateKey()
	if err != nil {
		return nil, emptyString, nil, err
	}

	c, err := formatCertificate(template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, emptyString, nil, err
	}
	return template, c, privateKey, nil
}

func formatCertificate(certificate *x509.Certificate, parent *x509.Certificate, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) (string, error) {
	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		certificate,
		parent,
		publicKey,
		privateKey,
	)
	if err != nil {
		return emptyString, err
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	return strings.TrimSpace(string(pem.EncodeToMemory(certBlock))), nil
}
