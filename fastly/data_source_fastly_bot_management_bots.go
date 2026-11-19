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

func dataSourceFastlyBotManagementBots() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to list the predefined bots in a Fastly Bot Management workspace, optionally limited to one category.",
		ReadContext: dataSourceFastlyBotManagementBotsRead,
		Schema: map[string]*schema.Schema{
			"bots": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The bots in the Bot Management workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The bot's action. One of `allow`, `block`, `challenge`, or `inherit` (no override; follows the category's action).",
						},
						"bot_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the bot.",
						},
						"category_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the category the bot belongs to.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The display name of the bot.",
						},
					},
				},
			},
			"category_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Only list the bots in this category. When omitted, bots in every category are listed.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Bot Management workspace.",
			},
		},
	}
}

func dataSourceFastlyBotManagementBotsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	workspaceID := d.Get("workspace_id").(string)
	categoryID := d.Get("category_id").(string)

	log.Printf("[DEBUG] Reading Bot Management bots for workspace %s, category %q", workspaceID, categoryID)

	input := &policy.ListBotsInput{WorkspaceID: new(workspaceID)}
	if categoryID != "" {
		input.CategoryID = new(categoryID)
	}

	bots, err := policy.ListBots(ctx, conn, input)
	if err != nil {
		return diag.Errorf("error listing Bot Management bots: %s", err)
	}

	ids := []string{workspaceID, categoryID}
	result := make([]map[string]any, 0, len(bots))
	for _, b := range bots {
		// The per-category endpoint omits category_id from each bot.
		botCategoryID := b.CategoryID
		if botCategoryID == "" {
			botCategoryID = categoryID
		}
		ids = append(ids, botCategoryID+"/"+b.BotID)
		result = append(result, map[string]any{
			"bot_id":      b.BotID,
			"category_id": botCategoryID,
			"name":        b.Name,
			"action":      b.Action.Type,
		})
	}

	d.SetId(strconv.Itoa(hashcode.String(strings.Join(ids, ","))))

	if err := d.Set("bots", result); err != nil {
		return diag.Errorf("error setting bots: %s", err)
	}

	return nil
}
