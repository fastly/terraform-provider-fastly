package fastly

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/botmanagement/v1/workspaces/policy"
)

var botManagementCategoryActions = []string{"allow", "block", "challenge"}

func resourceFastlyBotManagementCategory() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the action of a predefined bot category in a Fastly Bot Management workspace. This resource doesn't create or delete categories, and destroying the resource only removes it from Terraform state.",
		CreateContext: resourceFastlyBotManagementCategoryCreate,
		ReadContext:   resourceFastlyBotManagementCategoryRead,
		UpdateContext: resourceFastlyBotManagementCategoryUpdate,
		DeleteContext: resourceFastlyBotManagementCategoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyBotManagementCategoryImport,
		},
		Schema: map[string]*schema.Schema{
			"action": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The action applied to every bot in the category that is set to `inherit`, including bots added to the category later. One of `allow`, `block`, or `challenge`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(botManagementCategoryActions, false)),
			},
			"category_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the bot category, such as `cat_ai_crawler`. Use the `fastly_bot_management_categories` data source to list the available IDs.",
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
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Bot Management workspace the category belongs to.",
			},
		},
	}
}

func resourceFastlyBotManagementCategoryCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	workspaceID := d.Get("workspace_id").(string)
	categoryID := d.Get("category_id").(string)

	if diags := updateBotManagementCategory(ctx, d, meta, workspaceID, categoryID); diags.HasError() {
		return diags
	}

	d.SetId(workspaceID + "/" + categoryID)

	return resourceFastlyBotManagementCategoryRead(ctx, d, meta)
}

func resourceFastlyBotManagementCategoryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	workspaceID := d.Get("workspace_id").(string)
	categoryID := d.Get("category_id").(string)

	log.Printf("[DEBUG] REFRESH: Bot Management category: %s", d.Id())

	category, err := policy.GetCategory(ctx, conn, &policy.GetCategoryInput{
		WorkspaceID: new(workspaceID),
		CategoryID:  new(categoryID),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Bot Management category not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("action", category.Action.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", category.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", category.Description); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyBotManagementCategoryUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if diags := updateBotManagementCategory(ctx, d, meta, d.Get("workspace_id").(string), d.Get("category_id").(string)); diags.HasError() {
		return diags
	}
	return resourceFastlyBotManagementCategoryRead(ctx, d, meta)
}

func updateBotManagementCategory(ctx context.Context, d *schema.ResourceData, meta any, workspaceID, categoryID string) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := policy.UpdateCategoryInput{
		WorkspaceID: new(workspaceID),
		CategoryID:  new(categoryID),
		Action:      &policy.Action{Type: d.Get("action").(string)},
	}

	log.Printf("[DEBUG] UPDATE: Bot Management category input: %#v", i)

	if _, err := policy.UpdateCategory(ctx, conn, &i); err != nil {
		return diag.Errorf("error updating Bot Management category %s/%s: %s", workspaceID, categoryID, err)
	}
	return nil
}

// Categories have no neutral action, so destroying the resource
// only removes it from state.
func resourceFastlyBotManagementCategoryDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	log.Printf("[DEBUG] DELETE: removing Bot Management category %s from state; its action is left unchanged", d.Id())
	d.SetId("")
	return nil
}

func resourceFastlyBotManagementCategoryImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	parts, err := parseBotManagementImportID(d.Id(), "workspace_id", "category_id")
	if err != nil {
		return nil, err
	}
	if err := d.Set("workspace_id", parts[0]); err != nil {
		return nil, err
	}
	if err := d.Set("category_id", parts[1]); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

// parseBotManagementImportID splits a slash-separated import ID into exactly
// len(names) non-empty parts.
func parseBotManagementImportID(id string, names ...string) ([]string, error) {
	parts := strings.Split(id, "/")
	if len(parts) == len(names) && !slices.Contains(parts, "") {
		return parts, nil
	}
	return nil, fmt.Errorf("invalid ID format: %q. Expected format: <%s>", id, strings.Join(names, ">/<"))
}
