package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyDataSourceAuditLogEventMapping_byID(t *testing.T) {
	mappingName := fmt.Sprintf("Audit Log Event Mapping %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAuditLogEventMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAuditLogEventMappingConfigByID(mappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.fastly_audit_log_event_mapping.example", "id", "fastly_audit_log_event_mapping.example", "id"),
					resource.TestCheckResourceAttr("data.fastly_audit_log_event_mapping.example", "name", mappingName),
					resource.TestCheckResourceAttr("data.fastly_audit_log_event_mapping.example", "description", "Test audit log event mapping"),
					resource.TestCheckResourceAttr("data.fastly_audit_log_event_mapping.example", "scope_type", "account"),
					resource.TestCheckResourceAttr("data.fastly_audit_log_event_mapping.example", "event_types.#", "1"),
					resource.TestCheckResourceAttr("data.fastly_audit_log_event_mapping.example", "integration_ids.#", "1"),
					resource.TestCheckResourceAttr("data.fastly_audit_log_event_mapping.example", "mapping_status", "active"),
				),
			},
		},
	})
}

func TestAccFastlyDataSourceAuditLogEventMapping_byName(t *testing.T) {
	mappingName := fmt.Sprintf("Audit Log Event Mapping %s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAuditLogEventMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAuditLogEventMappingConfigByName(mappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.fastly_audit_log_event_mapping.example", "id", "fastly_audit_log_event_mapping.example", "id"),
					resource.TestCheckResourceAttr("data.fastly_audit_log_event_mapping.example", "scope_type", "account"),
				),
			},
		},
	})
}

func testAccFastlyDataSourceAuditLogEventMappingConfigByID(name string) string {
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
  event_types     = ["user.login"]
  integration_ids = [fastly_integration.foo.id]
}

data "fastly_audit_log_event_mapping" "example" {
  id = fastly_audit_log_event_mapping.example.id
}
`, name)
}

func testAccFastlyDataSourceAuditLogEventMappingConfigByName(name string) string {
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
  event_types     = ["user.login"]
  integration_ids = [fastly_integration.foo.id]
}

data "fastly_audit_log_event_mapping" "example" {
  name       = fastly_audit_log_event_mapping.example.name
  scope_type = "account"

  depends_on = [fastly_audit_log_event_mapping.example]
}
`, name)
}
