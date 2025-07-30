package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/common"
)

func resourceFastlyNGWAFWorkspaceRule() *schema.Resource {
	r := resourceFastlyNGWAFRuleBase()

	r.Importer = customNGWAFRuleImporter(common.ScopeTypeWorkspace)

	r.Schema["workspace_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Next-Gen WAF workspace this rule belongs to.",
	}

	return r
}

func resourceFastlyNGWAFAccountRule() *schema.Resource {
	r := resourceFastlyNGWAFRuleBase()

	r.Importer = customNGWAFRuleImporter(common.ScopeTypeAccount)

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
