package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("fastly_tls_subscription", &resource.Sweeper{
		Name: "fastly_tls_subscription",
		F:    testSweepTLSSubscription,
	})
}

func TestAccResourceFastlyTLSSubscription(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.test", name)

	resourceName := "fastly_tls_subscription.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFastlyTLSSubscriptionConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "configuration_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_authorization_challenges.#"),
					testAccResourceFastlyTLSSubscriptionExists(resourceName),
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

func testAccResourceFastlyTLSSubscriptionConfig(name, domain string) string {
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
resource "fastly_tls_subscription" "subject" {
  domains = [for domain in fastly_service_v1.test.domain : domain.name]
  certificate_authority = "lets-encrypt"
}
`, name, domain)
}

func testAccResourceFastlyTLSSubscriptionExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		r := state.RootModule().Resources[resourceName]

		conn := testAccProvider.Meta().(*FastlyClient).conn

		_, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
			ID: r.Primary.ID,
		})
		return err
	}
}

func testSweepTLSSubscription(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	subscriptions, err := client.ListTLSSubscriptions(&fastly.ListTLSSubscriptionsInput{PageSize: 1000})
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		for _, domain := range subscription.Domains {
			if !strings.HasPrefix(domain.ID, testResourcePrefix) {
				continue
			}

			err = client.DeleteTLSSubscription(&fastly.DeleteTLSSubscriptionInput{
				ID: subscription.ID,
			})
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}
