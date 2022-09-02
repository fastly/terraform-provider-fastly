package fastly

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v6/fastly"
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
	domain1 := fmt.Sprintf("%s.test", name)
	domain2 := fmt.Sprintf("%salt.test", name)
	domain2Bad := fmt.Sprintf("%sALT.test", name)
	commonName1 := domain1
	commonName2 := domain2
	var subscriptionID string

	resourceName := "fastly_tls_subscription.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFastlyTLSSubscriptionConfig(name, domain1, domain2, commonName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "configuration_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "managed_dns_challenge.%", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "managed_http_challenges.#"),
					resource.TestCheckResourceAttr(resourceName, "common_name", domain1),
					testAccResourceFastlyTLSSubscriptionExists(resourceName, &subscriptionID),
				),
			},
			{
				Config: testAccResourceFastlyTLSSubscriptionConfig(name, domain1, domain2, commonName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "common_name", domain2),
					// subscription is "updated" so the subscription ID should remain the same
					resource.TestCheckResourceAttrPtr(resourceName, "id", &subscriptionID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "force_update"},
			},
			{
				Config:      testAccResourceFastlyTLSSubscriptionConfigInvalidCommonName(),
				ExpectError: regexp.MustCompile(`Please add \S+ to an active service to begin TLS enablement`),
			},
			{
				Config:      testAccResourceFastlyTLSSubscriptionConfig(name, domain1, domain2Bad, commonName2),
				ExpectError: regexp.MustCompile("must not contain uppercase letters"),
			},
		},
	})
}

func testAccResourceFastlyTLSSubscriptionConfig(name, domain1, domain2, commonName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "test" {
  name = "%s"

  domain {
    name = "%s"
  }

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
  domains = [for domain in fastly_service_vcl.test.domain : domain.name]
  common_name = "%s"
  certificate_authority = "lets-encrypt"
}
`, name, domain1, domain2, commonName)
}

func testAccResourceFastlyTLSSubscriptionExists(resourceName string, id *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		r, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		*id = r.Primary.ID
		conn := testAccProvider.Meta().(*APIClient).conn

		_, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
			ID: r.Primary.ID,
		})
		return err
	}
}

func testAccResourceFastlyTLSSubscriptionConfigInvalidCommonName() string {
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
	client, diagnostics := sharedClientForRegion(region)
	if diagnostics.HasError() {
		return diagToErr(diagnostics)
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
