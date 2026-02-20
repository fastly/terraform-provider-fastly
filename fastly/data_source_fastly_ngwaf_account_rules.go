package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"
)

func dataSourceFastlyNGWAFAccountRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFAccountRulesRead,
		Schema: map[string]*schema.Schema{
			"rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time in ISO 8601 format when the rule was created.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the rule.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the rule is currently enabled.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the rule.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the rule.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time in ISO 8601 format when the rule was last updated.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyNGWAFAccountRulesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading NGWAF account rules")

	scopeObj := &scope.Scope{
		Type:      scope.ScopeTypeAccount,
		AppliesTo: []string{"*"},
	}

	remoteState, err := rules.List(ctx, conn, &rules.ListInput{
		Scope: scopeObj,
	})
	if err != nil {
		return diag.Errorf("error fetching rules: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	var rulePtrs []*rules.Rule
	for i := range remoteState.Data {
		rulePtrs = append(rulePtrs, &remoteState.Data[i])
	}

	if err := d.Set("rules", flattenNGWAFRules(rulePtrs)); err != nil {
		return diag.Errorf("error setting rules: %s", err)
	}

	return nil
}

func flattenNGWAFRules(remoteState []*rules.Rule) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, rule := range remoteState {
		result[i] = map[string]any{
			"created_at":  rule.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"description": rule.Description,
			"enabled":     rule.Enabled,
			"id":          rule.RuleID,
			"type":        rule.Type,
			"updated_at":  rule.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return result
}
