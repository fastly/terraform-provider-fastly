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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"fastly": testAccProvider,
	}
}

func generateKey() (string, error) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return "", err
	}
	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	bytes := pem.EncodeToMemory(privateKey)

	return string(bytes), nil
}

func generateKeyAndCert() (key, cert string, err error) {
	now := time.Now()
	serialNumber := new(big.Int).SetInt64(rnd.Int63())
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			SerialNumber: fmt.Sprintf("%d", serialNumber),
		},
		NotBefore:             now,
		NotAfter:              now.Add(24 * 365 * 20 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		IsCA:                  true,
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("Failed to generate private key: %s", err)
		return
	}
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		err = fmt.Errorf("Unable to marshal private key: %v", err)
		return
	}
	keyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	}
	key = strings.TrimSpace(string(pem.EncodeToMemory(keyBlock)))

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privKey.PublicKey,
		privKey,
	)
	if err != nil {
		err = fmt.Errorf("Failed to create certificate: %s", err)
		return
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	cert = strings.TrimSpace(string(pem.EncodeToMemory(certBlock)))

	return
}

func generateKeyAndCertWithSan(san ...string) (key, cert string, err error) {
	now := time.Now()
	serialNumber := new(big.Int).SetInt64(rnd.Int63())
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			SerialNumber: fmt.Sprintf("%d", serialNumber),
		},
		NotBefore:             now,
		NotAfter:              now.Add(24 * 365 * 20 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		IsCA:                  true,
		DNSNames:              san,
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("Failed to generate private key: %s", err)
		return
	}
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		err = fmt.Errorf("Unable to marshal private key: %v", err)
		return
	}
	keyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	}
	key = strings.TrimSpace(string(pem.EncodeToMemory(keyBlock)))

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privKey.PublicKey,
		privKey,
	)
	if err != nil {
		err = fmt.Errorf("Failed to create certificate: %s", err)
		return
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	cert = strings.TrimSpace(string(pem.EncodeToMemory(certBlock)))

	return
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("FASTLY_API_KEY"); v == "" {
		t.Fatal("FASTLY_API_KEY must be set for acceptance tests")
	}
}
