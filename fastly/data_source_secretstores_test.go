package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceSecretsStores_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceSecretsStoresConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_secretstores.example"]
						a := r.Primary.Attributes

						want := generateNames(h, 3)
						var (
							found int
							got   []string
						)

						// NOTE: API doesn't guarantee Secrets Store order.
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

func testAccFastlyDataSourceSecretsStoresConfig(h string) string {
	tf := `
resource "fastly_secretstore" "example_1" {
  name          = "tf_%s_1"
}

resource "fastly_secretstore" "example_2" {
  name          = "tf_%s_2"
}

resource "fastly_secretstore" "example_3" {
  name          = "tf_%s_3"
}

data "fastly_secretstores" "example" {
  depends_on = [fastly_secretstore.example_1, fastly_secretstore.example_2, fastly_secretstore.example_3]
}
`

	return fmt.Sprintf(tf, h, h, h)
}
