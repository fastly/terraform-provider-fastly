package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFastlyTLSSubscriptionIds_basic(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("tf-test-%s.com", name)

	datasourceName := "data.fastly_tls_subscription_ids.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSSubscriptionIdsConfigResources(name, domain),
			},
			{
				Config: testAccDataSourceFastlyTLSSubscriptionIdsConfigWithDataSource(name, domain),
				Check: resource.TestCheckTypeSetElemAttrPair(
					datasourceName, "ids.*",
					"fastly_tls_subscription.test", "id",
				),
			},
		},
	})
}

func testAccDataSourceFastlyTLSSubscriptionIdsConfigResources(name, domain string) string {
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
resource "fastly_tls_subscription" "test" {
  domains = [for domain in fastly_service_vcl.test.domain : domain.name]
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
