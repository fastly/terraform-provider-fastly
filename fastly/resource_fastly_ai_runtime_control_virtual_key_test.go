package fastly

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"
)

func TestAccFastlyAIRuntimeControlVirtualKey_basic(t *testing.T) {
	keyName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	updatedName := fmt.Sprintf("%s-updated", keyName)
	userID := testAccAIRuntimeControlUserID(t)
	expiresAt := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAIRuntimeControlVirtualKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAIRuntimeControlVirtualKeyConfig(keyName, "claude-sonnet-4-20250514", userID, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAIRuntimeControlVirtualKeyRemoteState("fastly_ai_runtime_control_virtual_key.foo", keyName, "claude-sonnet-4-20250514"),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_virtual_key.foo", "name", keyName),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_virtual_key.foo", "provider_name", "Anthropic"),
					resource.TestCheckResourceAttrSet("fastly_ai_runtime_control_virtual_key.foo", "created_at"),
				),
			},
			{
				Config: testAccAIRuntimeControlVirtualKeyConfig(updatedName, "claude-haiku-4-5-20251001", userID, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAIRuntimeControlVirtualKeyRemoteState("fastly_ai_runtime_control_virtual_key.foo", updatedName, "claude-haiku-4-5-20251001"),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_virtual_key.foo", "name", updatedName),
				),
			},
			{
				ResourceName:      "fastly_ai_runtime_control_virtual_key.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccAIRuntimeControlUserID resolves the authenticated user's ID, which the
// create endpoint requires. The config it feeds is built before
// resource.ParallelTest applies its own TF_ACC gate, so this repeats that check
// rather than calling the API during a unit test run.
func testAccAIRuntimeControlUserID(t *testing.T) string {
	t.Helper()

	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
	}

	apiKey := os.Getenv("FASTLY_API_KEY")
	if apiKey == "" {
		t.Fatal("FASTLY_API_KEY must be set for acceptance tests")
	}

	var (
		conn *gofastly.Client
		err  error
	)
	if endpoint := os.Getenv("FASTLY_API_URL"); endpoint != "" {
		conn, err = gofastly.NewClientForEndpoint(apiKey, endpoint)
	} else {
		conn, err = gofastly.NewClient(apiKey)
	}
	if err != nil {
		t.Fatalf("error creating Fastly client: %s", err)
	}

	user, err := conn.GetCurrentUser(context.TODO())
	if err != nil {
		t.Fatalf("error fetching current user: %s", err)
	}

	return gofastly.ToValue(user.UserID)
}

func testAccCheckAIRuntimeControlVirtualKeyRemoteState(resourceName, expectedName, expectedModel string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		got, err := key.Get(context.TODO(), conn, &key.GetInput{
			KeyID: gofastly.ToPointer(rs.Primary.ID),
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
			KeyID: gofastly.ToPointer(rs.Primary.ID),
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

func testAccAIRuntimeControlVirtualKeyConfig(name, model, userID, expiresAt string) string {
	return fmt.Sprintf(`
resource "fastly_ai_runtime_control_virtual_key" "foo" {
  name          = "%s"
  model         = "%s"
  provider_name = "Anthropic"
  user_id       = "%s"
  expires_at    = "%s"
}`, name, model, userID, expiresAt)
}
