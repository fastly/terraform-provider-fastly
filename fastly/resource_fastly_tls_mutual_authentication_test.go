package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyMTLS_basic(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))
	key, cert, cert2, err := generateKeyAndMultipleCerts(domain)
	require.NoError(t, err)
	_, mtlsCert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)
	cert2 = strings.ReplaceAll(cert2, "\n", `\n`)

	name := acctest.RandomWithPrefix(testResourcePrefix)
	updatedName := acctest.RandomWithPrefix(testResourcePrefix)

	resourceTLSActivationName := "fastly_tls_activation.test"
	resourceMTLSName := "fastly_tls_mutual_authentication.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccFastlyTLSActivationCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyMTLSConfig(name, name, key, name, cert, domain, mtlsCert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceTLSActivationName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceTLSActivationName, "configuration_id"),
					resource.TestCheckResourceAttr(resourceTLSActivationName, "domain", domain),
					resource.TestCheckResourceAttrSet(resourceTLSActivationName, "created_at"),
					testAccFastlyTLSActivationCheckExists(resourceTLSActivationName),
					resource.TestCheckResourceAttr(resourceMTLSName, "name", "example_mtls"),
					resource.TestCheckResourceAttr(resourceMTLSName, "enforced", "false"),
				),
			},
			{
				Config: testAccFastlyMTLSConfig(name, name, key, updatedName, cert2, domain, mtlsCert),
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyTLSActivationCheckExists(resourceTLSActivationName),
					resource.TestCheckResourceAttr(resourceMTLSName, "name", "example_mtls"),
					resource.TestCheckResourceAttr(resourceMTLSName, "enforced", "false"),
				),
			},
			{
				ResourceName:      resourceTLSActivationName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceMTLSName,
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"cert_bundle", "activation_id"},
			},
		},
	})
}

func testAccFastlyMTLSConfig(serviceName, keyName, key, certName, cert, domain, certBundle string) string {
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

resource "fastly_tls_mutual_authentication" "test" {
  activation_id = fastly_tls_activation.test.id
  cert_bundle = <<EOF
%s
EOF
  name = "example_mtls"
}
`, serviceName, domain, key, keyName, cert, certName, domain, certBundle)
}
