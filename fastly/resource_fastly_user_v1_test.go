package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccFastlyUserV1_basic(t *testing.T) {
	var user gofastly.User
	login := fmt.Sprintf("tf-test-%s@example.com", acctest.RandString(10))
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	role := "engineer"
	name2 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	role2 := "superuser"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserV1Config(login, name, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserV1Exists("fastly_user_v1.foo", &user),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "login", login),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "role", role),
				),
			},

			{
				Config: testAccUserV1Config(login, name2, role2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserV1Exists("fastly_user_v1.foo", &user),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "name", name2),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "role", role2),
				),
			},
		},
	})
}

func TestAccFastlyUserV1_updateLogin(t *testing.T) {
	var user gofastly.User
	login := fmt.Sprintf("tf-test-%s@example.com", acctest.RandString(10))
	login2 := fmt.Sprintf("tf-test-%s@example.com", acctest.RandString(10))
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	role := "engineer"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserV1Config(login, name, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserV1Exists("fastly_user_v1.foo", &user),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "login", login),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "role", role),
				),
			},

			{
				Config: testAccUserV1Config(login2, name, role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserV1Exists("fastly_user_v1.foo", &user),
					resource.TestCheckResourceAttr(
						"fastly_user_v1.foo", "login", login2),
				),
			},
		},
	})
}

func testAccCheckUserV1Exists(n string, user *gofastly.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User ID is set")
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
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

func testAccCheckUserV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_user_v1" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		u, err := conn.GetCurrentUser()
		l, err := conn.ListCustomerUsers(&gofastly.ListCustomerUsersInput{
			CustomerID: u.CustomerID,
		})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing users when deleting Fastly User (%s): %s", rs.Primary.ID, err)
		}

		for _, u := range l {
			if u.ID == rs.Primary.ID {
				// user still found
				return fmt.Errorf("[WARN] Tried deleting User (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccUserV1Config(login, name, role string) string {
	return fmt.Sprintf(`
resource "fastly_user_v1" "foo" {
	login = "%s"
	name  = "%s"
	role  = "%s"
}`, login, name, role)
}
