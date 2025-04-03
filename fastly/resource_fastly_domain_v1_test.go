package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	v1 "github.com/fastly/go-fastly/v10/fastly/domains/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDomainV1_Basic(t *testing.T) {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "fastly_domain_v1" "example" {
				    fqdn = "%s"
				}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_domain_v1.example", "fqdn", domainName),
					resource.TestCheckNoResourceAttr("fastly_domain_v1.example", "service_id"),
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "fastly_service_vcl" "example" {
					name = "%s"
					domain {
					    name = "%s"
					}
					force_destroy = true
				}
				resource "fastly_domain_v1" "example" {
				    fqdn = "%s"
					service_id = resource.fastly_service_vcl.example.id
				}
				`, domainName, domainName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_domain_v1.example", "fqdn", domainName),
					resource.TestCheckResourceAttrSet("fastly_domain_v1.example", "service_id"),
				),
			},
			{
				ResourceName:      "fastly_domain_v1.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDomainV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_domain_v1" {
			continue
		}
		a := rs.Primary.Attributes
		fqdn := a["fqdn"]
		conn := testAccProvider.Meta().(*APIClient).conn
		input := &v1.ListInput{
			FQDN: gofastly.ToPointer(fqdn),
		}
		cl, err := v1.List(conn, input)
		if err != nil {
			return fmt.Errorf("failed to list domains for fastly_domain_v1 resource: %w", err)
		}
		if cl != nil && len(cl.Data) > 0 {
			for _, d := range cl.Data {
				if d.FQDN == fqdn {
					return fmt.Errorf("tried deleting domain (%s), but was still found", fqdn)
				}
			}
		}
		return nil
	}
	return nil
}
