package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceNGWAFRedactions_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFRedactionsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_redactions.example"]
						a := r.Primary.Attributes
						var (
							found int
							got   []string
						)
						for k, v := range a {
							if strings.HasSuffix(k, ".field") {
								got = append(got, v)
								if strings.Contains(v, h) {
									found++
								}
							}
						}

						if found != 3 {
							return fmt.Errorf("want: %v, got: %v", h, got)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFRedactionsConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_redactions_workspace" {
  name                         = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_redaction" "example_1" {
  field                        = "%s 1"
  type                         = "request_header"
  workspace_id                 = fastly_ngwaf_workspace.test_redactions_workspace.id
}

resource "fastly_ngwaf_redaction" "example_2" {
  field                        = "%s 2"
  type                         = "request_header"
  workspace_id                 = fastly_ngwaf_workspace.test_redactions_workspace.id
}

resource "fastly_ngwaf_redaction" "example_3" {
  field                        = "%s 3"
  type                         = "request_header"
  workspace_id                 = fastly_ngwaf_workspace.test_redactions_workspace.id
}

data "fastly_ngwaf_redactions" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_redactions_workspace,
    fastly_ngwaf_redaction.example_1,
    fastly_ngwaf_redaction.example_2,
    fastly_ngwaf_redaction.example_3
  ]
  workspace_id = fastly_ngwaf_workspace.test_redactions_workspace.id
}
`
	return fmt.Sprintf(tf, h, h, h, h)
}
