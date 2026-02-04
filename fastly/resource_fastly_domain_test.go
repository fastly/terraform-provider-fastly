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

func TestAccFastlyDomain_Basic(t *testing.T) {
	description := "example"
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "fastly_domain" "example" {
				    description = "%s"
				    fqdn = "%s"
				}
				`, description, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_domain.example", "fqdn", domainName),
					resource.TestCheckResourceAttr("fastly_domain.example", "description", description),
					resource.TestCheckNoResourceAttr("fastly_domain.example", "service_id"),
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
				resource "fastly_domain" "example" {
				    description = "%s"
				    fqdn = "%s"
					service_id = resource.fastly_service_vcl.example.id
				}
				`, domainName, domainName, description+"-updated", domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_domain.example", "fqdn", domainName),
					resource.TestCheckResourceAttr("fastly_domain.example", "description", description+"-updated"),
					resource.TestCheckResourceAttrSet("fastly_domain.example", "service_id"),
				),
			},
			{
				ResourceName:      "fastly_domain.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_domain" {
			continue
		}
		a := rs.Primary.Attributes
		fqdn := a["fqdn"]
		conn := testAccProvider.Meta().(*APIClient).conn
		input := &domains.ListInput{
			FQDN: gofastly.ToPointer(fqdn),
		}
		cl, err := domains.List(context.TODO(), conn, input)
		if err != nil {
			return fmt.Errorf("failed to list domains for fastly_domain resource: %w", err)
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

func TestResourceFastlyDomainV1_DeprecationRegistered(t *testing.T) {
	p := Provider()

	resource, ok := p.ResourcesMap["fastly_domain_v1"]
	if !ok {
		t.Fatal("expected resource fastly_domain_v1 to be registered")
	}

	if resource.DeprecationMessage == "" {
		t.Fatal("expected fastly_domain_v1 to have a DeprecationMessage")
	}
}
