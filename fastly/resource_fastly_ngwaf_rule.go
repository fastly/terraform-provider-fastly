package fastly

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/common"
)

func resourceFastlyNGWAFWorkspaceRule() *schema.Resource {
	r := resourceFastlyNGWAFRuleBase()

	r.Importer = customNGWAFScopeImporter(common.ScopeTypeWorkspace, "rule")

	r.Schema["type"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Description:      "The type of the rule (`request`, `signal`, or `templated_signal`).",
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"request", "signal", "templated_signal"}, false)),
	}

	// Force recreation for templated_signal rules to avoid "templateSignal rules expect no actions"
	// API error
	r.CustomizeDiff = customdiff.All(
		// Validate description for templated_signal rules
		func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
			if diff.Get("type").(string) == "templated_signal" {
				if desc := diff.Get("description").(string); desc != "" {
					return fmt.Errorf("description must be an empty string for templated_signal rules")
				}
			}
			return nil
		},
		// Force recreation when specific fields change on templated_signal rules
		forceNewOnTemplatedSignalChange(r.Schema),
	)

	r.Schema["workspace_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Next-Gen WAF workspace this rule belongs to.",
	}

	return r
}

func resourceFastlyNGWAFAccountRule() *schema.Resource {
	r := resourceFastlyNGWAFRuleBase()

	r.Importer = customNGWAFScopeImporter(common.ScopeTypeAccount, "rule")

	r.Schema["applies_to"] = &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: "The list of workspace IDs or wildcard `*` this account-level rule applies to.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	// Remove fields that are invalid for account-scoped actions
	actions := r.Schema["action"].Elem.(*schema.Resource)
	delete(actions.Schema, "redirect_url")
	delete(actions.Schema, "response_code")

	return r
}

// forceNewOnTemplatedSignalChange returns a CustomizeDiffFunc that forces a new
// resource to be created if the rule type is templated_signal and any of the
// fields in the schema have changed.
func forceNewOnTemplatedSignalChange(s map[string]*schema.Schema) schema.CustomizeDiffFunc {
	return func(_ context.Context, d *schema.ResourceDiff, _ any) error {
		if d.Get("type").(string) == "templated_signal" {
			for k := range s {
				if d.HasChange(k) {
					if err := d.ForceNew(k); err != nil {
						return fmt.Errorf("error setting force new for %s: %w", k, err)
					}
				}
			}
		}
		return nil
	}
}
