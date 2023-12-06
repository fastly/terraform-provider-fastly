package fastly

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceCompute_package_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name01 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain01 := fmt.Sprintf("fastly-test.%s.com", name01)
	name02 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain02 := fmt.Sprintf("fastly-test.%s.com", name02)

	want := gofastly.Package{
		Metadata: &gofastly.PackageMetadata{
			Name:        gofastly.ToPointer("wasm-test"),
			Description: gofastly.ToPointer("Test Package"),
			Authors:     []string{"fastly@fastly.com"},
			Language:    gofastly.ToPointer("rust"),
			Size:        gofastly.ToPointer(int64(2015936)),
			FilesHash:   gofastly.ToPointer("a763d3c88968ebc17691900d3c14306762296df8e47a1c2d7661cee0e0c5aa6d4c082a7c128d6e719fe333b73b46fe3ae32694716ccd2efa21f5d9f049ceec6d"),
		},
	}

	want2 := gofastly.Package{
		Metadata: &gofastly.PackageMetadata{
			Name:        gofastly.ToPointer("edge-compute-test"),
			Description: gofastly.ToPointer("Test Package"),
			Authors:     []string{"fastly@fastly.com"},
			Language:    gofastly.ToPointer("rust"),
			Size:        gofastly.ToPointer(int64(2158517)),
			FilesHash:   gofastly.ToPointer("d8f8a0448ae4d3a6f5f230caf1269c2986e7cba86ebd14add5118034607992eafebf16714e33c1733ecbd61f13f4aef0d4dbe7582313baec343d00e8fdc424f7"),
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputePackageConfig(name01, domain01),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputePackageAttributes(&service, &want),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name01),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "1"),
				),
			},
			{
				Config: testAccServiceComputePackageConfig(name02, domain02),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputePackageAttributes(&service, &want),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name02),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "active_version", "2"),
				),
			},
			{
				Config: testAccServiceComputePackageConfigNew(name02, domain02),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputePackageAttributes(&service, &want2),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name02),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "active_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCompute_package_content(t *testing.T) {
	validPackageContent, _ := os.ReadFile("test_fixtures/package/valid.tar.gz")
	b64Content := base64.StdEncoding.EncodeToString(validPackageContent)

	validPackageContent2, _ := os.ReadFile("test_fixtures/package/valid2.tar.gz")
	b64Content2 := base64.StdEncoding.EncodeToString(validPackageContent2)

	var service gofastly.ServiceDetail
	name01 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain01 := fmt.Sprintf("fastly-test.%s.com", name01)
	name02 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain02 := fmt.Sprintf("fastly-test.%s.com", name02)

	want := gofastly.Package{
		Metadata: &gofastly.PackageMetadata{
			Name:        gofastly.ToPointer("wasm-test"),
			Description: gofastly.ToPointer("Test Package"),
			Authors:     []string{"fastly@fastly.com"},
			Language:    gofastly.ToPointer("rust"),
			Size:        gofastly.ToPointer(int64(2015936)),
			FilesHash:   gofastly.ToPointer("a763d3c88968ebc17691900d3c14306762296df8e47a1c2d7661cee0e0c5aa6d4c082a7c128d6e719fe333b73b46fe3ae32694716ccd2efa21f5d9f049ceec6d"),
		},
	}

	want2 := gofastly.Package{
		Metadata: &gofastly.PackageMetadata{
			Name:        gofastly.ToPointer("edge-compute-test"),
			Description: gofastly.ToPointer("Test Package"),
			Authors:     []string{"fastly@fastly.com"},
			Language:    gofastly.ToPointer("rust"),
			Size:        gofastly.ToPointer(int64(2158517)),
			FilesHash:   gofastly.ToPointer("d8f8a0448ae4d3a6f5f230caf1269c2986e7cba86ebd14add5118034607992eafebf16714e33c1733ecbd61f13f4aef0d4dbe7582313baec343d00e8fdc424f7"),
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputePackageConfigContent(name01, domain01, b64Content),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputePackageAttributes(&service, &want),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name01),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "1"),
				),
			},
			{
				Config: testAccServiceComputePackageConfigContent(name02, domain02, b64Content),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputePackageAttributes(&service, &want),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name02),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "active_version", "2"),
				),
			},
			{
				Config: testAccServiceComputePackageConfigContent(name02, domain02, b64Content2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputePackageAttributes(&service, &want2),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name02),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "active_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCompute_package_optional(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputePackageOptional(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "0"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceComputePackageAttributes(service *gofastly.ServiceDetail, want *gofastly.Package) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		got, err := conn.GetPackage(&gofastly.GetPackageInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Package for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if gofastly.ToValue(want.Metadata.Size) != gofastly.ToValue(got.Metadata.Size) {
			return fmt.Errorf("package size mismatch, expected: %v, got: %v", gofastly.ToValue(want.Metadata.Size), gofastly.ToValue(got.Metadata.Size))
		}

		if gofastly.ToValue(want.Metadata.FilesHash) != gofastly.ToValue(got.Metadata.FilesHash) {
			return fmt.Errorf("package files_hash mismatch, expected: %v, got: %v", gofastly.ToValue(want.Metadata.FilesHash), gofastly.ToValue(got.Metadata.FilesHash))
		}

		if gofastly.ToValue(want.Metadata.Language) != gofastly.ToValue(got.Metadata.Language) {
			return fmt.Errorf("package language mismatch, expected: %v, got: %v", gofastly.ToValue(want.Metadata.Language), gofastly.ToValue(got.Metadata.Language))
		}

		if gofastly.ToValue(want.Metadata.Name) != gofastly.ToValue(got.Metadata.Name) {
			return fmt.Errorf("package name mismatch, expected: %v, got: %v", gofastly.ToValue(want.Metadata.Name), gofastly.ToValue(got.Metadata.Name))
		}

		return nil
	}
}

func testAccServiceComputePackageConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

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
    source_code_hash = data.fastly_package_hash.example.hash
  }
  force_destroy = true
}
`, name, domain)
}

func testAccServiceComputePackageConfigNew(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid2.tar.gz"
}

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
    source_code_hash = data.fastly_package_hash.example.hash
  }
  force_destroy = true
}
`, name, domain)
}

// NOTE: Test config was unable to use input variable implementation.
// This is because we can't set `-var` via test suite.
// Instead we need to use exported environment variable: TF_VAR_package_content.
// Problem with that is the base64 encoded string is larger than the OS limit.
// So instead we had to use `locals` variable as a workaround.
// https://developer.hashicorp.com/terraform/language/values/locals
// The use of `-var` and input variables will work fine with actual TF project.
func testAccServiceComputePackageConfigContent(name, domain, b64Content string) string {
	return fmt.Sprintf(`
locals {
  package_content = "%s"
}

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
    content = local.package_content
  }
  force_destroy = true
}
`, b64Content, name, domain)
}

// NOTE: Config must set `activate = false` to avoid validation errors.
// You can't activate a compute service without a package.
func testAccServiceComputePackageOptional(name string, domain string) string {
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
  force_destroy = true
  activate = false
}
`, name, domain)
}
