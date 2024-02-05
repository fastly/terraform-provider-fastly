package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyUser_basic(t *testing.T) {
	var user gofastly.User
	login := fmt.Sprintf("tf-test-%s@example.com", acctest.RandString(10))
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	role := "engineer"
	name2 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	role2 := "superuser"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(login, name, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("fastly_user.foo", &user),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "login", login),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "role", role),
				),
			},

			{
				Config: testAccUserConfig(login, name2, role2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("fastly_user.foo", &user),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "role", role2),
				),
			},
		},
	})
}

func TestAccFastlyUser_updateLogin(t *testing.T) {
	var user gofastly.User
	login := fmt.Sprintf("tf-test-%s@example.com", acctest.RandString(10))
	login2 := fmt.Sprintf("tf-test-%s@example.com", acctest.RandString(10))
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	role := "engineer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(login, name, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("fastly_user.foo", &user),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "login", login),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "role", role),
				),
			},

			{
				Config: testAccUserConfig(login2, name, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("fastly_user.foo", &user),
					resource.TestCheckResourceAttr(
						"fastly_user.foo", "login", login2),
				),
			},
		},
	})
}

func testAccCheckUserExists(n string, user *gofastly.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no User ID is set")
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := conn.GetUser(&gofastly.GetUserInput{
			ID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		*user = *latest

		return nil
	}
}

func testAccCheckUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_user" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		u, err := conn.GetCurrentUser()
		if err != nil {
			return fmt.Errorf("error getting current user when deleting Fastly User (%s): %s", rs.Primary.ID, err)
		}

		l, err := conn.ListCustomerUsers(&gofastly.ListCustomerUsersInput{
			CustomerID: gofastly.ToValue(u.CustomerID),
		})
		if err != nil {
			return fmt.Errorf("error listing users when deleting Fastly User (%s): %s", rs.Primary.ID, err)
		}

		for _, u := range l {
			if gofastly.ToValue(u.ID) == rs.Primary.ID {
				// user still found
				return fmt.Errorf("tried deleting User (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccUserConfig(login, name, role string) string {
	return fmt.Sprintf(`
resource "fastly_user" "foo" {
	login = "%s"
	name  = "%s"
	role  = "%s"
}`, login, name, role)
}
