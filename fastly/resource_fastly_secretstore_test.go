package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlySecretStore_validate(t *testing.T) {
	var service gofastly.ServiceDetail
	secretStoreName := fmt.Sprintf("tf-test-secret-store-%s", acctest.RandString(10))
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	linkName := fmt.Sprintf("resource_link_%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretStoreConfig(secretStoreName, serviceName, domainName, linkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.example", &service),
					testAccCheckFastlySecretStoreRemoteState(&service, serviceName, secretStoreName, linkName),
				),
			},
			{
				ResourceName:      "fastly_secretstore.example",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "package.0.filename", "imported"},
			},
			{
				Config: testAccSecretStoreConfigDeleteStep1(secretStoreName, serviceName, domainName),
			},
			{
				Config: testAccSecretStoreConfigDeleteStep2(serviceName, domainName),
			},
		},
	})
}

func testAccCheckFastlySecretStoreRemoteState(service *gofastly.ServiceDetail, serviceName, storeName, linkName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != serviceName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", serviceName, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		ss := []gofastly.SecretStore{}

		var cursor string

		for {
			stores, err := conn.ListSecretStores(&gofastly.ListSecretStoresInput{
				Cursor: cursor,
			})
			if err != nil {
				return fmt.Errorf("error listing all Secret Stores: %s", err)
			}

			ss = append(ss, stores.Data...)

			cursor = stores.Meta.NextCursor
			if cursor == "" {
				break
			}
		}

		var found bool
		for _, store := range ss {
			if store.Name == storeName {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("error looking up the Secret Store")
		}

		links, err := conn.ListResources(&gofastly.ListResourcesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.Version.Number),
		})
		if err != nil {
			return fmt.Errorf("error listing all resource links: %s", err)
		}

		found = false
		for _, link := range links {
			if gofastly.ToValue(link.Name) == linkName {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("error looking up the resource link")
		}

		return nil
	}
}

func testAccSecretStoreConfig(configStoreName, serviceName, domainName, linkName string) string {
	return fmt.Sprintf(`
resource "fastly_secretstore" "example" {
  name          = "%s"
}
resource "fastly_service_compute" "example" {
  name = "%s"
  domain {
    name = "%s"
  }
  package {
    filename         = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }
  resource_link {
    name        = "%s"
    resource_id = fastly_secretstore.example.id
  }
  force_destroy = true
}
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}
    `, configStoreName, serviceName, domainName, linkName)
}

// IMPORTANT: Deleting a Secret Store requires first deleting its resource_link.
// This requires a two-step `terraform apply` as we can't guarantee deletion order.
// e.g. resource_link deletion within fastly_service_compute might not finish first.
func testAccSecretStoreConfigDeleteStep1(configStoreName, serviceName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_secretstore" "example" {
  name          = "%s"
}
resource "fastly_service_compute" "example" {
  name = "%s"
  domain {
    name = "%s"
  }
  package {
    filename         = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }
  force_destroy = true
}
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}
    `, configStoreName, serviceName, domainName)
}

// Step 1 deleted the resource_link first.
// Step 2 will now delete the fastly_secretstore.
func testAccSecretStoreConfigDeleteStep2(serviceName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "example" {
  name = "%s"
  domain {
    name = "%s"
  }
  package {
    filename         = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }
  force_destroy = true
}
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}
    `, serviceName, domainName)
}
