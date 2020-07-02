package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccFastlyServiceWASMV1_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test1.tf-%s.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceWASMV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceWASMV1Config(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "version_comment", ""),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "active_version", "0"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "backend.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "package.#", "1"),
				),
			},
		},
	})
}

func testAccCheckServiceWASMV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_compute" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		l, err := conn.ListServices(&gofastly.ListServicesInput{})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing servcies when deleting Fastly Service (%s): %s", rs.Primary.ID, err)
		}

		for _, s := range l {
			if s.ID == rs.Primary.ID {
				// service still found
				return fmt.Errorf("[WARN] Tried deleting Service (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccServiceWASMV1Config(name, domain string) string {
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
