package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyMTLS_validate(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.example.com", name)

	_, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)

	resourceName := "fastly_tls_mutual_authentication.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMTLSConfig(cert, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "enforced", "false"),
				),
			},
			{
				Config: testAccMTLSConfig(cert, name+"updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"updated"),
					resource.TestCheckResourceAttr(resourceName, "enforced", "false"),
				),
			},
			{
				ResourceName:      "fastly_tls_mutual_authentication.test",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"cert_bundle"},
			},
		},
	})
}

func testAccMTLSConfig(cert, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_mutual_authentication" "test" {
  cert_bundle = <<EOF
%[1]s
EOF
  name = "%[2]s"
}
    `, cert, name)
}
