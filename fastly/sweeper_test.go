package fastly

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const testResourcePrefix = "tf-test"

var sweeperClients map[string]*fastly.Client

func TestMain(m *testing.M) {
	sweeperClients = make(map[string]*fastly.Client)
	resource.TestMain(m)
}

func sharedClientForRegion(region string) (*fastly.Client, error) {
	if client, ok := sweeperClients[region]; ok {
		return client, nil
	}

	if os.Getenv("FASTLY_API_KEY") == "" {
		return nil, fmt.Errorf("must provide environment variables 'FASTLY_API_KEY'")
	}

	url := fastly.DefaultEndpoint
	if v := os.Getenv("FASTLY_API_URL"); v != "" {
		url = v
	}
	c := Config{
		ApiKey:           os.Getenv("FASTLY_API_KEY"),
		BaseURL:          url,
		terraformVersion: "test-sweepers",
	}

	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	sweeperClients[region] = client.conn

	return client.conn, nil
}

func init() {
	resource.AddTestSweepers("fastly_tls_private_key", &resource.Sweeper{
		Name:         "fastly_tls_private_key",
		Dependencies: []string{"fastly_tls_certificate"}, // in case a private key is used by a certificate
		F:            testSweepTLSPrivateKeys,
	})
	resource.AddTestSweepers("fastly_tls_certificate", &resource.Sweeper{
		Name: "fastly_tls_certificate",
		F:    testSweepTLSCertificates,
	})
}

func testSweepTLSCertificates(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	certificates, err := client.ListCustomTLSCertificates(&fastly.ListCustomTLSCertificatesInput{PageSize: 1000})
	if err != nil {
		return err
	}

	for _, certificate := range certificates {
		if !strings.HasPrefix(certificate.Name, testResourcePrefix) {
			continue
		}

		err := client.DeleteCustomTLSCertificate(&fastly.DeleteCustomTLSCertificateInput{ID: certificate.ID})
		if err != nil {
			return err
		}
	}

	return nil
}

func testSweepTLSPrivateKeys(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	keys, err := client.ListPrivateKeys(&fastly.ListPrivateKeysInput{PageSize: 1000})
	if err != nil {
		return err
	}

	for _, key := range keys {
		if !strings.HasPrefix(key.Name, testResourcePrefix) {
			continue
		}

		err := client.DeletePrivateKey(&fastly.DeletePrivateKeyInput{ID: key.ID})
		if err != nil {
			return err
		}
	}

	return nil
}
