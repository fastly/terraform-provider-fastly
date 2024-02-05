package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyKVStore_validate(t *testing.T) {
	var (
		kvStore gofastly.KVStore
		service gofastly.ServiceDetail
	)
	kvStoreName := fmt.Sprintf("KV Store %s", acctest.RandString(10))
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
				Config: testAccKVStoreConfig(kvStoreName, serviceName, domainName, linkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.example", &service),
					testAccCheckFastlyKVStoreRemoteState(&service, &kvStore, serviceName, kvStoreName, linkName),
				),
			},
			{
				ResourceName:      "fastly_kvstore.example",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy", "package.0.filename", "imported"},
			},
			// IMPORTANT: Add a key to the store so we can validate force delete.
			{
				Config: testAccKVStoreConfig(kvStoreName, serviceName, domainName, linkName),
				Check:  testAccAddKVStoreItems(&kvStore), // triggers side-effect of adding a KV Store key/value
			},
			{
				Config: testAccKVStoreConfigDeleteStep1(kvStoreName, serviceName, domainName),
			},
			{
				Config: testAccKVStoreConfigDeleteStep2(serviceName, domainName),
			},
		},
	})
}

func testAccCheckFastlyKVStoreRemoteState(service *gofastly.ServiceDetail, kvStore *gofastly.KVStore, serviceName, kvStoreName, linkName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != serviceName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", serviceName, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		stores, err := conn.ListKVStores(&gofastly.ListKVStoresInput{})
		if err != nil {
			return fmt.Errorf("error listing all KV Stores: %s", err)
		}

		var found bool
		for _, store := range stores.Data {
			if store.Name == kvStoreName {
				found = true
				*kvStore = store
				break
			}
		}
		if !found {
			return fmt.Errorf("error looking up the KV Store")
		}

		links, err := conn.ListResources(&gofastly.ListResourcesInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.Version.Number),
		})
		if err != nil {
			return fmt.Errorf("error listing all resource links: %s", err)
		}

		found = false
		for _, link := range links {
			if gofastly.ToValue(link.Name) == linkName {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("error looking up the resource link")
		}

		return nil
	}
}

func testAccKVStoreConfig(kvStoreName, serviceName, domainName, linkName string) string {
	return fmt.Sprintf(`
resource "fastly_kvstore" "example" {
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
    resource_id = fastly_kvstore.example.id
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}
    `, kvStoreName, serviceName, domainName, linkName)
}

// IMPORTANT: Deleting a KV Store requires first deleting its resource_link.
// This requires a two-step `terraform apply` as we can't guarantee deletion order.
// e.g. resource_link deletion within fastly_service_compute might not finish first.
func testAccKVStoreConfigDeleteStep1(kvStoreName, serviceName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_kvstore" "example" {
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
    `, kvStoreName, serviceName, domainName)
}

// Step 1 deleted the resource_link first.
// Step 2 will now delete the fastly_kvstore.
func testAccKVStoreConfigDeleteStep2(serviceName, domainName string) string {
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

// testAccAddKVStoreItems doesn't technically check for anything despite
// returning a TestCheckFunc. Instead it is used for its side effect of adding
// a single KV Store entry so we can later validate we're able to force delete
// the KV Store resource even though it contains data.
func testAccAddKVStoreItems(kvStore *gofastly.KVStore) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		err := conn.InsertKVStoreKey(&gofastly.InsertKVStoreKeyInput{
			ID:    kvStore.ID,
			Key:   "test_key",
			Value: "test_value",
		})
		if err != nil {
			return fmt.Errorf("error adding item to KV Store '%s': %w", kvStore.ID, err)
		}
		return nil
	}
}
