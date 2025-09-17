package fastly

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/lists"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

func TestAccFastlyNGWAFAccountList_basic(t *testing.T) {
	listName := fmt.Sprintf("acc-list-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAccountListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAccountListConfig(listName, "Initial account list", []string{"1.2.3.4"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_account_list.example", "name", listName),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_list.example", "type", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_list.example", "description", "Initial account list"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_list.example", "entries.0", "1.2.3.4"),
				),
			},
			{
				Config: testAccNGWAFAccountListConfig(listName, "Updated account list", []string{"192.168.1.1", "172.16.0.1"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_account_list.example", "description", "Updated account list"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_list.example", "entries.0", "192.168.1.1"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_list.example", "entries.1", "172.16.0.1"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_account_list.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					list := s.RootModule().Resources["fastly_ngwaf_account_list.example"]
					return list.Primary.ID, nil
				},
			},
		},
	})
}

func testAccNGWAFAccountListConfig(name, description string, entries []string) string {
	quoted := make([]string, len(entries))
	for i, entry := range entries {
		quoted[i] = fmt.Sprintf(`"%s"`, entry)
	}
	return fmt.Sprintf(`
resource "fastly_ngwaf_account_list" "example" {
  name        = "%s"
  description = "%s"
  type        = "ip"
  entries     = [%s]
}
`, name, description, strings.Join(quoted, ", "))
}

func testAccCheckNGWAFAccountListDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_account_list" {
			continue
		}

		_, err := lists.Get(context.Background(), conn, &lists.GetInput{
			ListID: &rs.Primary.ID,
			Scope: &scope.Scope{
				Type:      scope.ScopeTypeAccount,
				AppliesTo: []string{"*"},
			},
		})
		if err == nil {
			return fmt.Errorf("NGWAF account list %s still exists", rs.Primary.ID)
		}
	}
	return nil
}
