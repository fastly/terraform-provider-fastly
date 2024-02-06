package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenConfigStoreEntries(t *testing.T) {
	cases := []struct {
		remote []*gofastly.ConfigStoreItem
		local  map[string]string
	}{
		{
			remote: []*gofastly.ConfigStoreItem{
				{
					StoreID: "1234567890",
					Key:     "key-1",
					Value:   "value-1",
				},
				{
					StoreID: "1234567890",
					Key:     "key-2",
					Value:   "value-2",
				},
			},
			local: map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
			},
		},
	}

	for _, c := range cases {
		out := flattenConfigStoreEntries(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyConfigStoreEntries_validate(t *testing.T) {
	storeName := fmt.Sprintf("store_%s", acctest.RandString(10))

	want1 := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	want2 := map[string]string{
		"key1": "value1_updated",
		"key2": "value2_updated",
		"key3": "value3",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfigStoreEntries(storeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyServiceConfigStoreEntriesRemoteState(storeName, want1),
					resource.TestCheckResourceAttr("fastly_configstore_entries.example", "entries.%", "2"),
				),
			},
			{
				Config: testAccServiceConfigStoreEntriesUpdate(storeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyServiceConfigStoreEntriesRemoteState(storeName, want2),
					resource.TestCheckResourceAttr("fastly_configstore_entries.example", "entries.%", "3"),
				),
			},
			{
				ResourceName:            "fastly_configstore_entries.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manage_entries"},
			},
		},
	})
}

func testAccServiceConfigStoreEntries(storeName string) string {
	return fmt.Sprintf(`
resource "fastly_configstore" "example" {
  name          = "%s"
  force_destroy = true
}

resource "fastly_configstore_entries" "example" {
  store_id = fastly_configstore.example.id
  entries = {
    key1: "value1"
    key2: "value2"
  }
  manage_entries = true
}
`, storeName)
}

func testAccServiceConfigStoreEntriesUpdate(storeName string) string {
	return fmt.Sprintf(`
resource "fastly_configstore" "example" {
  name          = "%s"
  force_destroy = true
}

resource "fastly_configstore_entries" "example" {
  store_id = fastly_configstore.example.id
  entries = {
    key1: "value1_updated"
    key2: "value2_updated"
    key3: "value3"
  }
  manage_entries = true
}
`, storeName)
}

func testAccCheckFastlyServiceConfigStoreEntriesRemoteState(storeName string, want map[string]string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn

		stores, err := conn.ListConfigStores(&gofastly.ListConfigStoresInput{})
		if err != nil {
			return fmt.Errorf("failed to get list of Config Stores")
		}

		var found *gofastly.ConfigStore

		for _, store := range stores {
			if store.Name == storeName {
				found = store
				break
			}
		}

		if found == nil {
			return fmt.Errorf("failed to find Config Store")
		}

		entries, err := conn.ListConfigStoreItems(&gofastly.ListConfigStoreItemsInput{
			StoreID: found.StoreID,
		})
		if err != nil {
			return fmt.Errorf("failed to get Config Store entries")
		}

		got := flattenConfigStoreEntries(entries)

		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("error matching:\nexpected: %#v\ngot: %#v", want, got)
		}

		return nil
	}
}
