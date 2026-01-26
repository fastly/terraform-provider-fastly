package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/domainmanagement/v1/domains"
)

func TestAccFastlyDomainServiceLink_Basic(t *testing.T) {
	suffix := acctest.RandString(6)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainServiceLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainServiceLinkDynamicConfig("svc1", suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("fastly_domain_service_link.example", "domain_id"),
					resource.TestCheckResourceAttrSet("fastly_domain_service_link.example", "service_id"),
				),
			},
			{
				Config: testAccDomainServiceLinkDynamicConfig("svc2", suffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("fastly_domain_service_link.example", "service_id"),
				),
			},
			{
				ResourceName:      "fastly_domain_service_link.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDomainServiceLinkDynamicConfig(serviceRef string, suffix string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-%s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

resource "fastly_service_vcl" "svc2" {
  name          = "test-svc-2-%s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-2"
  }
}

resource "fastly_domain" "domain" {
  fqdn = "test-%s.example.com"

  lifecycle {
    ignore_changes = [service_id]
  }
}

resource "fastly_domain_service_link" "example" {
  domain_id  = fastly_domain.domain.domain_id
  service_id = fastly_service_vcl.%s.id
}
`, suffix, suffix, suffix, serviceRef)
}

func testAccCheckDomainServiceLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_domain_service_link" {
			continue
		}

		domainID := rs.Primary.ID
		if domainID == "" {
			continue
		}

		input := &domains.GetInput{
			DomainID: gofastly.ToPointer(domainID),
		}

		domain, err := domains.Get(context.Background(), conn, input)
		if err != nil {
			if httpErr, ok := err.(*gofastly.HTTPError); ok && httpErr.StatusCode == 404 {
				return nil
			}
			return fmt.Errorf("error retrieving domain %s during destroy check: %w", domainID, err)
		}

		if domain.ServiceID != nil {
			return fmt.Errorf(
				"expected domain %s to have no service_id after destroy, but found %s",
				domainID,
				*domain.ServiceID,
			)
		}
	}

	return nil
}
