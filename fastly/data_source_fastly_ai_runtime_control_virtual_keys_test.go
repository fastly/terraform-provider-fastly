package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceAIRuntimeControlVirtualKeys_Config(t *testing.T) {
	h := generateHex()
	userID := testAccAIRuntimeControlUserID(t)
	keyName := fmt.Sprintf("tf_%s", h)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAIRuntimeControlVirtualKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAIRuntimeControlVirtualKeysConfig(keyName, userID),
				Check: resource.ComposeTestCheckFunc(
					// The search filter is a substring match on the key name,
					// so the generated name isolates this test's key.
					resource.TestCheckResourceAttr("data.fastly_ai_runtime_control_virtual_keys.example", "virtual_keys.#", "1"),
					resource.TestCheckResourceAttr("data.fastly_ai_runtime_control_virtual_keys.example", "virtual_keys.0.name", keyName),
					resource.TestCheckResourceAttr("data.fastly_ai_runtime_control_virtual_keys.example", "virtual_keys.0.provider", "Anthropic"),
					resource.TestCheckResourceAttr("data.fastly_ai_runtime_control_virtual_keys.example", "total", "1"),
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ai_runtime_control_virtual_keys.example"]
						if got := r.Primary.Attributes["virtual_keys.0.id"]; got == "" {
							return fmt.Errorf("virtual key has no id")
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceAIRuntimeControlVirtualKeysConfig(keyName, userID string) string {
	tf := `
resource "fastly_ai_runtime_control_virtual_key" "example" {
  name          = "%s"
  model         = "claude-sonnet-4-20250514"
  provider_name = "Anthropic"
  user_id       = "%s"
}

data "fastly_ai_runtime_control_virtual_keys" "example" {
  search = "%s"

  depends_on = [fastly_ai_runtime_control_virtual_key.example]
}
`
	return fmt.Sprintf(tf, keyName, userID, keyName)
}
