package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/go-fastly/v16/fastly/dns/v1/tsigkeys"
)

func TestAccFastlyTSIGKey_Basic(t *testing.T) {
	keyName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	create := tsigkeys.TSIGKey{
		Name:      gofastly.ToPointer(keyName),
		Algorithm: gofastly.ToPointer("hmac-sha256"),
	}
	update := tsigkeys.TSIGKey{
		Name:        gofastly.ToPointer(keyName),
		Algorithm:   gofastly.ToPointer("hmac-sha256"),
		Description: gofastly.ToPointer("updated description"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckTSIGKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTSIGKeyConfig(create, "dGVzdHNlY3JldA=="),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyTSIGKeyRemoteState("fastly_tsig_key.foo", create),
				),
			},
			{
				Config: testAccTSIGKeyConfigWithDescription(update, "dGVzdHNlY3JldA=="),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyTSIGKeyRemoteState("fastly_tsig_key.foo", update),
				),
			},
			{
				ResourceName:            "fastly_tsig_key.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccFastlyTSIGKey_Algorithms(t *testing.T) {
	type algoCase struct {
		algorithm string
		secret    string
	}
	cases := []algoCase{
		{"hmac-sha224", "c2VjcmV0Zm9yc2hhMjI0aGFzaGFsZ28="},
		{"hmac-sha256", "c2VjcmV0Zm9yc2hhMjU2aGFzaGFsZ28="},
		{"hmac-sha384", "c2VjcmV0Zm9yc2hhMzg0aGFzaGFsZ28="},
		{"hmac-sha512", "c2VjcmV0Zm9yc2hhNTEyaGFzaGFsZ28="},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.algorithm, func(t *testing.T) {
			keyName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
			key := tsigkeys.TSIGKey{
				Name:      gofastly.ToPointer(keyName),
				Algorithm: gofastly.ToPointer(tc.algorithm),
			}

			resource.ParallelTest(t, resource.TestCase{
				PreCheck: func() {
					testAccPreCheck(t)
				},
				ProviderFactories: testAccProviders,
				CheckDestroy:      testAccCheckTSIGKeyDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccTSIGKeyConfig(key, tc.secret),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckFastlyTSIGKeyRemoteState("fastly_tsig_key.foo", key),
						),
					},
				},
			})
		})
	}
}

func testAccCheckFastlyTSIGKeyRemoteState(resourceName string, expected tsigkeys.TSIGKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		got, err := tsigkeys.Get(context.TODO(), conn, &tsigkeys.GetInput{
			TSIGKeyID: gofastly.ToPointer(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("error fetching TSIG key (%s): %s", rs.Primary.ID, err)
		}

		if gofastly.ToValue(expected.Name) != gofastly.ToValue(got.Name) {
			return fmt.Errorf("bad name, expected (%s), got (%s)", gofastly.ToValue(expected.Name), gofastly.ToValue(got.Name))
		}
		if gofastly.ToValue(expected.Algorithm) != gofastly.ToValue(got.Algorithm) {
			return fmt.Errorf("bad algorithm, expected (%s), got (%s)", gofastly.ToValue(expected.Algorithm), gofastly.ToValue(got.Algorithm))
		}
		if expected.Description != nil {
			if gofastly.ToValue(expected.Description) != gofastly.ToValue(got.Description) {
				return fmt.Errorf("bad description, expected (%s), got (%s)", gofastly.ToValue(expected.Description), gofastly.ToValue(got.Description))
			}
		}

		return nil
	}
}

func testAccCheckTSIGKeyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_tsig_key" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		_, err := tsigkeys.Get(context.TODO(), conn, &tsigkeys.GetInput{
			TSIGKeyID: gofastly.ToPointer(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("tried deleting TSIG key (%s), but was still found", rs.Primary.ID)
		}
	}
	return nil
}

func testAccTSIGKeyConfig(key tsigkeys.TSIGKey, secret string) string {
	return fmt.Sprintf(`
resource "fastly_tsig_key" "foo" {
  name      = "%s"
  algorithm = "%s"
  secret    = "%s"
}`, gofastly.ToValue(key.Name), gofastly.ToValue(key.Algorithm), secret)
}

func testAccTSIGKeyConfigWithDescription(key tsigkeys.TSIGKey, secret string) string {
	return fmt.Sprintf(`
resource "fastly_tsig_key" "foo" {
  name        = "%s"
  algorithm   = "%s"
  secret      = "%s"
  description = "%s"
}`, gofastly.ToValue(key.Name), gofastly.ToValue(key.Algorithm), secret, gofastly.ToValue(key.Description))
}
