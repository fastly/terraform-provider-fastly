package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/domains"
)

func TestAccFastlyRoutingConfigDomainLink_Basic(t *testing.T) {
	suffix := acctest.RandString(6)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoutingConfigDomainLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingConfigDomainLinkConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("fastly_routing_config_domain_link.example", "domain_id"),
					resource.TestCheckResourceAttrSet("fastly_routing_config_domain_link.example", "routing_config_id"),
				),
			},
			{
				ResourceName:      "fastly_routing_config_domain_link.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyRoutingConfigDomainLink_Drift verifies that if the link is
// removed out-of-band (e.g. by unlinking the domain directly through the
// API), Read drops the resource from state instead of erroring or leaving a
// stale link behind, so the next plan proposes recreating it.
func TestAccFastlyRoutingConfigDomainLink_Drift(t *testing.T) {
	suffix := acctest.RandString(6)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoutingConfigDomainLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingConfigDomainLinkConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("fastly_routing_config_domain_link.example", "domain_id"),
					testAccRoutingConfigDomainLinkUnlink("fastly_routing_config_domain_link.example"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccRoutingConfigDomainLinkConfig(suffix),
				Check:              resource.TestCheckResourceAttrSet("fastly_routing_config_domain_link.example", "domain_id"),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccRoutingConfigDomainLinkUnlink unlinks the domain identified by the
// named resource's `domain_id` directly through the API, simulating drift
// that happened outside of Terraform.
func testAccRoutingConfigDomainLinkUnlink(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		domainID := rs.Primary.Attributes["domain_id"]
		if domainID == "" {
			return fmt.Errorf("no domain_id set for %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		_, err := domains.Update(context.Background(), conn, &domains.UpdateInput{
			DomainID:               new(domainID),
			RoutingConfigurationID: gofastly.NullValue[string](),
		})
		if err != nil {
			return fmt.Errorf("error unlinking domain %s: %w", domainID, err)
		}
		return nil
	}
}

func testAccRoutingConfigDomainLinkConfig(suffix string) string {
	return fmt.Sprintf(`
resource "fastly_routing_config" "example" {
  name = "tf-test-routing-config-%s"
}

resource "fastly_domain" "domain" {
  fqdn = "test-%s.example.com"

  lifecycle {
    ignore_changes = [service_id]
  }
}

resource "fastly_routing_config_domain_link" "example" {
  domain_id         = fastly_domain.domain.domain_id
  routing_config_id = fastly_routing_config.example.routing_config_id
}
`, suffix, suffix)
}

func testAccCheckRoutingConfigDomainLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_routing_config_domain_link" {
			continue
		}

		domainID := rs.Primary.ID
		if domainID == "" {
			continue
		}

		input := &domains.GetInput{
			DomainID: new(domainID),
		}

		domain, err := domains.Get(context.Background(), conn, input)
		if err != nil {
			if httpErr, ok := err.(*gofastly.HTTPError); ok && httpErr.StatusCode == 404 {
				return nil
			}
			return fmt.Errorf("error retrieving domain %s during destroy check: %w", domainID, err)
		}

		if domain.RoutingConfigurationID != nil {
			return fmt.Errorf(
				"expected domain %s to have no routing_config_id after destroy, but found %s",
				domainID,
				*domain.RoutingConfigurationID,
			)
		}
	}

	return nil
}
