package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"strings"
	"testing"
)

func TestAccFastlyDataSourceTLSPrivateKeyBasic(t *testing.T) {
	key, _, err := generateKeyAndCert()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	key = strings.ReplaceAll(key, "\n", `\n`)

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSPrivateKeyConfig(key, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.fastly_tls_private_key.subject", "name", name),
					resource.TestCheckResourceAttr("data.fastly_tls_private_key.subject", "key_type", "RSA"),
				),
			},
		},
	})
}

func testAccFastlyDataSourceTLSPrivateKeyConfig(key, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "test" {
  key_pem = "%s"
  name = "%s"
}

data "fastly_tls_private_key" "subject" {
  name = fastly_tls_private_key.test.name
}
`, key, name)
}
