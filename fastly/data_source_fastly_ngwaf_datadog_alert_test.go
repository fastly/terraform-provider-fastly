package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceNGWAFDatadogAlert_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFDatadogAlertConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_datadog_alerts.example"]
						a := r.Primary.Attributes
						var (
							found int
							got   []string
						)
						for k, v := range a {
							if strings.HasSuffix(k, ".description") {
								got = append(got, v)
								if strings.Contains(v, h) {
									found++
								}
							}
						}

						if found != 1 {
							return fmt.Errorf("want: %v, got: %v", h, got)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFDatadogAlertConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_datadog_alerts_workspace" {
  name                             = "%s"
  description                      = "Test NGWAF Workspace"
  mode                             = "block"
  ip_anonymization                 = "hashed"
  client_ip_headers                = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code   = 429

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_datadog_alert" "example_1" {
  description      = "%s 1"
  key              = "123456789"
  site             = "us1"
  workspace_id     = fastly_ngwaf_workspace.test_datadog_alerts_workspace.id
}

data "fastly_ngwaf_datadog_alerts" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_datadog_alerts_workspace,
    fastly_ngwaf_datadog_alert.example_1
  ]
  workspace_id = fastly_ngwaf_workspace.test_datadog_alerts_workspace.id
}
`
	return fmt.Sprintf(tf, h, h)
}
