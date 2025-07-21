package fastly

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceComputeACLs_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceComputeACLsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_compute_acls.example"]
						a := r.Primary.Attributes

						want := generateNames(h)
						var (
							found int
							got   []string
						)

						for k, v := range a {
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

func testAccFastlyDataSourceComputeACLsConfig(h string) string {
	tf := `
resource "fastly_compute_acl" "example_1" {
  name = "tf_%s_1"
}

resource "fastly_compute_acl" "example_2" {
  name = "tf_%s_2"
}

resource "fastly_compute_acl" "example_3" {
  name = "tf_%s_3"
}

data "fastly_compute_acls" "example" {
  depends_on = [
    fastly_compute_acl.example_1,
    fastly_compute_acl.example_2,
    fastly_compute_acl.example_3
  ]
}
`
	return fmt.Sprintf(tf, h, h, h)
}
