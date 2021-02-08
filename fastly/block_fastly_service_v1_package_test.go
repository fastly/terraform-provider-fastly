package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccFastlyServiceV1_package_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name01 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain01 := fmt.Sprintf("fastly-test.%s.com", name01)
	name02 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain02 := fmt.Sprintf("fastly-test.%s.com", name02)

	wp1 := gofastly.Package{
		Metadata: gofastly.PackageMetadata{
			Name:        "wasm-test",
			Description: "Test Package",
			Authors:     []string{"fastly@fastly.com"},
			Language:    "rust",
			Size:        2015936,
			HashSum:     "f99485bd301e23f028474d26d398da525de17a372ae9e7026891d7f85361d2540d14b3b091929c3f170eade573595e20b3405a9e29651ede59915f2e1652f616",
		},
	}

	wp2 := gofastly.Package{
		Metadata: gofastly.PackageMetadata{
			Name:        "edge-compute-test",
			Description: "Test Package",
			Authors:     []string{"fastly@fastly.com"},
			Language:    "rust",
			Size:        2158517,
			HashSum:     "ef62109f363007037d678120459008efb17b4cba5af2188d2eb0c6c6a69113b1925c44f5cbc7792b4421cad6f307bf3dd59adf0a73387291e0b854d3c25f2e48",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1PackageConfig(name01, domain01),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1PackageAttributes(&service, &wp1),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name01),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "package.#", "1"),
				),
			},
			{
				Config: testAccServiceV1PackageConfig(name02, domain02),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1PackageAttributes(&service, &wp1),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name02),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "package.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "active_version", "2"),
				),
			},
			{
				Config: testAccServiceV1PackageConfig_New(name02, domain02),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1PackageAttributes(&service, &wp2),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name02),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "package.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "active_version", "3"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1PackageAttributes(service *gofastly.ServiceDetail, computePackage *gofastly.Package) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wp, err := conn.GetPackage(&gofastly.GetPackageInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Package for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if computePackage.Metadata.Size != wp.Metadata.Size {
			return fmt.Errorf("Package size mismatch, expected: %v, got: %v", computePackage.Metadata.Size, wp.Metadata.Size)
		}

		if computePackage.Metadata.HashSum != wp.Metadata.HashSum {
			return fmt.Errorf("Package hashsum mismatch, expected: %v, got: %v", computePackage.Metadata.HashSum, wp.Metadata.HashSum)
		}

		if computePackage.Metadata.Language != wp.Metadata.Language {
			return fmt.Errorf("Package language mismatch, expected: %v, got: %v", computePackage.Metadata.Language, wp.Metadata.Language)
		}

		if computePackage.Metadata.Name != wp.Metadata.Name {
			return fmt.Errorf("Package name mismatch, expected: %v, got: %v", computePackage.Metadata.Name, wp.Metadata.Name)
		}

		return nil
	}
}

func testAccServiceV1PackageConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-package-test"
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
}
`, name, domain)
}

func testAccServiceV1PackageConfig_New(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-package-test"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  package {
    filename = "test_fixtures/package/valid2.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid2.tar.gz")
  }
  force_destroy = true
}
`, name, domain)
}
