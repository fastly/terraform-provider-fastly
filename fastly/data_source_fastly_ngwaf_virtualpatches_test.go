package fastly

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceNGWAFVirtualPatches_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFVirtualPatchesConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_virtualpatches.sample"]
						a := r.Primary.Attributes

						want := []string{"CVE-2017-5638", "CVE-2019-0193", "CVE-2021-44228"} // Expected virtual patch IDs
						var (
							found int
							got   []string
						)

						// Check for virtual patch IDs like:
						// "virtualpatches.0.id": "CVE-2017-5638"
						// "virtualpatches.1.id": "CVE-2019-0193"
						// "virtualpatches.1.id": "CVE-2021-44228"
						for k, v := range a {
							if strings.HasSuffix(k, ".id") {
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

func testAccFastlyDataSourceNGWAFVirtualPatchesConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "example_1" {
  name = "tf_%s_1"
  description = "Test NGWAF Workspace %s_1"
  mode = "block"
  ip_anonymization = "hashed"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}
data "fastly_ngwaf_virtualpatches" "sample" {
	workspace_id = fastly_ngwaf_workspace.example.id
}
`
	return fmt.Sprintf(tf, h, h)
}
