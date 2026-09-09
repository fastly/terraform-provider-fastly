package fastly

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceAIRuntimeControlUsageMetrics_Config(t *testing.T) {
	from := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAIRuntimeControlUsageMetricsConfig(from),
				Check: resource.ComposeTestCheckFunc(
					// An account with no AI traffic returns no records, so this
					// asserts the shape of the result rather than its contents.
					resource.TestCheckResourceAttrSet("data.fastly_ai_runtime_control_usage_metrics.example", "total"),
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ai_runtime_control_usage_metrics.example"]
						a := r.Primary.Attributes

						count, err := strconv.Atoi(a["usage_metrics.#"])
						if err != nil {
							return fmt.Errorf("error parsing usage_metrics count: %s", err)
						}
						if a["total"] != strconv.Itoa(count) {
							return fmt.Errorf("total (%s) does not match usage_metrics count (%d)", a["total"], count)
						}

						for i := range count {
							usageType := a[fmt.Sprintf("usage_metrics.%d.usage_type", i)]
							switch usageType {
							case "requests", "sessions", "input_tokens", "output_tokens":
							default:
								return fmt.Errorf("unexpected usage_type (%s) at index %d", usageType, i)
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceAIRuntimeControlUsageMetricsConfig(from string) string {
	tf := `
data "fastly_ai_runtime_control_usage_metrics" "example" {
  from = "%s"
}
`
	return fmt.Sprintf(tf, from)
}
