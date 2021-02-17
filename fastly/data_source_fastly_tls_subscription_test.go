package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceFastlyTLSSubscription_basic(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))

	resourceName := "data.fastly_tls_subscription.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSSubscriptionConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "certificate_authority", "lets-encrypt"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "common_name", domain),
					resource.TestCheckResourceAttr(resourceName, "state", "pending"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceFastlyTLSSubscription_byID(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))

	resourceName := "data.fastly_tls_subscription.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSSubscriptionConfig_byID(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "common_name", domain),
					resource.TestCheckResourceAttr(resourceName, "state", "pending"),
				),
			},
		},
	})
}

func testAccDataSourceFastlyTLSSubscriptionConfig_basic(domain string) string {
	name := acctest.RandomWithPrefix(testResourcePrefix)

	return fmt.Sprintf(`
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
data "fastly_tls_subscription" "subject" {
  domains = fastly_tls_subscription.test.domains
}
`,
		name,
		domain,
	)
}

func testAccDataSourceFastlyTLSSubscriptionConfig_byID(domain string) string {
	name := acctest.RandomWithPrefix(testResourcePrefix)

	return fmt.Sprintf(`
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
data "fastly_tls_subscription" "subject" {
  id = fastly_tls_subscription.test.id
}
`,
		name,
		domain,
	)
}
