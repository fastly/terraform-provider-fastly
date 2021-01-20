package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strings"
	"testing"
)

func TestAccFastlyResourceTLSPrivateKey_basic(t *testing.T) {
	key, _, err := generateKeyAndCert()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	key = strings.ReplaceAll(key, "\n", `\n`)

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPrivateKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyTLSPrivateKeyConfig_simple_private_key(key, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateKeyExists("fastly_tls_private_key.foo"),
					resource.TestCheckResourceAttr("fastly_tls_private_key.foo", "name", name),
				),
			},
			{
				ResourceName:            "fastly_tls_private_key.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key_pem"},
			},
		},
	})
}

func testAccCheckPrivateKeyDestroy(state *terraform.State) error {
	for _, resourceState := range state.RootModule().Resources {
		if resourceState.Type != "fastly_tls_private_key" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		keys, err := conn.ListPrivateKeys(&gofastly.ListPrivateKeysInput{})
		if err != nil {
			return fmt.Errorf(
				"[WARN] Error listing private keys when deleting private key %s: %w",
				resourceState.Primary.ID,
				err,
			)
		}

		for _, key := range keys {
			if key.ID == resourceState.Primary.ID {
				return fmt.Errorf(
					"[WARN] Tried deleting private key (%s) but was still found",
					resourceState.Primary.ID,
				)
			}
		}
	}
	return nil
}

func testAccFastlyTLSPrivateKeyConfig_simple_private_key(key, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "foo" {
  key_pem = "%s"
  name    = "%s"
}`, key, name)
}

func testAccCheckPrivateKeyExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if res.Primary.ID == "" {
			return fmt.Errorf("no id set on resource %s", resourceName)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn

		_, err := conn.GetPrivateKey(&gofastly.GetPrivateKeyInput{
			ID: res.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("error getting private key from Fastly: %w", err)
		}

		return nil
	}
}
