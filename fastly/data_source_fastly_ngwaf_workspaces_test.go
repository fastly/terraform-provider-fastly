package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceNGWAFWorkspaces_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFWorkspacesConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_workspaces.example"]
						a := r.Primary.Attributes

						want := generateNames(h) // e.g., ["tf_<hex>_1", "tf_<hex>_2", "tf_<hex>_3"]
						var (
							found int
							got   []string
						)

						// We're expecting attributes like:
						// "workspaces.0.name": "tf_<hex>_1",
						// "workspaces.1.name": "tf_<hex>_2",
						// "workspaces.2.name": "tf_<hex>_3",
						for k, v := range a {
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

func testAccFastlyDataSourceNGWAFWorkspacesConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "example_1" {
  name = "tf_%s_1"
  description = "Test NGWAF Workspace 1"
  mode = "block"
  ip_anonymization = "hashed"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace" "example_2" {
  name = "tf_%s_2"
  description = "Test NGWAF Workspace 2"
  mode = "block"
  ip_anonymization = "hashed"

  attack_signal_thresholds {
    one_minute  = 200
    ten_minutes = 1000
    one_hour    = 2000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace" "example_3" {
  name = "tf_%s_3"
  description = "Test NGWAF Workspace 3"
  mode = "block"
  ip_anonymization = "hashed"

  attack_signal_thresholds {
    one_minute  = 300
    ten_minutes = 1500
    one_hour    = 2500
    immediate   = true
  }
}

data "fastly_ngwaf_workspaces" "example" {
  depends_on = [
    fastly_ngwaf_workspace.example_1,
    fastly_ngwaf_workspace.example_2,
    fastly_ngwaf_workspace.example_3
  ]
}
`
	return fmt.Sprintf(tf, h, h, h)
}
