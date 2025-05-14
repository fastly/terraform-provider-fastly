package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceVCLSnippets_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceVCLSnippetsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_vcl_snippets.example"]
						a := r.Primary.Attributes

						snippets, err := strconv.Atoi(a["vcl_snippets.#"])
						if err != nil {
							return err
						}

						if snippets != 3 {
							return fmt.Errorf("expected three snippets to be returned (as per the config)")
						}

						// NOTE: API doesn't guarantee order.
						for i := 0; i < 3; i++ {
							var found bool
							for _, d := range generateNames(h) {
								if a[fmt.Sprintf("vcl_snippets.%d.name", i)] == d {
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

func testAccFastlyDataSourceVCLSnippetsConfig(h string) string {
	tf := `
resource "fastly_service_vcl" "example" {
  name = "tf_example_service_for_vcl_snippets_data_source"

  domain {
    name = "%s.com"
  }

  snippet {
    name     = "tf_%s_1"
    content  = "# EXAMPLE 1"
    type     = "init"
    priority = 1
  }

  snippet {
    name     = "tf_%s_2"
    content  = "# EXAMPLE 2"
    type     = "init"
    priority = 2
  }

  snippet {
    name     = "tf_%s_3"
    content  = "# EXAMPLE 3"
    type     = "init"
    priority = 3
  }

  force_destroy = true
}

data "fastly_vcl_snippets" "example" {
  depends_on      = [fastly_service_vcl.example]
  service_id      = fastly_service_vcl.example.id
  service_version = fastly_service_vcl.example.active_version
}
`

	return fmt.Sprintf(tf, h, h, h, h)
}
