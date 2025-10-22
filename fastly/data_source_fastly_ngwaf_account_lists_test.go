package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyNGWAFAccountLists_Config(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceNGWAFAccountListsConfig(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ngwaf_account_lists.test"]
						a := r.Primary.Attributes

						// Check that we have the lists attribute
						listCount := a["lists.#"]
						if listCount == "" {
							return fmt.Errorf("expected lists attribute to be present")
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

func testAccFastlyDataSourceNGWAFAccountListsConfig() string {
	return `
data "fastly_ngwaf_account_lists" "test" {
}
`
}
