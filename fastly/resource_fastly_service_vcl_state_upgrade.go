package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// upgradeServiceVCLStateV0toV1 upgrades the state schema from version 0 to version 1.
// This handles the breaking change in v9.0.0 where bot_management was changed from
// a boolean to a list block with nested enabled and contentguard attributes.
func upgradeServiceVCLStateV0toV1(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	if rawState == nil {
		return rawState, nil
	}

	log.Println("[DEBUG] Upgrading fastly_service_vcl state from v0 to v1")

	// Check if product_enablement block exists
	// In rawState, Sets and Lists are both represented as []any
	productEnablementRaw, ok := rawState["product_enablement"]
	if !ok {
		return rawState, nil
	}

	// Product_enablement is a Set in the schema, but in rawState it's a []any
	productEnablementList, ok := productEnablementRaw.([]any)
	if !ok {
		log.Printf("[DEBUG] product_enablement has unexpected type: %T", productEnablementRaw)
		return rawState, nil
	}

	if len(productEnablementList) == 0 {
		return rawState, nil
	}

	// Get the product_enablement block
	productEnablement, ok := productEnablementList[0].(map[string]any)
	if !ok {
		log.Printf("[DEBUG] product_enablement element has unexpected type: %T", productEnablementList[0])
		return rawState, nil
	}

	// Check if bot_management exists and needs migration
	botManagementRaw, exists := productEnablement["bot_management"]
	if !exists {
		return rawState, nil
	}

	switch botManagement := botManagementRaw.(type) {
	case bool:
		log.Printf("[DEBUG] Migrating bot_management from bool (%v) to list block", botManagement)

		// Convert boolean to new list structure
		// If bot_management was true, set enabled=true and contentguard="off" (default)
		// If bot_management was false, remove the block entirely (empty list)
		if botManagement {
			productEnablement["bot_management"] = []any{
				map[string]any{
					"enabled":      true,
					"contentguard": "off",
				},
			}
		} else {
			// If bot_management was false, set it to an empty list
			productEnablement["bot_management"] = []any{}
		}

	case []any:
		// Already in new format, no migration needed
		log.Println("[DEBUG] bot_management already in list format, skipping migration")

	default:
		log.Printf("[DEBUG] Unexpected bot_management type: %T", botManagement)
	}

	return rawState, nil
}

// upgradeServiceVCLStateV1toV2 upgrades the state schema from version 1 to version 2,
// where the snippet block moved from TypeSet to TypeList. Both serialize as a JSON
// array, so the raw state passes through unchanged.
func upgradeServiceVCLStateV1toV2(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	if rawState == nil {
		return rawState, nil
	}

	log.Println("[DEBUG] Upgrading fastly_service_vcl state from v1 to v2 (snippet TypeSet -> TypeList)")

	return rawState, nil
}

// serviceVCLStateUpgraderV1 returns the parts of the version 1 schema of the service
// VCL resource that are relevant to the v1 -> v2 upgrade, where snippet was a TypeSet.
func serviceVCLStateUpgraderV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"snippet": {
				Type:        schema.TypeSet,
				Description: "VCL snippets (v1 schema, set container)",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content":  {Type: schema.TypeString, Description: "Snippet content", Required: true},
						"name":     {Type: schema.TypeString, Description: "Snippet name", Required: true},
						"priority": {Type: schema.TypeInt, Description: "Snippet priority", Optional: true},
						"type":     {Type: schema.TypeString, Description: "Snippet type", Required: true},
					},
				},
			},
		},
	}
}

// serviceVCLStateUpgraderV0 returns the schema for version 0 of the service VCL resource.
// This represents the schema before the bot_management change in v9.0.0.
func serviceVCLStateUpgraderV0() *schema.Resource {
	// Return a resource with the old schema (v0) where bot_management was a boolean
	// We only need to define the parts of the schema that are relevant to the upgrade
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"product_enablement": {
				Type:        schema.TypeList,
				Description: "Product Enablement",
				Optional:    true,
				MaxItems:    1,
				MinItems:    1,
				Elem: &schema.Resource{
					Description: "Product Enablement values",
					Schema: map[string]*schema.Schema{
						"bot_management": {
							Description: "Bot management enablement",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						// Other fields are omitted as they don't affect the upgrade
					},
				},
			},
		},
	}
}
