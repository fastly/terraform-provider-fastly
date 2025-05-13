package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceKVStores_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceKVStoresConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_kvstores.example"]
						a := r.Primary.Attributes

						want := generateNames(h)
						var (
							found int
							got   []string
						)

						// NOTE: API doesn't guarantee KV Store order.
						for k, v := range a {
							// Example of keys we're looking for:
							// "stores.0.name":"tf_677f63804c9351ac31fd0cb1db697b95_1",
							// "stores.1.name":"tf_677f63804c9351ac31fd0cb1db697b95_2",
							// "stores.2.name":"tf_677f63804c9351ac31fd0cb1db697b95_3",
							if strings.HasSuffix(k, ".name") {
								got = append(got, v)
								for _, name := range want {
									if v == name {
										found++
										break
									}
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

func testAccFastlyDataSourceKVStoresConfig(h string) string {
	tf := `
resource "fastly_kvstore" "example_1" {
  name          = "tf_%s_1"
  force_destroy = true
}

resource "fastly_kvstore" "example_2" {
  name          = "tf_%s_2"
  force_destroy = true
}

resource "fastly_kvstore" "example_3" {
  name          = "tf_%s_3"
  force_destroy = true
}

data "fastly_kvstores" "example" {
  depends_on = [fastly_kvstore.example_1, fastly_kvstore.example_2, fastly_kvstore.example_3]
}
`

	return fmt.Sprintf(tf, h, h, h)
}
