package acceptancetests

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyKVStore_basic(t *testing.T) {
	t.Parallel()
	storeName := fmt.Sprintf("tf_test_kvstore_%s", acctest.RandString(10))
	storeNameUpdated := fmt.Sprintf("tf_test_kvstore_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckKVStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: ConfigKVStore(storeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_kvstore.store", "name", storeName),
					resource.TestCheckResourceAttr("fastly_kvstore.store", "force_destroy", "false"),
					resource.TestCheckResourceAttrSet("fastly_kvstore.store", "id"),
				),
			},
			{
				// Changing the name forces replacement.
				Config: ConfigKVStore(storeNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_kvstore.store", "name", storeNameUpdated),
					resource.TestCheckResourceAttrSet("fastly_kvstore.store", "id"),
				),
			},
			{
				ResourceName:      "fastly_kvstore.store",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyKVStore_forceDestroy(t *testing.T) {
	t.Parallel()
	storeName := fmt.Sprintf("tf_test_kvstore_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckKVStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: ConfigKVStoreForceDestroy(storeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_kvstore.store", "name", storeName),
					resource.TestCheckResourceAttr("fastly_kvstore.store", "force_destroy", "true"),
					InsertKVStoreKey("fastly_kvstore.store", "test-key", "test-value"),
				),
			},
		},
	})
}

// InsertKVStoreKey inserts a key/value pair into the KV Store managed by the
// given resource address, so that force_destroy behavior can be exercised
// when the resource is destroyed at the end of the test.
func InsertKVStoreKey(resourceName, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		return client.InsertKVStoreKey(context.Background(), &fastly.InsertKVStoreKeyInput{
			StoreID: rs.Primary.ID,
			Key:     key,
			Value:   value,
		})
	}
}

// CheckServiceAndKVStoreDestroy composes CheckServiceDestroy for the given service resource
// type with CheckKVStoreDestroy, for tests that manage both a service and one or more
// fastly_kvstore resources.
func CheckServiceAndKVStoreDestroy(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if err := CheckServiceDestroy(resourceType)(s); err != nil {
			return err
		}
		return CheckKVStoreDestroy(s)
	}
}

func CheckKVStoreDestroy(s *terraform.State) error {
	client, err := NewFastlyClient()
	if err != nil {
		return fmt.Errorf("error creating Fastly client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_kvstore" {
			continue
		}

		id := rs.Primary.ID
		_, err := client.GetKVStore(context.Background(), &fastly.GetKVStoreInput{StoreID: id})
		if errors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking if KV Store was destroyed: %w", err)
		}

		return fmt.Errorf("KV Store %s still exists", id)
	}

	return nil
}

func TestAccFastlyDataSourceKVStores(t *testing.T) {
	t.Parallel()
	h := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckKVStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: ConfigKVStoresDataSource(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["data.fastly_kvstores.example"]
						if !ok {
							return fmt.Errorf("not found: data.fastly_kvstores.example")
						}

						want := []string{
							fmt.Sprintf("tf_%s_1", h),
							fmt.Sprintf("tf_%s_2", h),
							fmt.Sprintf("tf_%s_3", h),
						}

						var found int
						var got []string
						for k, v := range rs.Primary.Attributes {
							if strings.HasSuffix(k, ".name") {
								got = append(got, v)
								if slices.Contains(want, v) {
									found++
								}
							}
						}

						if found != len(want) {
							return fmt.Errorf("want: %v, got: %v", want, got)
						}

						return nil
					},
				),
			},
		},
	})
}
