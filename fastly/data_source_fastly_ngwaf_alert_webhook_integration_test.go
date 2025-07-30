package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFAlertWebhookIntegration_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAlertWebhookIntegrationConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_alert_webhook_integration.example"]
						a := r.Primary.Attributes

						// Check that we have at least one webhook alert with an ID
						webhookCount := a["webhook_alerts.#"]
						if webhookCount == "0" {
							return fmt.Errorf("expected at least one webhook alert, got %s", webhookCount)
						}

						// Check that the first webhook alert has an ID
						webhookID := a["webhook_alerts.0.id"]
						if webhookID == "" {
							return fmt.Errorf("expected webhook alert to have an ID, got empty string")
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFAlertWebhookIntegrationConfig(h string) string {
	tf := `
resource "fastly_ngwaf_workspace" "test_webhook_alerts_workspace" {
  name                             = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_alert_webhook_integration" "sample" {
  description      = "%s 1"
  webhook          = "https://example.com/webhooks/my-service"
  workspace_id     = fastly_ngwaf_workspace.test_webhook_alerts_workspace.id
}

data "fastly_ngwaf_alert_webhook_integration" "example" {
  depends_on = [
    fastly_ngwaf_workspace.test_webhook_alerts_workspace,
    fastly_ngwaf_alert_webhook_integration.sample
  ]
  workspace_id = fastly_ngwaf_workspace.test_webhook_alerts_workspace.id
}
`
	return fmt.Sprintf(tf, h, h)
}
