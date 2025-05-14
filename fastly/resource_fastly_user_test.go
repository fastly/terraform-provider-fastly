package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
)

const fastlyUser = "fastly_user.foo"

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
					testAccCheckFastlyUserExists(&user),
					resource.TestCheckResourceAttr(
						fastlyUser, "login", login),
					resource.TestCheckResourceAttr(
						fastlyUser, "name", name),
					resource.TestCheckResourceAttr(
						fastlyUser, "role", role),
				),
			},

			{
				Config: testAccUserConfig(login, name2, role2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyUserExists(&user),
					resource.TestCheckResourceAttr(
						fastlyUser, "name", name2),
					resource.TestCheckResourceAttr(
						fastlyUser, "role", role2),
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
					testAccCheckFastlyUserExists(&user),
					resource.TestCheckResourceAttr(
						fastlyUser, "login", login),
					resource.TestCheckResourceAttr(
						fastlyUser, "name", name),
					resource.TestCheckResourceAttr(
						fastlyUser, "role", role),
				),
			},

			{
				Config: testAccUserConfig(login2, name, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyUserExists(&user),
					resource.TestCheckResourceAttr(
						fastlyUser, "login", login2),
				),
			},
		},
	})
}

func testAccCheckFastlyUserExists(user *gofastly.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[fastlyUser]
		if !ok {
			return fmt.Errorf("not found: %s", fastlyUser)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no User ID is set")
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		latest, err := conn.GetUser(&gofastly.GetUserInput{
			UserID: rs.Primary.ID,
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
			if gofastly.ToValue(u.UserID) == rs.Primary.ID {
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
