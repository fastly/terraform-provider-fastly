package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTLSDomain_basic(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.example", name)

	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSDomainResources(key, cert, name),
			},
			{
				Config: testAccFastlyDataSourceTLSDomainWithDataSource(key, cert, domain, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.fastly_tls_domain.subject", "tls_certificate_ids.#", "1"),
					testAccTLSDomainCertIDIncluded("data.fastly_tls_domain.subject", "fastly_tls_certificate.example"),
				),
			},
		},
	})
}

// This can be replaced with `TestCheckTypeSetElemNestedAttrs` when using SDK 2.x
func testAccTLSDomainCertIDIncluded(dataSourceName string, resourceName string) resource.TestCheckFunc {
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
			if k == "tls_certificate_ids.#" {
				continue
			}
			if !strings.HasPrefix(k, "tls_certificate_ids.") {
				continue
			}
			if v == r.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("unable to find certificate %s in list of certificate ids", r.Primary.ID)
	}
}

func testAccFastlyDataSourceTLSDomainResources(key string, cert string, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  key_pem = <<EOF
%[1]s
EOF
  name = "%[3]s"
}
resource "fastly_tls_certificate" "example" {
  name = fastly_tls_private_key.key.name
  certificate_body = <<EOF
%[2]s
EOF
}
`, key, cert, name)
}

func testAccFastlyDataSourceTLSDomainWithDataSource(key string, cert string, domain string, name string) string {
	return fmt.Sprintf(`
%s
data "fastly_tls_domain" "subject" {
  domain = "%s"
}
`, testAccFastlyDataSourceTLSDomainResources(key, cert, name), domain)
}
