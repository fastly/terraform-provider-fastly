package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyConfigStore_validate(t *testing.T) {
	var (
		configStore gofastly.ConfigStore
		service     gofastly.ServiceDetail
	)
	configStoreName := fmt.Sprintf("Config Store %s", acctest.RandString(10))
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
				Config: testAccConfigStoreConfig(configStoreName, serviceName, domainName, linkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.example", &service),
					testAccCheckFastlyConfigStoreRemoteState(&service, &configStore, serviceName, configStoreName, linkName),
				),
			},
			{
				ResourceName:      "fastly_configstore.example",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy", "package.0.filename", "imported"},
			},
			// IMPORTANT: Add a key to the store so we can validate force delete.
			{
				Config: testAccConfigStoreConfig(configStoreName, serviceName, domainName, linkName),
				Check:  testAccAddConfigStoreItems(&configStore), // triggers side-effect of adding a Config Store key/value
			},
			{
				Config: testAccConfigStoreConfigDeleteStep1(configStoreName, serviceName, domainName),
			},
			{
				Config: testAccConfigStoreConfigDeleteStep2(serviceName, domainName),
			},
		},
	})
}

func testAccCheckFastlyConfigStoreRemoteState(service *gofastly.ServiceDetail, configStore *gofastly.ConfigStore, serviceName, configStoreName, linkName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if service.Name != serviceName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", serviceName, service.Name)
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		stores, err := conn.ListConfigStores()
		if err != nil {
			return fmt.Errorf("error listing all Config Stores: %s", err)
		}

		var found bool
		for _, store := range stores {
			if store.Name == configStoreName {
				found = true
				*configStore = *store
				break
			}
		}
		if !found {
			return fmt.Errorf("error looking up the Config Store")
		}

		links, err := conn.ListResources(&gofastly.ListResourcesInput{
			ServiceID:      service.ID,
			ServiceVersion: service.Version.Number,
		})
		if err != nil {
			return fmt.Errorf("error listing all resource links: %s", err)
		}

		found = false
		for _, link := range links {
			if link.Name == linkName {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("error looking up the resource link")
		}

		return nil
	}
}

func testAccConfigStoreConfig(configStoreName, serviceName, domainName, linkName string) string {
	return fmt.Sprintf(`
resource "fastly_configstore" "example" {
  name          = "%s"
  force_destroy = true
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
    resource_id = fastly_configstore.example.id
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}
    `, configStoreName, serviceName, domainName, linkName)
}

// IMPORTANT: Deleting a Config Store requires first deleting its resource_link.
// This requires a two-step `terraform apply` as we can't guarantee deletion order.
// e.g. resource_link deletion within fastly_service_compute might not finish first.
func testAccConfigStoreConfigDeleteStep1(configStoreName, serviceName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_configstore" "example" {
  name          = "%s"
  force_destroy = true
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
// Step 2 will now delete the fastly_kvstore.
func testAccConfigStoreConfigDeleteStep2(serviceName, domainName string) string {
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

// testAccAddConfigStoreItems doesn't technically check for anything despite
// returning a TestCheckFunc. Instead it is used for its side effect of adding
// a single Config Store entry so we can later validate we're able to force
// delete the Config Store resource even though it contains data.
func testAccAddConfigStoreItems(configStore *gofastly.ConfigStore) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		_, err := conn.CreateConfigStoreItem(&gofastly.CreateConfigStoreItemInput{
			StoreID: configStore.ID,
			Key:     "test_key",
			Value:   "test_value",
		})
		if err != nil {
			return fmt.Errorf("error adding item to Config Store '%s': %w", configStore.ID, err)
		}
		return nil
	}
}
