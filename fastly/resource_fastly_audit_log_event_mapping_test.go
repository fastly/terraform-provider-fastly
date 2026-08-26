package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	eventmappings "github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"
)

func TestAccFastlyAuditLogEventMapping_validate(t *testing.T) {
	mappingName := fmt.Sprintf("Audit Log Event Mapping %s", acctest.RandString(10))
	newMappingName := fmt.Sprintf("Audit Log Event Mapping %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAuditLogEventMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAuditLogEventMappingConfig(mappingName, "user.login"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "name", mappingName),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "description", "Test audit log event mapping"),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "scope_type", eventmappings.ScopeTypeAccount),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "event_types.#", "1"),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "integration_ids.#", "1"),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "mapping_status", eventmappings.MappingStatusActive),
					testAccAuditLogEventMappingExists("fastly_audit_log_event_mapping.example"),
				),
			},
			{
				Config: testAccAuditLogEventMappingConfigUpdate(newMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "name", newMappingName),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "description", "Test audit log event mapping updated"),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "scope_type", eventmappings.ScopeTypeAccount),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "event_types.#", "2"),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "integration_ids.#", "1"),
					resource.TestCheckResourceAttr("fastly_audit_log_event_mapping.example", "mapping_status", eventmappings.MappingStatusActive),
					testAccAuditLogEventMappingExists("fastly_audit_log_event_mapping.example"),
				),
			},
			{
				ResourceName:      "fastly_audit_log_event_mapping.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAuditLogEventMappingExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		mapping, err := eventmappings.Get(context.TODO(), conn, &eventmappings.GetInput{
			MappingID: &rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("unable to retrieve audit log event mapping %s: %v", rs.Primary.ID, err)
		}

		if mapping == nil {
			return fmt.Errorf("audit log event mapping %s not found in API", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAuditLogEventMappingDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_audit_log_event_mapping" {
			continue
		}

		_, err := eventmappings.Get(context.TODO(), conn, &eventmappings.GetInput{
			MappingID: &rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("audit log event mapping %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccAuditLogEventMappingConfig(name, eventType string) string {
	return fmt.Sprintf(`
resource "fastly_integration" "foo" {
  name        = "%[1]s integration"
  description = "Test integration for audit log event mapping"
  type        = "webhook"

  config = {
    webhook = "https://example.com/webhook"
  }
}

resource "fastly_audit_log_event_mapping" "example" {
  name            = "%[1]s"
  description     = "Test audit log event mapping"
  scope_type      = "account"
  event_types     = ["%[2]s"]
  integration_ids = [fastly_integration.foo.id]
}
`, name, eventType)
}

func testAccAuditLogEventMappingConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "fastly_integration" "foo" {
  name        = "%[1]s integration"
  description = "Test integration for audit log event mapping"
  type        = "webhook"

  config = {
    webhook = "https://example.com/webhook"
  }
}

resource "fastly_audit_log_event_mapping" "example" {
  name            = "%[1]s"
  description     = "Test audit log event mapping updated"
  scope_type      = "account"
  event_types     = ["user.login", "user.create"]
  integration_ids = [fastly_integration.foo.id]
}
`, name)
}
