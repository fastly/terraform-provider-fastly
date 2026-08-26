package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFWorkspaceLists_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFWorkspaceListsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_workspace_lists.test"]
						a := r.Primary.Attributes

						// Check that we have the lists attribute
						listCount := a["lists.#"]
						if listCount == "" {
							return fmt.Errorf("expected lists attribute to be present")
						}

						// Check that workspace_id is set
						workspaceID := a["workspace_id"]
						if workspaceID == "" {
							return fmt.Errorf("expected workspace_id to be set")
						}

						// If there are lists, check that the first one has an ID
						if listCount != "0" {
							listID := a["lists.0.id"]
							if listID == "" {
								return fmt.Errorf("expected list to have an ID, got empty string")
							}

							// Check that list has required fields
							listName := a["lists.0.name"]
							if listName == "" {
								return fmt.Errorf("expected list to have a name, got empty string")
							}

							listType := a["lists.0.type"]
							if listType == "" {
								return fmt.Errorf("expected list to have a type, got empty string")
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceNGWAFWorkspaceListsConfig(h string) string {
	return fmt.Sprintf(`
%s

data "fastly_ngwaf_workspace_lists" "test" {
  workspace_id = fastly_ngwaf_workspace.example.id
}
`, testAccNGWAFWorkspaceConfig(fmt.Sprintf("tf_%s", h)))
}
