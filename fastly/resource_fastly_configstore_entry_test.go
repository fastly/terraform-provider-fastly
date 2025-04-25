package fastly

import (
	"fmt"
	"strings"
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyConfigStoreEntry_basic(t *testing.T) {
	var configStore gofastly.ConfigStore
	storeName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key := fmt.Sprintf("tf-test-key-%s", acctest.RandString(5))
	value := acctest.RandString(15)
	updatedValue := acctest.RandString(15)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckConfigStoreEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigStoreEntryConfig(storeName, key, value),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyConfigStoreExists("fastly_configstore.foo", &configStore),
					testAccCheckConfigStoreEntryExists("fastly_configstore_entry.test", key, value),
					resource.TestCheckResourceAttr("fastly_configstore_entry.test", "key", key),
					resource.TestCheckResourceAttr("fastly_configstore_entry.test", "value", value),
				),
			},
			{
				Config: testAccConfigStoreEntryConfigUpdated(storeName, key, updatedValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyConfigStoreExists("fastly_configstore.foo", &configStore),
					testAccCheckConfigStoreEntryExists("fastly_configstore_entry.test", key, updatedValue),
					resource.TestCheckResourceAttr("fastly_configstore_entry.test", "key", key),
					resource.TestCheckResourceAttr("fastly_configstore_entry.test", "value", updatedValue),
				),
			},
			{
				ResourceName:      "fastly_configstore_entry.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckFastlyConfigStoreExists(n string, configStore *gofastly.ConfigStore) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		store, err := conn.GetConfigStore(&gofastly.GetConfigStoreInput{
			StoreID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("error fetching ConfigStore: %s", err)
		}

		*configStore = *store
		return nil
	}
}

func testAccCheckConfigStoreEntryExists(resourceName, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		idParts := strings.Split(rs.Primary.ID, "/")
		if len(idParts) != 2 {
			return fmt.Errorf("invalid ID format: %s", rs.Primary.ID)
		}
		storeID := idParts[0]
		entryKey := idParts[1]

		if key != entryKey {
			return fmt.Errorf("resource key mismatch, expected: %s, got: %s", key, entryKey)
		}

		item, err := conn.GetConfigStoreItem(&gofastly.GetConfigStoreItemInput{
			StoreID: storeID,
			Key:     key,
		})
		if err != nil {
			return fmt.Errorf("error looking up configstore entry (%s/%s): %s", storeID, key, err)
		}

		if item.Value != value {
			return fmt.Errorf("unexpected value for configstore entry, expected: %s, got: %s", value, item.Value)
		}

		return nil
	}
}

func testAccCheckConfigStoreEntryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_configstore_entry" {
			continue
		}

		idParts := strings.Split(rs.Primary.ID, "/")
		if len(idParts) != 2 {
			return fmt.Errorf("invalid ID format: %s", rs.Primary.ID)
		}
		storeID := idParts[0]
		key := idParts[1]

		_, err := conn.GetConfigStoreItem(&gofastly.GetConfigStoreItemInput{
			StoreID: storeID,
			Key:     key,
		})
		if err == nil {
			return fmt.Errorf("configstore entry still exists: %s/%s", storeID, key)
		}

		if e, ok := err.(*gofastly.HTTPError); !ok || !e.IsNotFound() {
			return fmt.Errorf("error checking configstore entry (%s/%s) destruction: %s", storeID, key, err)
		}
	}

	return nil
}

func testAccConfigStoreEntryConfig(storeName, key, value string) string {
	return fmt.Sprintf(`
resource "fastly_configstore" "foo" {
  name = "%s"
}

resource "fastly_configstore_entry" "test" {
  store_id = fastly_configstore.foo.id
  key      = "%s"
  value    = "%s"
}
`, storeName, key, value)
}

func testAccConfigStoreEntryConfigUpdated(storeName, key, value string) string {
	return fmt.Sprintf(`
resource "fastly_configstore" "foo" {
  name = "%s"
}

resource "fastly_configstore_entry" "test" {
  store_id = fastly_configstore.foo.id
  key      = "%s"
  value    = "%s"
}
`, storeName, key, value)
}