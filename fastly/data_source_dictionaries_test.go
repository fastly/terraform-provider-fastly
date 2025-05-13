package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceDictionaries_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceDictionariesConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_dictionaries.example"]
						a := r.Primary.Attributes

						dictionaries, err := strconv.Atoi(a["dictionaries.#"])
						if err != nil {
							return err
						}

						if dictionaries != 3 {
							return fmt.Errorf("expected three dictionaries to be returned (as per the config)")
						}

						// NOTE: API doesn't guarantee dictionary order.
						for i := 0; i < 3; i++ {
							var found bool
							for _, d := range generateNames(h) {
								if a[fmt.Sprintf("dictionaries.%d.name", i)] == d {
									found = true
									break
								}
							}
							if !found {
								want := fmt.Sprintf("tf_%s_%d", h, i+1)
								return fmt.Errorf("expected: %s", want)
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceDictionariesConfig(h string) string {
	tf := `
resource "fastly_service_vcl" "example" {
  name = "tf_example_service_for_dictionaries_data_source"

  domain {
		name = "%s.com"
	}

  dictionary {
    name       = "tf_%s_1"
  }

  dictionary {
    name       = "tf_%s_2"
  }

  dictionary {
    name       = "tf_%s_3"
  }

  force_destroy = true
}

data "fastly_dictionaries" "example" {
  depends_on      = [fastly_service_vcl.example]
  service_id      = fastly_service_vcl.example.id
  service_version = fastly_service_vcl.example.active_version
}
`

	return fmt.Sprintf(tf, h, h, h, h)
}
