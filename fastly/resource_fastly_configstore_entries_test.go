package fastly

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
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

func TestResourceFastlyConfigStoreEntriesReadClearsEntriesWhenManageEntriesFalse(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceFastlyConfigStoreEntries().Schema, map[string]any{
		"store_id":       "store-id",
		"manage_entries": false,
	})
	if err := d.Set("entries", map[string]string{
		"key1": "value1",
		"key2": "value2",
	}); err != nil {
		t.Fatalf("failed to seed Config Store entries in test state: %v", err)
	}
	d.SetId("store-id/entries")

	diags := resourceFastlyConfigStoreEntriesRead(context.Background(), d, nil)
	if diags.HasError() {
		t.Fatalf("expected Config Store entries read to clear unmanaged entries without refreshing, got diagnostics: %v", diags)
	}

	entries := d.Get("entries").(map[string]any)
	if len(entries) != 0 {
		t.Fatalf("expected unmanaged Config Store entries to be cleared from state, got %d entries", len(entries))
	}
}

func TestResourceFastlyConfigStoreEntriesImportEnablesManageEntries(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceFastlyConfigStoreEntries().Schema, map[string]any{})
	d.SetId("store-id/entries")

	result, err := resourceConfigStoreEntriesImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected import error: %s", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected one imported resource, got %d", len(result))
	}

	if got := result[0].Get("manage_entries").(bool); !got {
		t.Fatal("expected manage_entries to be true after import")
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

func TestAccFastlyConfigStoreEntries_manage_entries_false(t *testing.T) {
	storeName := fmt.Sprintf("store_%s", acctest.RandString(10))

	initialEntries := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	updatedEntries := map[string]string{
		"key1": "value1_updated",
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
				Config: testAccServiceConfigStoreEntriesManageEntriesFalse(storeName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyServiceConfigStoreEntriesRemoteState(storeName, initialEntries),
				),
			},
			{
				Config: testAccServiceConfigStoreEntriesManageEntriesFalse(storeName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyServiceConfigStoreEntriesRemoteState(storeName, initialEntries),
					resource.TestCheckResourceAttr("fastly_configstore_entries.example", "entries.%", "0"),
				),
			},
			{
				Config: testAccServiceConfigStoreEntriesManageEntriesTrue(storeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyServiceConfigStoreEntriesRemoteState(storeName, updatedEntries),
					resource.TestCheckResourceAttr("fastly_configstore_entries.example", "entries.%", "2"),
				),
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

func testAccServiceConfigStoreEntriesManageEntriesFalse(storeName string, updated bool) string {
	entries := `
    key1 = "value1"
    key2 = "value2"
`
	if updated {
		entries = `
    key1 = "value1_updated"
    key3 = "value3"
`
	}

	return fmt.Sprintf(`
resource "fastly_configstore" "example" {
  name          = "%s"
  force_destroy = true
}

resource "fastly_configstore_entries" "example" {
  store_id = fastly_configstore.example.id
  entries = {%s  }
  manage_entries = false
}
`, storeName, entries)
}

func testAccServiceConfigStoreEntriesManageEntriesTrue(storeName string) string {
	return fmt.Sprintf(`
resource "fastly_configstore" "example" {
  name          = "%s"
  force_destroy = true
}

resource "fastly_configstore_entries" "example" {
  store_id = fastly_configstore.example.id
  entries = {
    key1 = "value1_updated"
    key3 = "value3"
  }
  manage_entries = true
}
`, storeName)
}

func testAccCheckFastlyServiceConfigStoreEntriesRemoteState(storeName string, want map[string]string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn

		stores, err := conn.ListConfigStores(context.TODO(), &gofastly.ListConfigStoresInput{})
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

		entries, err := conn.ListConfigStoreItems(context.TODO(), &gofastly.ListConfigStoreItemsInput{
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
