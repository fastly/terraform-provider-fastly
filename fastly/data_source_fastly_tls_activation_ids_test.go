package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestAccDataSourceFastlyTLSActivationIds_basic(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))
	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)

	datasourceName := "data.fastly_tls_activation_ids.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSActivationIdsConfig(key, cert, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "1"),
					testAccTLSActivationIDIncluded(datasourceName, "fastly_tls_activation.test"),
				),
			},
		},
	})
}

func testAccTLSActivationIDIncluded(dataSourceName string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		d, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}

		for k, v := range d.Primary.Attributes {
			if k == "ids.#" {
				continue
			}
			if !strings.HasPrefix(k, "ids.") {
				continue
			}
			if v == r.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("unable to find private key %s in list of private key ids", r.Primary.ID)
	}
}

func testAccDataSourceFastlyTLSActivationIdsConfig(key, cert, domain string) string {
	name := acctest.RandomWithPrefix(testResourcePrefix)

	return fmt.Sprintf(
		`
resource "fastly_service_v1" "test" {
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
  depends_on = [fastly_service_v1.test]
}

data "fastly_tls_activation_ids" "subject" {
  certificate_id = fastly_tls_activation.test.certificate_id
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
