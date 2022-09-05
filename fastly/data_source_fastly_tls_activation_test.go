package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccDataSourceFastlyTLSActivation_basic(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))
	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)

	resourceName := "data.fastly_tls_activation.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSActivationConfig(key, cert, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_id"),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func TestAccFastlyDataSourceFastlyTLSActivation_byID(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))
	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSActivationConfigByID(key, cert, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.fastly_tls_activation.subject", "domain", domain),
				),
			},
		},
	})
}

func testAccDataSourceFastlyTLSActivationConfig(key, cert, domain string) string {
	name := acctest.RandomWithPrefix(testResourcePrefix)

	return fmt.Sprintf(
		`
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

data "fastly_tls_activation" "test" {
  domain = fastly_tls_activation.test.domain
}
`,
		name,
		domain,
		key,
		name,
		cert,
		name,
		domain,
	)
}

func testAccFastlyDataSourceTLSActivationConfigByID(key, cert, domain string) string {
	name := acctest.RandomWithPrefix(testResourcePrefix)

	return fmt.Sprintf(
		`
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

data "fastly_tls_activation" "subject" {
  id = fastly_tls_activation.test.id
}
`,
		name,
		domain,
		key,
		name,
		cert,
		name,
		domain,
	)
}
