package fastly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"

	"github.com/fastly/go-fastly/v17/fastly"
)

func TestResourceFastlyTLSCertificateAllowUntrustedRoot(t *testing.T) {
	testCases := []struct {
		name   string
		method string
		allow  bool
	}{
		{name: "create with untrusted root", method: http.MethodPost, allow: true},
		{name: "create with trusted root", method: http.MethodPost, allow: false},
		{name: "update with untrusted root", method: http.MethodPatch, allow: true},
		{name: "update with trusted root", method: http.MethodPatch, allow: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			type observedAttribute struct {
				present bool
				value   bool
			}

			observed := make(chan observedAttribute, 1)
			certificateResponse := `{
				"data": {
					"type": "tls_certificate",
					"id": "test-certificate",
					"attributes": {
						"created_at": "2026-09-05T00:00:00Z",
						"issued_to": "example.com",
						"issuer": "Test CA",
						"name": "test-certificate",
						"replace": false,
						"serial_number": "1",
						"signature_algorithm": "SHA256-RSA",
						"updated_at": "2026-09-05T00:00:00Z"
					},
					"relationships": {
						"tls_domains": {"data": []}
					}
				}
			}`

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/vnd.api+json")
				if r.Method == testCase.method {
					var payload struct {
						Data struct {
							Attributes map[string]any `json:"attributes"`
						} `json:"data"`
					}
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}

					value, present := payload.Data.Attributes["allow_untrusted_root"]
					boolValue, _ := value.(bool)
					observed <- observedAttribute{present: present, value: boolValue}
				}
				_, _ = w.Write([]byte(certificateResponse))
			}))
			defer server.Close()

			conn, err := fastly.NewClientForEndpoint("test-key", server.URL)
			require.NoError(t, err)
			client := &APIClient{conn: conn}
			d := schema.TestResourceDataRaw(t, resourceFastlyTLSCertificate().Schema, map[string]any{
				"allow_untrusted_root": testCase.allow,
				"certificate_body":     "test-certificate-body",
				"name":                 "test-certificate",
			})

			var diagnostics diag.Diagnostics
			if testCase.method == http.MethodPost {
				diagnostics = resourceFastlyTLSCertificateCreate(context.Background(), d, client)
			} else {
				d.SetId("test-certificate")
				diagnostics = resourceFastlyTLSCertificateUpdate(context.Background(), d, client)
			}
			require.False(t, diagnostics.HasError(), diagnostics)

			attribute := <-observed
			if testCase.allow {
				require.True(t, attribute.present)
				require.True(t, attribute.value)
			} else {
				require.False(t, attribute.present)
			}
		})
	}
}

func init() {
	resource.AddTestSweepers("fastly_tls_certificate", &resource.Sweeper{
		Name:         "fastly_tls_certificate",
		Dependencies: []string{"fastly_tls_activation"}, // in case certificate used by an activation
		F:            testSweepTLSCertificates,
	})
}

func TestAccFastlyTLSCertificate_withName(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	updatedName := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.example.com", name)

	key, cert, cert2, err := generateKeyAndMultipleCerts(domain)
	require.NoError(t, err)

	resourceName := "fastly_tls_certificate.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckTLSCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTLSCertificateWithName(name, key, name, cert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "allow_untrusted_root", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "issued_to", domain),
					resource.TestCheckResourceAttrSet(resourceName, "issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "replace"),
					resource.TestCheckResourceAttrSet(resourceName, "serial_number"),
					resource.TestCheckResourceAttrSet(resourceName, "signature_algorithm"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					testAccTLSCertificateExists(resourceName),
				),
			},
			{
				Config: testAccTLSCertificateWithName(name, key, updatedName, cert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "allow_untrusted_root", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_untrusted_root", "certificate_body"},
			},
		},
	})
}

func TestAccFastlyTLSCertificate_withoutName(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.example.com", name)

	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)

	resourceName := "fastly_tls_certificate.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckTLSCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTLSCertificateWithoutName(name, key, cert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "allow_untrusted_root", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", domain),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "issued_to", domain),
					resource.TestCheckResourceAttrSet(resourceName, "issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "replace"),
					resource.TestCheckResourceAttrSet(resourceName, "serial_number"),
					resource.TestCheckResourceAttrSet(resourceName, "signature_algorithm"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					testAccTLSCertificateExists(resourceName),
				),
			},
		},
	})
}

func testAccTLSCertificateExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		_, err := conn.GetCustomTLSCertificate(context.TODO(), &fastly.GetCustomTLSCertificateInput{
			ID: r.Primary.ID,
		})
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccTLSCertificateWithName(keyName string, key string, certName string, cert string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  name = "%[1]s"
  key_pem = <<EOF
%[2]s
EOF
}

resource "fastly_tls_certificate" "test" {
  allow_untrusted_root = true
  name = "%[3]s"
  certificate_body = <<EOF
%[4]s
EOF
  depends_on = [fastly_tls_private_key.key]
}
`, keyName, key, certName, cert)
}

func testAccTLSCertificateWithoutName(keyName string, key string, cert string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  name = "%[1]s"
  key_pem = <<EOF
%[2]s
EOF
}

resource "fastly_tls_certificate" "test" {
  allow_untrusted_root = true
  certificate_body = <<EOF
%[3]s
EOF
  depends_on = [fastly_tls_private_key.key]
}
`, keyName, key, cert)
}

func testAccCheckTLSCertificateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, r := range s.RootModule().Resources {
		if r.Type != "fastly_tls_certificate" {
			continue
		}

		certificates, err := conn.ListCustomTLSCertificates(context.TODO(), &fastly.ListCustomTLSCertificatesInput{})
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

func testSweepTLSCertificates(region string) error {
	client, diagnostics := sharedClientForRegion(region)
	if diagnostics.HasError() {
		return diagToErr(diagnostics)
	}

	certificates, err := client.ListCustomTLSCertificates(context.TODO(), &fastly.ListCustomTLSCertificatesInput{PageSize: 1000})
	if err != nil {
		return err
	}

	for _, certificate := range certificates {
		if !strings.HasPrefix(certificate.Name, testResourcePrefix) {
			continue
		}

		err := client.DeleteCustomTLSCertificate(context.TODO(), &fastly.DeleteCustomTLSCertificateInput{ID: certificate.ID})
		if err != nil {
			return err
		}
	}

	return nil
}
