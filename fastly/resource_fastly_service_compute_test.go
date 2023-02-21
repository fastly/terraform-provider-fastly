package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceCompute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test1.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceComputeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputeConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "version_comment", ""),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "active_version", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "1"),
				),
			},
			{
				ResourceName:      "fastly_service_compute.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy", "package.0.filename", "imported", "product_enablement"},
				// Validate product_enablement state after import.
				// If nothing has been configured, which is the default, we expect the
				// following state to be the result.
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected 1 state: %+v", s)
					}
					if s[0].Attributes["product_enablement.#"] != "1" {
						return fmt.Errorf("expected product_enablement to be imported into state")
					}
					if s[0].Attributes["product_enablement.0.%"] != "3" {
						return fmt.Errorf("expected product_enablement to contain 3 keys")
					}
					if s[0].Attributes["product_enablement.0.fanout"] != "false" {
						return fmt.Errorf("expected brotli_compression to be false")
					}
					if s[0].Attributes["product_enablement.0.websockets"] != "false" {
						return fmt.Errorf("expected websockets to be false")
					}
					if s[0].Attributes["product_enablement.0.name"] != "product_enablement" {
						return fmt.Errorf("expected the generated 'name' key to be 'product_enablement'")
					}
					return nil
				},
			},
		},
	})
}

func testAccCheckServiceComputeDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_compute" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		l, err := conn.ListServices(&gofastly.ListServicesInput{})
		if err != nil {
			return fmt.Errorf("error listing services when deleting Fastly Service (%s): %s", rs.Primary.ID, err)
		}

		for _, s := range l {
			if s.ID == rs.Primary.ID {
				// service still found
				return fmt.Errorf("tried deleting Service (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccServiceComputeConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }
  force_destroy = true
  activate = false
}`, name, domain)
}
