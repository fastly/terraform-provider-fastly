package fastly

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

func resourceFastlyNGWAFWorkspaceRule() *schema.Resource {
	r := resourceFastlyNGWAFRuleBase()

	r.Importer = customNGWAFScopeImporter(scope.ScopeTypeWorkspace, "rule")

	r.Schema["type"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Description:      "The type of the rule. Accepted values are `request`, `signal`, `rate_limit`, and `templated_signal`.",
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"request", "signal", "templated_signal", "rate_limit"}, false)),
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
		Description: "The ID of the workspace.",
	}

	r.Schema["rate_limit"] = &schema.Schema{
		Description: "Block specifically for rate_limit rules.",
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"client_identifiers": {
					Description: "List of client identifiers used for rate limiting. Can only be length 1 or 2.",
					Type:        schema.TypeSet,
					Required:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"key": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Key for the Client Identifier.",
							},
							"name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Name for the Client Identifier.",
							},
							"type": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Type of the Client Identifier. Accepted values are `ip`, `post_parameter`, `request_cookie`, `request_header`, and `signal_payload`.",
							},
						},
					},
				},
				"duration": {
					Type:        schema.TypeInt,
					Required:    true,
					Description: "Duration in seconds for the rate limit.",
				},
				"interval": {
					Type:         schema.TypeInt,
					Required:     true,
					Description:  "Time interval for the rate limit in seconds. Accepted values are 60, 600, and 3600.",
					ValidateFunc: validation.IntInSlice([]int{60, 600, 3600}),
				},
				"signal": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Reference ID of the custom signal this rule uses to count requests.",
				},
				"threshold": {
					Type:         schema.TypeInt,
					Required:     true,
					Description:  "Rate limit threshold. Minimum 1 and maximum 10,000.",
					ValidateFunc: validation.IntBetween(1, 10000),
				},
			},
		},
	}

	return r
}

func resourceFastlyNGWAFAccountRule() *schema.Resource {
	r := resourceFastlyNGWAFRuleBase()

	r.Importer = customNGWAFScopeImporter(scope.ScopeTypeAccount, "rule")

	r.Schema["applies_to"] = &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: "The list of workspace IDs this signal applies to, or the wildcard `*` if it applies to all workspaces.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	// Remove fields that are invalid for account-scoped actions
	actions := r.Schema["action"].Elem.(*schema.Resource)
	delete(actions.Schema, "allow_interactive")
	delete(actions.Schema, "deception_type")
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
