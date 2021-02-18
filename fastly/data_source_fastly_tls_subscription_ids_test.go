package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strings"
	"testing"
)

func TestAccDataSourceFastlyTLSSubscriptionIds_basic(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("tf-test-%s.com", name)

	datasourceName := "data.fastly_tls_subscription_ids.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSSubscriptionIdsConfigResources(name, domain),
			},
			{
				Config: testAccDataSourceFastlyTLSSubscriptionIdsConfigWithDataSource(name, domain),
				Check:  testAccTLSSubscriptionIDIncluded(datasourceName, "fastly_tls_subscription.test"),
			},
		},
	})
}

func testAccTLSSubscriptionIDIncluded(dataSourceName string, resourceName string) resource.TestCheckFunc {
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

func testAccDataSourceFastlyTLSSubscriptionIdsConfigResources(name, domain string) string {
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
resource "fastly_tls_subscription" "test" {
  domains = [for domain in fastly_service_v1.test.domain : domain.name]
  certificate_authority = "lets-encrypt"
}
`,
		name,
		domain,
	)
}

func testAccDataSourceFastlyTLSSubscriptionIdsConfigWithDataSource(name, domain string) string {
	return fmt.Sprintf(`
%s
data "fastly_tls_subscription_ids" "subject" {}
`, testAccDataSourceFastlyTLSSubscriptionIdsConfigResources(name, domain))
}
