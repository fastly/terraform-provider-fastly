package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/domainmanagement/v1/domains"
)

const (
	domainID   = "VrU9QvSHFBqvO3WJZoro0w"
	serviceID  = "4Y3PRaauBiVcQXPOJUURj6"
	serviceID2 = "iWBciPXoEl8PfUOW2hvIm4"
)

// TestAccFastlyDomainV1ServiceLink_Basic tests resource creation from scratch (without import).
// This ensures the Create → Read flow works correctly, as import can mask certain behaviors where
// setting d.Id() before Read is called [CDTOOL-1198].
func TestAccFastlyDomainV1ServiceLink_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainV1ServiceLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "fastly_domain_service_link" "example" {
				    domain_id = "%s"
					service_id = "%s"
				}
				`, domainID, serviceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_domain_service_link.example", "domain_id", domainID),
					resource.TestCheckResourceAttr("fastly_domain_service_link.example", "service_id", serviceID),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "fastly_domain_service_link" "example" {
				    domain_id = "%s"
					service_id = "%s"
				}
				`, domainID, serviceID2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_domain_service_link.example", "domain_id", domainID),
					resource.TestCheckResourceAttr("fastly_domain_service_link.example", "service_id", serviceID2),
				),
			},
		},
	})
}

func testAccCheckDomainV1ServiceLinkDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_domain_service_link" {
			continue
		}
		conn := testAccProvider.Meta().(*APIClient).conn
		input := &domains.ListInput{
			ServiceID: gofastly.ToPointer(serviceID2),
		}
		cl, err := domains.List(context.TODO(), conn, input)
		if err != nil {
			return fmt.Errorf("failed to list domains for fastly_domain resource: %w", err)
		}
		if cl != nil && len(cl.Data) > 0 {
			for _, d := range cl.Data {
				if d.DomainID == domainID && *d.ServiceID == serviceID2 {
					return fmt.Errorf("tried deleting domain service link (%s), but was still found", serviceID2)
				}
			}
		}
		return nil
	}
	return nil
}
