package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenKVStore(t *testing.T) {
	cases := []struct {
		remote []gofastly.KVStore
		local  []map[string]any
	}{
		{
			remote: []gofastly.KVStore{
				{
					ID:   "1234567890",
					Name: "dictionary-example",
				},
			},
			local: []map[string]any{
				{
					"store_id": "1234567890",
					"name":     "dictionary-example",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenKVStores(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceCompute_kvstore(t *testing.T) {
	var service gofastly.ServiceDetail
	var kvstore gofastly.KVStore
	name := acctest.RandomWithPrefix(testResourcePrefix)
	storeName := fmt.Sprintf("kvstore %s", acctest.RandString(10))
	updatedStoreName := fmt.Sprintf("new kvstore %s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	// Six part test:
	// 1. Create service with kvstore
	// 2. Rename the kvstore, should succeed because it is empty
	// 3. Keep kvstore the same and add an entry to it
	// 4. Try to rename it, expect to fail with "kvstore not empty error"
	// 5. Without renaming, set force_destroy=true to skip the deletion check
	// 6. Try to rename again, expect to succeed
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceComputeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputeConfigKVStore(name, storeName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputeAttributesKVStore(&service, &kvstore, name, storeName),
				),
			},
			{
				Config: testAccServiceComputeConfigKVStore(name, updatedStoreName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputeAttributesKVStore(&service, &kvstore, name, updatedStoreName),
				),
			},
			{
				Config: testAccServiceComputeConfigKVStore(name, updatedStoreName, domainName),
				Check:  testAccAddKVStoreEntries(&kvstore), // triggers side-effect of adding a KVStore entry
			},
			{
				Config:      testAccServiceComputeConfigKVStore(name, storeName, domainName),
				ExpectError: regexp.MustCompile("cannot delete.*not empty.*"),
			},
			{
				Config: testAccServiceComputeConfigKVStoreForceDestroy(name, updatedStoreName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputeAttributesKVStore(&service, &kvstore, name, updatedStoreName),
				),
			},
			{
				Config: testAccServiceComputeConfigKVStoreForceDestroy(name, storeName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceComputeAttributesKVStore(&service, &kvstore, name, storeName),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceComputeAttributesKVStore(service *gofastly.ServiceDetail, kvstore *gofastly.KVStore, name, dictName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if service.Name != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		dict, err := conn.GetKVStore(&gofastly.GetKVStoreInput{
			ID: service.ID,
		})
		if err != nil {
			return fmt.Errorf("error looking up KV Store records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if dict.Name != dictName {
			return fmt.Errorf("kvstore name mismatch, expected: %s, got: %#v", dictName, dict.Name)
		}

		*kvstore = *dict

		return nil
	}
}

// testAccAddKVStoreEntries doesn't technically check for anything despite
// returning a TestCheckFunc. Instead it is used for its side effect of adding
// a key to the KV Store.
func testAccAddKVStoreEntries(kvstore *gofastly.KVStore) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		err := conn.InsertKVStoreKey(&gofastly.InsertKVStoreKeyInput{
			ID:    kvstore.ID,
			Key:   "testKey",
			Value: "testValue",
		})
		if err != nil {
			return fmt.Errorf("error adding item to KV Store (%s): %w", kvstore.ID, err)
		}
		return nil
	}
}

func testAccServiceComputeConfigKVStore(name, storeName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  kv_store {
    name = "%s"
  }

  force_destroy = true
}`, name, domainName, storeName)
}

func testAccServiceComputeConfigKVStoreForceDestroy(name, storeName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  kv_store {
    name          = "%s"
    force_destroy = true
  }

  force_destroy = true
}`, name, domainName, storeName)
}
