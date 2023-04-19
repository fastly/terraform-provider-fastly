package fastly

import (
	"crypto/rand"
	"encoding/hex"
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
							for _, d := range genDictNames(h) {
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

// generateHex produces a slice of 16 random bytes.
// This is useful for dynamically generating resource names.
func generateHex() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func genDictNames(h string) []string {
	dicts := []string{}
	for i := 1; i < 4; i++ {
		dicts = append(dicts, fmt.Sprintf("tf_%s_%d", h, i))
	}
	return dicts
}
