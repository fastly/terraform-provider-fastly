package fastly

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"
)

func TestAccFastlyAIRuntimeControlProviderConnection_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAIRuntimeControlProviderConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAIRuntimeControlProviderConnectionConfig(name, "https://api.openai.com", []string{"gpt-4o"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAIRuntimeControlProviderConnectionRemoteState("fastly_ai_runtime_control_provider_connection.foo", name, []string{"gpt-4o"}),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_provider_connection.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_provider_connection.foo", "models.#", "1"),
					resource.TestCheckResourceAttrSet("fastly_ai_runtime_control_provider_connection.foo", "created_at"),
				),
			},
			{
				Config: testAccAIRuntimeControlProviderConnectionConfig(name, "https://api.openai.com", []string{"gpt-4o", "gpt-4o-mini"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAIRuntimeControlProviderConnectionRemoteState("fastly_ai_runtime_control_provider_connection.foo", name, []string{"gpt-4o", "gpt-4o-mini"}),
					resource.TestCheckResourceAttr("fastly_ai_runtime_control_provider_connection.foo", "models.#", "2"),
				),
			},
			{
				ResourceName:      "fastly_ai_runtime_control_provider_connection.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// The API doesn't return the provider's secret key.
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccCheckAIRuntimeControlProviderConnectionRemoteState(resourceName, expectedName string, expectedModels []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		got, err := providerconnection.Get(context.TODO(), conn, &providerconnection.GetInput{
			ID: gofastly.ToPointer(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("error fetching AI Runtime Control provider connection (%s): %s", rs.Primary.ID, err)
		}

		if got.Name != expectedName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", expectedName, got.Name)
		}
		if len(got.Models) != len(expectedModels) {
			return fmt.Errorf("bad models, expected (%v), got (%v)", expectedModels, got.Models)
		}

		for _, want := range expectedModels {
			var found bool
			for _, have := range got.Models {
				if want == have {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("bad models, expected (%v) to contain (%s)", got.Models, want)
			}
		}

		return nil
	}
}

func testAccCheckAIRuntimeControlProviderConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ai_runtime_control_provider_connection" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		_, err := providerconnection.Get(context.TODO(), conn, &providerconnection.GetInput{
			ID: gofastly.ToPointer(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("tried deleting AI Runtime Control provider connection (%s), but was still found", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAIRuntimeControlProviderConnectionConfig(name, baseURL string, models []string) string {
	quoted := make([]string, len(models))
	for i, m := range models {
		quoted[i] = fmt.Sprintf("%q", m)
	}

	return fmt.Sprintf(`
resource "fastly_ai_runtime_control_provider_connection" "foo" {
  name     = "%s"
  base_url = "%s"
  api_key  = "tf-test-secret-key"
  models   = [%s]
}`, name, baseURL, strings.Join(quoted, ", "))
}
