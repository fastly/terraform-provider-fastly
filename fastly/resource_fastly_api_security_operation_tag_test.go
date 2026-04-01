package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v14/fastly"
	"github.com/fastly/go-fastly/v14/fastly/apisecurity/operations"
)

func TestAccFastlyAPISecurityOperationTag_Basic(t *testing.T) {
	serviceName := fmt.Sprintf("tf-api-sec-tag-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	tagName := "tf-tag-" + acctest.RandString(8)
	desc1 := "example"
	desc2 := "example-updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAPISecurityOperationTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyAPISecurityOperationTagConfig(serviceName, domainName, tagName, desc1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"fastly_api_security_operation_tag.example", "service_id",
						"fastly_service_vcl.example", "id",
					),
					resource.TestCheckResourceAttr("fastly_api_security_operation_tag.example", "name", tagName),
					resource.TestCheckResourceAttr("fastly_api_security_operation_tag.example", "description", desc1),
					resource.TestCheckResourceAttrSet("fastly_api_security_operation_tag.example", "tag_id"),
				),
			},
			{
				Config: testAccFastlyAPISecurityOperationTagConfig(serviceName, domainName, tagName, desc2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_api_security_operation_tag.example", "description", desc2),
				),
			},
			{
				ResourceName:      "fastly_api_security_operation_tag.example",
				ImportState:       true,
				ImportStateVerify: true,
				// Import uses composite ID: <service_id>/<tag_id>
			},
		},
	})
}

func testAccFastlyAPISecurityOperationTagConfig(serviceName, domainName, tagName, desc string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "example" {
  name = "%s"

  domain {
    name = "%s"
  }

  force_destroy = true
}

resource "fastly_api_security_operation_tag" "example" {
  service_id  = fastly_service_vcl.example.id
  name        = "%s"
  description = "%s"
}
`, serviceName, domainName, tagName, desc)
}

func testAccCheckAPISecurityOperationTagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_api_security_operation_tag" {
			continue
		}

		// Resource ID is composite: service_id/tag_id
		serviceID, tagID, err := parseTwoPartImportID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = operations.DescribeTag(context.Background(), conn, &operations.DescribeTagInput{
			ServiceID: gofastly.ToPointer(serviceID),
			TagID:     gofastly.ToPointer(tagID),
		})
		if err == nil {
			return fmt.Errorf("API Security operation tag still exists after destroy: %s/%s", serviceID, tagID)
		}
		if httpErr, ok := err.(*gofastly.HTTPError); ok {
			if httpErr.IsNotFound() {
				continue
			}
		}
		return fmt.Errorf("error verifying API Security operation tag destroy (%s/%s): %w", serviceID, tagID, err)
	}

	return nil
}
