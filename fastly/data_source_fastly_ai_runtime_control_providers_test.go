package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceAIRuntimeControlProviders_Config(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "fastly_ai_runtime_control_providers" "example" {}`,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ai_runtime_control_providers.example"]
						a := r.Primary.Attributes

						count, err := strconv.Atoi(a["providers.#"])
						if err != nil {
							return fmt.Errorf("error parsing providers count: %s", err)
						}
						if count == 0 {
							return fmt.Errorf("expected at least one supported provider, got none")
						}

						// Every provider must be identifiable, and models are
						// nested within rather than fetched separately.
						for i := range count {
							if a[fmt.Sprintf("providers.%d.id", i)] == "" {
								return fmt.Errorf("provider %d has no id", i)
							}
							if _, ok := a[fmt.Sprintf("providers.%d.models.#", i)]; !ok {
								return fmt.Errorf("provider %d has no models attribute", i)
							}
						}

						return nil
					},
					resource.TestCheckResourceAttrSet("data.fastly_ai_runtime_control_providers.example", "total"),
				),
			},
		},
	})
}
