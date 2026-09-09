package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceAIRuntimeControlProviderConnections_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAIRuntimeControlProviderConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAIRuntimeControlProviderConnectionsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ai_runtime_control_provider_connections.example"]
						a := r.Primary.Attributes

						want := []string{fmt.Sprintf("tf_%s_1", h), fmt.Sprintf("tf_%s_2", h)}
						var (
							found int
							got   []string
						)

						for k, v := range a {
							if !strings.HasSuffix(k, ".name") {
								continue
							}
							got = append(got, v)
							for _, name := range want {
								if v == name {
									found++
									break
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

func testAccFastlyDataSourceAIRuntimeControlProviderConnectionsConfig(h string) string {
	tf := `
resource "fastly_ai_runtime_control_provider_connection" "example_1" {
  name     = "tf_%s_1"
  base_url = "https://api.openai.com"
  api_key  = "tf-test-secret-key"
  models   = ["gpt-4o"]
}

resource "fastly_ai_runtime_control_provider_connection" "example_2" {
  name     = "tf_%s_2"
  base_url = "https://api.anthropic.com"
  api_key  = "tf-test-secret-key"
  models   = ["claude-sonnet-4-20250514"]
}

data "fastly_ai_runtime_control_provider_connections" "example" {
  depends_on = [
    fastly_ai_runtime_control_provider_connection.example_1,
    fastly_ai_runtime_control_provider_connection.example_2
  ]
}
`
	return fmt.Sprintf(tf, h, h)
}
