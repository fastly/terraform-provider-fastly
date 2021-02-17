package fastly

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					resource.TestCheckResourceAttr(resourceName, "managed_dns_challenge.%", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "managed_http_challenges.#"),
					resource.TestCheckResourceAttr(resourceName, "common_name", domain),
					testAccResourceFastlyTLSSubscriptionExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config:      testAccResourceFastlyTLSSubscriptionConfig_invalidCommonName(),
				ExpectError: regexp.MustCompile("Domain specified as common_name.*"),
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
		r, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn

		_, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
			ID: r.Primary.ID,
		})
		return err
	}
}

func testAccResourceFastlyTLSSubscriptionConfig_invalidCommonName() string {
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))
	commonName := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))
	return fmt.Sprintf(`
resource "fastly_tls_subscription" "subject" {
  domains = ["%s"]
  common_name = "%s"
  certificate_authority = "lets-encrypt"
}
`, domain, commonName)
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
