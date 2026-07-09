package acceptancetests

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly/computeacls"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyACL_basic(t *testing.T) {
	t.Parallel()
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("tf_test_acl_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: ConfigACL(aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_acl.acl", "name", aclName),
					resource.TestCheckResourceAttrSet("fastly_acl.acl", "id"),
				),
			},
			{
				// Changing the name forces replacement.
				Config: ConfigACL(aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_acl.acl", "name", aclNameUpdated),
					resource.TestCheckResourceAttrSet("fastly_acl.acl", "id"),
				),
			},
			{
				ResourceName:      "fastly_acl.acl",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// CheckServiceAndACLDestroy composes CheckServiceDestroy for the given service resource
// type with CheckACLDestroy, for tests that manage both a service and one or more
// fastly_acl resources.
func CheckServiceAndACLDestroy(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if err := CheckServiceDestroy(resourceType)(s); err != nil {
			return err
		}
		return CheckACLDestroy(s)
	}
}

func CheckACLDestroy(s *terraform.State) error {
	client, err := NewFastlyClient()
	if err != nil {
		return fmt.Errorf("error creating Fastly client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_acl" {
			continue
		}

		id := rs.Primary.ID
		_, err := computeacls.Describe(context.Background(), client, &computeacls.DescribeInput{ComputeACLID: &id})
		if errors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("error checking if ACL was destroyed: %w", err)
		}

		return fmt.Errorf("ACL %s still exists", id)
	}

	return nil
}

func ConfigACL(name string) string {
	return fmt.Sprintf(`
resource "fastly_acl" "acl" {
  name = %q
}
`, name)
}

func TestAccFastlyDataSourceACLs(t *testing.T) {
	t.Parallel()
	h := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: ConfigACLsDataSource(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["data.fastly_acls.example"]
						if !ok {
							return fmt.Errorf("not found: data.fastly_acls.example")
						}

						want := []string{
							fmt.Sprintf("tf_%s_1", h),
							fmt.Sprintf("tf_%s_2", h),
							fmt.Sprintf("tf_%s_3", h),
						}

						var found int
						var got []string
						for k, v := range rs.Primary.Attributes {
							if strings.HasSuffix(k, ".name") {
								got = append(got, v)
								if slices.Contains(want, v) {
									found++
								}
							}
						}

						if found != len(want) {
							return fmt.Errorf("want: %v, got: %v", want, got)
						}

						return nil
					},
				),
			},
		},
	})
}

func ConfigACLsDataSource(h string) string {
	return fmt.Sprintf(`
resource "fastly_acl" "acl_1" {
  name = "tf_%s_1"
}

resource "fastly_acl" "acl_2" {
  name = "tf_%s_2"
}

resource "fastly_acl" "acl_3" {
  name = "tf_%s_3"
}

data "fastly_acls" "example" {
  depends_on = [
    fastly_acl.acl_1,
    fastly_acl.acl_2,
    fastly_acl.acl_3,
  ]
}
`, h, h, h)
}
