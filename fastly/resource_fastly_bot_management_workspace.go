package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
	bmws "github.com/fastly/go-fastly/v19/fastly/botmanagement/v1/workspaces"
)

var botManagementProtectionModes = []string{"off", "log", "block"}

func resourceFastlyBotManagementWorkspace() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages the settings of an existing Fastly Bot Management workspace. This resource doesn't create or delete the workspace, and destroying the resource only removes it from Terraform state.",
		CreateContext: resourceFastlyBotManagementWorkspaceCreate,
		ReadContext:   resourceFastlyBotManagementWorkspaceRead,
		UpdateContext: resourceFastlyBotManagementWorkspaceUpdate,
		DeleteContext: resourceFastlyBotManagementWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyBotManagementWorkspaceImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "A description of the Bot Management workspace. Must be at most 1000 characters. When omitted, the current description is kept.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 1000)),
			},
			"name": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "The display name of the Bot Management workspace. Must be between 1 and 255 characters. When omitted, the current name is kept.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 255)),
			},
			"protection_mode": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The protection mode of the Bot Management workspace, which gates all enforcement. One of `off` (no enforcement), `log` (bots are detected and logged, but no requests are blocked or challenged), or `block` (category and bot actions are enforced).",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(botManagementProtectionModes, false)),
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Bot Management workspace to manage. Use the `fastly_bot_management_workspace` data source to look it up from a service ID.",
			},
		},
	}
}

func resourceFastlyBotManagementWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	workspaceID := d.Get("workspace_id").(string)

	i := bmws.UpdateInput{
		WorkspaceID:    new(workspaceID),
		ProtectionMode: new(d.Get("protection_mode").(string)),
	}
	// Only send name/description when they are set in config, so that
	// omitting them keeps the workspace's current values.
	rawConfig := d.GetRawConfig()
	if v := rawConfig.GetAttr("name"); !v.IsNull() {
		i.Name = new(v.AsString())
	}
	if v := rawConfig.GetAttr("description"); !v.IsNull() {
		i.Description = new(v.AsString())
	}

	log.Printf("[DEBUG] CREATE: Bot Management workspace input: %#v", i)

	if _, err := bmws.Update(ctx, conn, &i); err != nil {
		return diag.Errorf("error updating Bot Management workspace %s: %s", workspaceID, err)
	}

	d.SetId(workspaceID)

	return resourceFastlyBotManagementWorkspaceRead(ctx, d, meta)
}

func resourceFastlyBotManagementWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] REFRESH: Bot Management workspace: %s", d.Id())

	workspace, err := bmws.Get(ctx, conn, &bmws.GetInput{WorkspaceID: new(d.Id())})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Bot Management workspace not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("workspace_id", workspace.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", workspace.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", workspace.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("protection_mode", workspace.ProtectionMode); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyBotManagementWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := bmws.UpdateInput{WorkspaceID: new(d.Id())}
	if d.HasChange("name") {
		i.Name = new(d.Get("name").(string))
	}
	if d.HasChange("description") {
		i.Description = new(d.Get("description").(string))
	}
	if d.HasChange("protection_mode") {
		i.ProtectionMode = new(d.Get("protection_mode").(string))
	}

	log.Printf("[DEBUG] UPDATE: Bot Management workspace input: %#v", i)

	if _, err := bmws.Update(ctx, conn, &i); err != nil {
		return diag.Errorf("error updating Bot Management workspace %s: %s", d.Id(), err)
	}

	return resourceFastlyBotManagementWorkspaceRead(ctx, d, meta)
}

// Workspaces can't be deleted through the API, so destroying the resource only
// removes it from state.
func resourceFastlyBotManagementWorkspaceDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	log.Printf("[DEBUG] DELETE: removing Bot Management workspace %s from state; its settings are left unchanged", d.Id())
	d.SetId("")
	return nil
}

func resourceFastlyBotManagementWorkspaceImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	if err := d.Set("workspace_id", d.Id()); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
