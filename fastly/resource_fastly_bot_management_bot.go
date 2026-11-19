package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/botmanagement/v1/workspaces/policy"
)

var botManagementBotActions = []string{"allow", "block", "challenge", "inherit"}

func resourceFastlyBotManagementBot() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the action of a predefined bot in a Fastly Bot Management workspace. This resource doesn't create or delete bots, and destroying the resource only removes it from Terraform state.",
		CreateContext: resourceFastlyBotManagementBotCreate,
		ReadContext:   resourceFastlyBotManagementBotRead,
		UpdateContext: resourceFastlyBotManagementBotUpdate,
		DeleteContext: resourceFastlyBotManagementBotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyBotManagementBotImport,
		},
		Schema: map[string]*schema.Schema{
			"action": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The action applied to requests from the bot. One of `allow`, `block`, or `challenge` to override the category's action, or `inherit` to clear the override so the bot uses the category's action.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(botManagementBotActions, false)),
			},
			"bot_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the bot, such as `bot_gptbot`. Use the `fastly_bot_management_bots` data source to list the available IDs.",
			},
			"category_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the category the bot belongs to, such as `cat_ai_crawler`.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display name of the bot.",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Bot Management workspace the bot belongs to.",
			},
		},
	}
}

func resourceFastlyBotManagementBotCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	workspaceID := d.Get("workspace_id").(string)
	categoryID := d.Get("category_id").(string)
	botID := d.Get("bot_id").(string)

	if err := updateBotManagementBot(ctx, meta, workspaceID, categoryID, botID, d.Get("action").(string)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(workspaceID + "/" + categoryID + "/" + botID)

	return resourceFastlyBotManagementBotRead(ctx, d, meta)
}

func resourceFastlyBotManagementBotRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] REFRESH: Bot Management bot: %s", d.Id())

	bot, err := policy.GetBot(ctx, conn, &policy.GetBotInput{
		WorkspaceID: new(d.Get("workspace_id").(string)),
		CategoryID:  new(d.Get("category_id").(string)),
		BotID:       new(d.Get("bot_id").(string)),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Bot Management bot not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("action", bot.Action.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", bot.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyBotManagementBotUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	err := updateBotManagementBot(ctx, meta,
		d.Get("workspace_id").(string),
		d.Get("category_id").(string),
		d.Get("bot_id").(string),
		d.Get("action").(string),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceFastlyBotManagementBotRead(ctx, d, meta)
}

// Destroying the resource only removes it from state; the bot's action is left
// unchanged.
func resourceFastlyBotManagementBotDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	log.Printf("[DEBUG] DELETE: removing Bot Management bot %s from state; its action is left unchanged", d.Id())
	d.SetId("")
	return nil
}

func updateBotManagementBot(ctx context.Context, meta any, workspaceID, categoryID, botID, action string) error {
	conn := meta.(*APIClient).conn

	i := policy.UpdateBotInput{
		WorkspaceID: new(workspaceID),
		CategoryID:  new(categoryID),
		BotID:       new(botID),
		Action:      &policy.Action{Type: action},
	}

	log.Printf("[DEBUG] UPDATE: Bot Management bot input: %#v", i)

	_, err := policy.UpdateBot(ctx, conn, &i)
	return err
}

func resourceFastlyBotManagementBotImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	parts, err := parseBotManagementImportID(d.Id(), "workspace_id", "category_id", "bot_id")
	if err != nil {
		return nil, err
	}
	for i, k := range []string{"workspace_id", "category_id", "bot_id"} {
		if err := d.Set(k, parts[i]); err != nil {
			return nil, err
		}
	}
	return []*schema.ResourceData{d}, nil
}
