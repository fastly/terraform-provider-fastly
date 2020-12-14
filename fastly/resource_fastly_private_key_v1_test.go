package fastly

import (
	"fmt"
	"strings"
	"testing"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccFastlyPrivateKeyV1_basic(t *testing.T) {
	var pk gofastly.PrivateKey
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key, _, err := generateKeyAndCert()
	if err != nil {
		t.Fatal(err)
	}
	key = strings.ReplaceAll(key, "\n", `\n`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPrivateKeyV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateKeyV1Config(key, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateKeyV1Exists("fastly_private_key_v1.foo", &pk),
					resource.TestCheckResourceAttr(
						"fastly_private_key_v1.foo", "name", name),
				),
			},
		},
	})
}

func TestAccFastlyPrivateKeyV1_import(t *testing.T) {
	var pk gofastly.PrivateKey
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	key, _, err := generateKeyAndCert()
	if err != nil {
		t.Fatal(err)
	}
	key = strings.ReplaceAll(key, "\n", `\n`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateKeyV1Config(key, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateKeyV1Exists("fastly_private_key_v1.foo", &pk),
				),
			},
			{
				ResourceName:            "fastly_private_key_v1.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
		},
	})
}

func testAccCheckPrivateKeyV1Exists(n string, pk *gofastly.PrivateKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Private Key ID is set")
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		latest, err := conn.GetPrivateKey(&gofastly.GetPrivateKeyInput{
			ID: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		*pk = *latest

		return nil
	}
}

func testAccCheckPrivateKeyV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_private_key_v1" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		l, err := conn.ListPrivateKeys(&gofastly.ListPrivateKeysInput{})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing Private Keys when deleting Fastly Private Key (%s): %s", rs.Primary.ID, err)
		}

		for _, pk := range l {
			if pk.ID == rs.Primary.ID {
				// private key still found
				return fmt.Errorf("[WARN] Tried deleting Private Key (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccPrivateKeyV1Config(key, name string) string {
	return fmt.Sprintf(`
resource "fastly_private_key_v1" "foo" {
	key = "%s"
	name  = "%s"
}`, key, name)
}
