package fastly

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"

	"github.com/fastly/go-fastly/v11/fastly"
)

func init() {
	resource.AddTestSweepers("fastly_tls_activation", &resource.Sweeper{
		Name: "fastly_tls_activation",
		F:    testSweepTLSActivation,
	})
}

func TestAccFastlyTLSActivation_basic(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))
	key, cert, cert2, err := generateKeyAndMultipleCerts(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)
	cert2 = strings.ReplaceAll(cert2, "\n", `\n`)

	name := acctest.RandomWithPrefix(testResourcePrefix)
	updatedName := acctest.RandomWithPrefix(testResourcePrefix)

	resourceName := "fastly_tls_activation.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccFastlyTLSActivationCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyTLSActivationConfig(name, name, key, name, cert, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_id"),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					testAccFastlyTLSActivationCheckExists(resourceName),
				),
			},
			{
				Config: testAccFastlyTLSActivationConfig(name, name, key, updatedName, cert2, domain),
				Check:  testAccFastlyTLSActivationCheckExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyTLSActivation_mTLS(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))
	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	_, mtlsCert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)
	name := acctest.RandomWithPrefix(testResourcePrefix)

	resourceName := "fastly_tls_activation.mtls_test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccFastlyTLSActivationCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyTLSActivationMTLSConfig(name, name, key, name, cert, domain, mtlsCert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_id"),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					testAccFastlyTLSActivationCheckExists(resourceName),
				),
			},
			{
				Config: testAccFastlyTLSActivationMTLSConfig(name, name, key, name, cert, domain, mtlsCert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_id"),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "mutual_authentication_id"),
					testAccFastlyTLSActivationCheckExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFastlyTLSActivationConfig(serviceName, keyName, key, certName, cert, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "test" {
  name = "%s"

  domain {
    name = "%s"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_private_key" "test" {
  key_pem = "%s"
  name = "%s"
}

resource "fastly_tls_certificate" "test" {
  certificate_body = "%s"
  name = "%s"
  depends_on = [fastly_tls_private_key.test]
}

resource "fastly_tls_activation" "test" {
  certificate_id = fastly_tls_certificate.test.id
  domain = "%s"
  depends_on = [fastly_service_vcl.test]
}
`, serviceName, domain, key, keyName, cert, certName, domain)
}

func testAccFastlyTLSActivationMTLSConfig(serviceName, keyName, key, certName, cert, domain, certBundle string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "mtls_test" {
  name = "%s"

  domain {
    name = "%s"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_private_key" "mtls_test" {
  key_pem = "%s"
  name = "%s"
}

resource "fastly_tls_certificate" "mtls_test" {
  certificate_body = "%s"
  name = "%s"
  depends_on = [fastly_tls_private_key.mtls_test]
}

resource "fastly_tls_activation" "mtls_test" {
  certificate_id = fastly_tls_certificate.mtls_test.id
  domain = "%s"
  depends_on = [fastly_service_vcl.mtls_test]
  mutual_authentication_id = fastly_tls_mutual_authentication.mtls_test.id
}

resource "fastly_tls_mutual_authentication" "mtls_test" {
  cert_bundle = <<EOF
%s
EOF
  name = "example_mtls"
}
`, serviceName, domain, key, keyName, cert, certName, domain, certBundle)
}

func testAccFastlyTLSActivationCheckExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn

		r, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		_, err := conn.GetTLSActivation(context.TODO(), &fastly.GetTLSActivationInput{
			ID: r.Primary.ID,
		})
		return err
	}
}

func testAccFastlyTLSActivationCheckDestroy(state *terraform.State) error {
	for _, resourceState := range state.RootModule().Resources {
		if resourceState.Type != "fastly_tls_activation" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		activations, err := conn.ListTLSActivations(context.TODO(), &fastly.ListTLSActivationsInput{})
		if err != nil {
			return fmt.Errorf(
				"[WARN] Error listing TLS activations when deleting activation %s: %w",
				resourceState.Primary.ID,
				err,
			)
		}

		for _, activation := range activations {
			if activation.ID == resourceState.Primary.ID {
				return fmt.Errorf(
					"[WARN] Tried disabling TLS activation (%s) but was still found",
					resourceState.Primary.ID,
				)
			}
		}
	}
	return nil
}

func testSweepTLSActivation(region string) error {
	client, diagnostics := sharedClientForRegion(region)
	if diagnostics.HasError() {
		return diagToErr(diagnostics)
	}

	activations, err := client.ListTLSActivations(context.TODO(), &fastly.ListTLSActivationsInput{PageSize: 1000})
	if err != nil {
		return err
	}

	for _, activation := range activations {
		if !strings.HasPrefix(activation.Domain.ID, testResourcePrefix) {
			continue
		}

		err = client.DeleteTLSActivation(context.TODO(), &fastly.DeleteTLSActivationInput{ID: activation.ID})
		if err != nil {
			return err
		}
	}

	return nil
}
