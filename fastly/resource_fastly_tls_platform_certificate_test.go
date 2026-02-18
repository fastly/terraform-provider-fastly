package fastly

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"

	"github.com/fastly/go-fastly/v13/fastly"
)

func init() {
	resource.AddTestSweepers("fastly_tls_platform_certificate", &resource.Sweeper{
		Name: "fastly_tls_platform_certificate",
		F:    testSweepTLSPlatformCertificates,
	})
}

func TestAccFastlyTLSPlatformCertificate_basic(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.test", name)

	key, cert, ca, err := generateKeyAndCertWithCA(domain)
	require.NoError(t, err)

	resourceName := "fastly_tls_platform_certificate.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckTLSPlatformCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTLSPlatformCertificateWithName(name, key, cert, ca),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "not_after"),
					resource.TestCheckResourceAttrSet(resourceName, "not_before"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "replace"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					testAccTLSPlatformCertificateExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_body", "intermediates_blob", "allow_untrusted_root"},
			},
		},
	})
}

func testAccCheckTLSPlatformCertificateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, r := range s.RootModule().Resources {
		if r.Type != "fastly_tls_platform_certificate" {
			continue
		}

		certificates, err := conn.ListBulkCertificates(context.TODO(), &fastly.ListBulkCertificatesInput{})
		if err != nil {
			return err
		}

		for _, certificate := range certificates {
			if certificate.ID == r.Primary.ID {
				return fmt.Errorf("certificate %s still exists", r.Primary.ID)
			}
		}
	}

	return nil
}

func testAccTLSPlatformCertificateExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		_, err := conn.GetBulkCertificate(context.TODO(), &fastly.GetBulkCertificateInput{
			ID: r.Primary.ID,
		})
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccTLSPlatformCertificateWithName(name string, key string, cert string, ca string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  name = "%[1]s"
  key_pem = <<EOF
%[2]s
EOF
}

data "fastly_tls_configuration" "config" {
  tls_service = "PLATFORM"
}

resource "fastly_tls_platform_certificate" "subject" {
  certificate_body = <<EOF
%[3]s
EOF
  intermediates_blob = <<EOF
%[4]s
EOF
  configuration_id = data.fastly_tls_configuration.config.id
  allow_untrusted_root = true
  depends_on = [fastly_tls_private_key.key]
}
`, name, key, cert, ca)
}

func testSweepTLSPlatformCertificates(region string) error {
	client, diagnostics := sharedClientForRegion(region)
	if diagnostics.HasError() {
		return diagToErr(diagnostics)
	}

	certificates, err := client.ListBulkCertificates(context.TODO(), &fastly.ListBulkCertificatesInput{PageSize: 1000})
	if err != nil {
		return err
	}

	for _, certificate := range certificates {
		for _, domain := range certificate.Domains {
			// SA4017 ignoring because HasPrefix returned value IS being used.
			//nolint: staticcheck
			if !strings.HasPrefix(domain.ID, testResourcePrefix) || !strings.HasPrefix(domain.ID, ".test") {
				continue
			}
		}

		err := client.DeleteBulkCertificate(context.TODO(), &fastly.DeleteBulkCertificateInput{ID: certificate.ID})
		if err != nil {
			return err
		}
	}

	return nil
}
