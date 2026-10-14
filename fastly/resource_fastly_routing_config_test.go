package fastly

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths"
)

func TestAccFastlyRoutingConfig_Basic(t *testing.T) {
	suffix := acctest.RandString(6)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoutingConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingConfigConfig(suffix, "svc1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("fastly_routing_config.example", "routing_config_id"),
					resource.TestCheckResourceAttr("fastly_routing_config.example", "path.#", "1"),
					resource.TestCheckResourceAttr("fastly_routing_config.example", "path.0.path", "/api/*"),
					resource.TestCheckResourceAttr("fastly_routing_config.example", "path.0.rule.#", "1"),
					resource.TestCheckResourceAttrSet("fastly_routing_config.example", "path.0.rule.0.rule_id"),
					resource.TestCheckResourceAttrPair("fastly_routing_config.example", "path.0.rule.0.action_value", "fastly_service_vcl.svc1", "id"),
				),
			},
			{
				Config: testAccRoutingConfigConfig(suffix, "svc2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("fastly_routing_config.example", "path.0.rule.0.action_value", "fastly_service_vcl.svc2", "id"),
				),
			},
			{
				ResourceName:      "fastly_routing_config.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyRoutingConfig_DuplicatePath(t *testing.T) {
	suffix := acctest.RandString(6)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoutingConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRoutingConfigDuplicatePathConfig(suffix),
				ExpectError: regexp.MustCompile(`duplicate path`),
			},
		},
	})
}

// TestAccFastlyRoutingConfig_DetectsOutOfBandPath verifies that Read
// discovers a path added directly through the API (outside of Terraform),
// since it now lists the routing config's current paths/rules rather than
// only re-fetching ones already known to state.
func TestAccFastlyRoutingConfig_DetectsOutOfBandPath(t *testing.T) {
	suffix := acctest.RandString(6)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoutingConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingConfigConfig(suffix, "svc1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_routing_config.example", "path.#", "1"),
					testAccRoutingConfigCreateOutOfBandPath("fastly_routing_config.example"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccRoutingConfigCreateOutOfBandPath creates an extra path directly
// through the API, simulating a change made outside of Terraform.
func testAccRoutingConfigCreateOutOfBandPath(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		routingConfigID := rs.Primary.ID
		if routingConfigID == "" {
			return fmt.Errorf("no ID set for %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		path := "/out-of-band/*"
		if _, err := paths.Create(context.Background(), conn, &paths.CreateInput{
			RoutingConfigID: &routingConfigID,
			Path:            &path,
		}); err != nil {
			return fmt.Errorf("error creating out-of-band path: %w", err)
		}
		return nil
	}
}

func testAccRoutingConfigDuplicatePathConfig(suffix string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "svc1" {
  name          = "tf-test-routing-config-dup-%[1]s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

resource "fastly_routing_config" "example" {
  name = "tf-test-routing-config-dup-%[1]s"

  path {
    path = "/api/*"

    rule {
      action_type  = "service"
      action_value = fastly_service_vcl.svc1.id
    }
  }

  path {
    path = "/api/*"

    rule {
      action_type  = "service"
      action_value = fastly_service_vcl.svc1.id
    }
  }
}
`, suffix)
}

func testAccRoutingConfigConfig(suffix, serviceRef string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "svc1" {
  name          = "tf-test-routing-config-svc1-%[1]s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

resource "fastly_service_vcl" "svc2" {
  name          = "tf-test-routing-config-svc2-%[1]s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-2"
  }
}

resource "fastly_routing_config" "example" {
  name = "tf-test-routing-config-%[1]s"

  path {
    path = "/api/*"

    rule {
      action_type  = "service"
      action_value = fastly_service_vcl.%[2]s.id
    }
  }
}
`, suffix, serviceRef)
}

func testAccCheckRoutingConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_routing_config" {
			continue
		}

		id := rs.Primary.ID
		if id == "" {
			continue
		}

		_, err := routingconfigs.Get(context.Background(), conn, &routingconfigs.GetInput{RoutingConfigID: &id})
		if err == nil {
			return fmt.Errorf("routing config %s still exists", id)
		}
		if httpErr, ok := err.(*gofastly.HTTPError); !ok || httpErr.StatusCode != 404 {
			return fmt.Errorf("error checking routing config %s destroy: %w", id, err)
		}
	}

	return nil
}
