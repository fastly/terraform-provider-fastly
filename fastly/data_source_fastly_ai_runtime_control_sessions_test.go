package fastly

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceAIRuntimeControlSessions_Config(t *testing.T) {
	from := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAIRuntimeControlSessionsConfig(from),
				Check: resource.ComposeTestCheckFunc(
					// An account with no AI traffic returns no sessions, so this
					// asserts the shape of the result rather than its contents.
					resource.TestCheckResourceAttrSet("data.fastly_ai_runtime_control_sessions.example", "total"),
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_ai_runtime_control_sessions.example"]
						a := r.Primary.Attributes

						count, err := strconv.Atoi(a["sessions.#"])
						if err != nil {
							return fmt.Errorf("error parsing sessions count: %s", err)
						}
						if a["total"] != strconv.Itoa(count) {
							return fmt.Errorf("total (%s) does not match sessions count (%d)", a["total"], count)
						}

						for i := range count {
							if a[fmt.Sprintf("sessions.%d.id", i)] == "" {
								return fmt.Errorf("session %d has no id", i)
							}
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceAIRuntimeControlSessionsConfig(from string) string {
	tf := `
data "fastly_ai_runtime_control_sessions" "example" {
  from = "%s"
}
`
	return fmt.Sprintf(tf, from)
}
