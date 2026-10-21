package fastly

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/fastly/go-fastly/v18/fastly/airuntimecontrol/v1/key"
)

func TestAccFastlyAIRuntimeControlVirtualKey_basic(t *testing.T) {
	keyName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	updatedName := fmt.Sprintf("%s-updated", keyName)
	expiresAt := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAIRuntimeControlVirtualKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAIRuntimeControlVirtualKeyConfig(keyName, "claude-sonnet-4-20250514", expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAIRuntimeControlVirtualKeyRemoteState("fastly_ai_runtime_control_virtual_key.foo", keyName, "claude-sonnet-4-20250514"),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_virtual_key.foo", "name", keyName),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_virtual_key.foo", "provider_name", "Anthropic"),
					resource.TestCheckResourceAttrSet("fastly_ai_runtime_control_virtual_key.foo", "created_at"),
					resource.TestCheckResourceAttrSet("fastly_ai_runtime_control_virtual_key.foo", "access_token"),
				),
			},
			{
				Config: testAccAIRuntimeControlVirtualKeyConfig(updatedName, "claude-haiku-4-5-20251001", expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAIRuntimeControlVirtualKeyRemoteState("fastly_ai_runtime_control_virtual_key.foo", updatedName, "claude-haiku-4-5-20251001"),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_virtual_key.foo", "name", updatedName),
				),
			},
			{
				ResourceName:      "fastly_ai_runtime_control_virtual_key.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// The API only returns the access token on create; it cannot
				// be recovered via import.
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccCheckAIRuntimeControlVirtualKeyRemoteState(resourceName, expectedName, expectedModel string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		got, err := key.Get(context.TODO(), conn, &key.GetInput{
			KeyID: new(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("error fetching AI Runtime Control virtual key (%s): %s", rs.Primary.ID, err)
		}

		if got.Name != expectedName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", expectedName, got.Name)
		}
		if got.Model != expectedModel {
			return fmt.Errorf("bad model, expected (%s), got (%s)", expectedModel, got.Model)
		}

		return nil
	}
}

func testAccCheckAIRuntimeControlVirtualKeyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ai_runtime_control_virtual_key" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		got, err := key.Get(context.TODO(), conn, &key.GetInput{
			KeyID: new(rs.Primary.ID),
		})
		if err != nil {
			continue
		}
		// Virtual keys are soft deleted, so a deleted key may still be
		// returned rather than 404ing.
		if got.DeletedAt == nil {
			return fmt.Errorf("tried deleting AI Runtime Control virtual key (%s), but was still found", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAIRuntimeControlVirtualKeyConfig(name, model, expiresAt string) string {
	return fmt.Sprintf(`
resource "fastly_ai_runtime_control_virtual_key" "foo" {
  name          = "%s"
  model         = "%s"
  provider_name = "Anthropic"
  expires_at    = "%s"
}`, name, model, expiresAt)
}
