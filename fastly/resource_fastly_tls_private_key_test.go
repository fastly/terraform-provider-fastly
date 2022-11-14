package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("fastly_tls_private_key", &resource.Sweeper{
		Name:         "fastly_tls_private_key",
		Dependencies: []string{"fastly_tls_certificate"}, // in case a private key is used by a certificate
		F:            testSweepTLSPrivateKeys,
	})
}

func TestAccFastlyResourceTLSPrivateKey_basic(t *testing.T) {
	key, _, err := generateKeyAndCert()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	key = strings.ReplaceAll(key, "\n", `\n`)

	name := acctest.RandomWithPrefix(testResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckPrivateKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyTLSPrivateKeyConfigSimplePrivateKey(key, name),
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

		conn := testAccProvider.Meta().(*APIClient).conn
		keys, err := conn.ListPrivateKeys(&fastly.ListPrivateKeysInput{})
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

func testAccFastlyTLSPrivateKeyConfigSimplePrivateKey(key, name string) string {
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

		conn := testAccProvider.Meta().(*APIClient).conn

		_, err := conn.GetPrivateKey(&fastly.GetPrivateKeyInput{
			ID: res.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("error getting private key from Fastly: %w", err)
		}

		return nil
	}
}

func testSweepTLSPrivateKeys(region string) error {
	client, diagnostics := sharedClientForRegion(region)
	if diagnostics.HasError() {
		return diagToErr(diagnostics)
	}

	keys, err := client.ListPrivateKeys(&fastly.ListPrivateKeysInput{PageSize: 1000})
	if err != nil {
		return err
	}

	for _, key := range keys {
		if !strings.HasPrefix(key.Name, testResourcePrefix) {
			continue
		}

		err := client.DeletePrivateKey(&fastly.DeletePrivateKeyInput{ID: key.ID})
		if err != nil {
			return err
		}
	}

	return nil
}
