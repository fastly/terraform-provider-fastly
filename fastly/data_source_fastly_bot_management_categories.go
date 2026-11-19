package fastly

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	"github.com/fastly/go-fastly/v19/fastly/botmanagement/v1/workspaces/policy"
)

func dataSourceFastlyBotManagementCategories() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to list the predefined bot categories in a Fastly Bot Management workspace.",
		ReadContext: dataSourceFastlyBotManagementCategoriesRead,
		Schema: map[string]*schema.Schema{
			"categories": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The bot categories in the Bot Management workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The category's action. One of `allow`, `block`, or `challenge`.",
						},
						"category_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the category.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A description of the bots the category contains.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The display name of the category.",
						},
					},
				},
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Bot Management workspace.",
			},
		},
	}
}

func dataSourceFastlyBotManagementCategoriesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading Bot Management categories for workspace %s", workspaceID)

	categories, err := policy.ListCategories(ctx, conn, &policy.ListCategoriesInput{WorkspaceID: new(workspaceID)})
	if err != nil {
		return diag.Errorf("error listing Bot Management categories: %s", err)
	}

	ids := []string{workspaceID}
	result := make([]map[string]any, 0, len(categories))
	for _, c := range categories {
		ids = append(ids, c.CategoryID)
		result = append(result, map[string]any{
			"category_id": c.CategoryID,
			"name":        c.Name,
			"description": c.Description,
			"action":      c.Action.Type,
		})
	}

	d.SetId(strconv.Itoa(hashcode.String(strings.Join(ids, ","))))

	if err := d.Set("categories", result); err != nil {
		return diag.Errorf("error setting categories: %s", err)
	}

	return nil
}
