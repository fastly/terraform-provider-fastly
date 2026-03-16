package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"
)

func TestAccFastlyAPISecurityOperation_Basic(t *testing.T) {
	serviceName := fmt.Sprintf("tf-api-sec-op-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	// Operation identity (ForceNew fields) must remain stable across updates.
	method := "GET"
	opDomain := domainName
	opPath := "/tf-test"

	desc1 := "example"
	desc2 := "example-updated"
	tagName := "tf-tag-" + acctest.RandString(6)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAPISecurityOperationDestroy,
		Steps: []resource.TestStep{
			{
				// Create op with description only (no tags)
				Config: testAccFastlyAPISecurityOperationConfig(serviceName, domainName, method, opDomain, opPath, desc1, false, tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"fastly_api_security_operation.example", "service_id",
						"fastly_service_vcl.example", "id",
					),
					resource.TestCheckResourceAttr("fastly_api_security_operation.example", "method", method),
					resource.TestCheckResourceAttr("fastly_api_security_operation.example", "domain", opDomain),
					resource.TestCheckResourceAttr("fastly_api_security_operation.example", "path", opPath),
					resource.TestCheckResourceAttr("fastly_api_security_operation.example", "description", desc1),
					resource.TestCheckResourceAttrSet("fastly_api_security_operation.example", "operation_id"),
				),
			},
			{
				// Update op: change description + set tag_ids by creating a tag
				Config: testAccFastlyAPISecurityOperationConfig(serviceName, domainName, method, opDomain, opPath, desc2, true, tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_api_security_operation.example", "description", desc2),
					resource.TestCheckResourceAttr("fastly_api_security_operation.example", "tag_ids.#", "1"),
				),
			},
			{
				ResourceName:      "fastly_api_security_operation.example",
				ImportState:       true,
				ImportStateVerify: true,
				// Import uses composite ID: <service_id>/<operation_id>
			},
		},
	})
}

func testAccFastlyAPISecurityOperationConfig(
	serviceName, domainName, method, opDomain, opPath, description string,
	withTag bool,
	tagName string,
) string {
	tagBlock := ""
	tagIDs := "[]"
	if withTag {
		tagBlock = fmt.Sprintf(`
resource "fastly_api_security_operation_tag" "tag" {
  service_id   = fastly_service_vcl.example.id
  name         = "%s"
  description  = "tag for operation test"
}
`, tagName)
		tagIDs = "[fastly_api_security_operation_tag.tag.tag_id]"
	}

	return fmt.Sprintf(`
resource "fastly_service_vcl" "example" {
  name = "%s"

  domain {
    name = "%s"
  }

  force_destroy = true
}

%s

resource "fastly_api_security_operation" "example" {
  service_id   = fastly_service_vcl.example.id
  method       = "%s"
  domain       = "%s"
  path         = "%s"
  description  = "%s"
  tag_ids      = %s
}
`, serviceName, domainName, tagBlock, method, opDomain, opPath, description, tagIDs)
}

func testAccCheckAPISecurityOperationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_api_security_operation" {
			continue
		}

		// Resource ID is composite: service_id/operation_id
		serviceID, opID, err := parseTwoPartImportID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = operations.Describe(context.Background(), conn, &operations.DescribeInput{
			ServiceID:   gofastly.ToPointer(serviceID),
			OperationID: gofastly.ToPointer(opID),
		})
		if err == nil {
			return fmt.Errorf("API Security operation still exists after destroy: %s/%s", serviceID, opID)
		}
		if httpErr, ok := err.(*gofastly.HTTPError); ok {
			if httpErr.IsNotFound() {
				continue
			}
		}
		return fmt.Errorf("error verifying API Security operation destroy (%s/%s): %w", serviceID, opID, err)
	}

	return nil
}
