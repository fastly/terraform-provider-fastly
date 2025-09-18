package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/signals"
)

func TestAccFastlyNGWAFAccountSignal_basic(t *testing.T) {
	signalName := fmt.Sprintf("Signal Test %s", acctest.RandString(5))
	signalDescription := fmt.Sprintf("Account Signal %s", acctest.RandString(5))
	updatedDescription := signalDescription + " updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAccountSignalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAccountSignalConfig(signalName, signalDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_account_signal.example", "description", signalDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_signal.example", "name", signalName),
				),
			},
			{
				Config: testAccNGWAFAccountSignalConfig(signalName, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_account_signal.example", "description", updatedDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_signal.example", "name", signalName),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_account_signal.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					signal := s.RootModule().Resources["fastly_ngwaf_account_signal.example"]
					return signal.Primary.ID, nil
				},
			},
		},
	})
}

func testAccNGWAFAccountSignalConfig(name, description string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_account_signal" "example" {
  applies_to       = ["*"]
  name             = "%s"
  description      = "%s"
}
`, name, description)
}

func testAccCheckNGWAFAccountSignalDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_account_signal" {
			continue
		}

		_, err := signals.Get(context.TODO(), conn, &signals.GetInput{
			SignalID: &rs.Primary.ID,
			Scope: &scope.Scope{
				Type:      scope.ScopeTypeAccount,
				AppliesTo: []string{"*"},
			},
		})
		if err == nil {
			return fmt.Errorf("NGWAF account signal %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
