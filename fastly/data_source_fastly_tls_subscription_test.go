package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFastlyTLSSubscription_basic(t *testing.T) {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))

	resourceName := "data.fastly_tls_subscription.subject"
	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSSubscriptionConfigBasic(domain),
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
	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSSubscriptionConfigByID(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "common_name", domain),
					resource.TestCheckResourceAttr(resourceName, "state", "pending"),
				),
			},
		},
	})
}

func testAccDataSourceFastlyTLSSubscriptionConfigBasic(domain string) string {
	name := acctest.RandomWithPrefix(testResourcePrefix)

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
resource "fastly_tls_subscription" "test" {
  domains = [for domain in fastly_service_vcl.test.domain : domain.name]
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

func testAccDataSourceFastlyTLSSubscriptionConfigByID(domain string) string {
	name := acctest.RandomWithPrefix(testResourcePrefix)

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
resource "fastly_tls_subscription" "test" {
  domains = [for domain in fastly_service_vcl.test.domain : domain.name]
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
