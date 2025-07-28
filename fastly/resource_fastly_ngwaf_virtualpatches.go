package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	ws "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/virtualpatches"
)

func resourceFastlyNGWAFVirtualPatches() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFVirtualPatchCreate,
		DeleteContext: resourceFastlyNGWAFVirtualPatchDelete,
		ReadContext:   resourceFastlyNGWAFVirtualPatchRead,
		UpdateContext: resourceFastlyNGWAFVirtualPatchUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFVirtualPatchImport,
		},
		Schema: map[string]*schema.Schema{
			"action": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Action to take when a signal for virtual patch is detected. One of `log` or `block`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{"log", "block"},
					false,
				)),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Control for enabling and disabling a virtual patch. One of `true` or `false`.",
			},
			"virtual_patch_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The ID of the virtual patch.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
			"workspace_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The ID of the workspace.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
		},
	}
}

func resourceFastlyNGWAFVirtualPatchCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.GetInput{
		WorkspaceID:    gofastly.ToPointer(d.Get("workspace_id").(string)),
		VirtualPatchID: gofastly.ToPointer(d.Get("virtual_patch_id").(string)),
	}

	log.Printf("[DEBUG] CREATE: NGWAF virtual patch input: %#v", i)

	// Check if the virtual patch exists
	_, err := ws.Get(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			return diag.Errorf("Virtual Patches cannot be created. Use this resource to configure existing Virtual Patches. Virtual patch ID '%s' does not exist in workspace '%s'.",
				d.Get("virtual_patch_id").(string),
				d.Get("workspace_id").(string))
		}
		return diag.FromErr(err)
	}

	// Virtual patch exists, set the ID and update it with the desired configuration
	d.SetId(d.Get("virtual_patch_id").(string))

	// Apply the desired configuration by calling update
	return resourceFastlyNGWAFVirtualPatchUpdate(ctx, d, meta)
}

func resourceFastlyNGWAFVirtualPatchRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.GetInput{
		WorkspaceID:    gofastly.ToPointer(d.Get("workspace_id").(string)),
		VirtualPatchID: gofastly.ToPointer(d.Get("virtual_patch_id").(string)),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF virtual patch input: %#v", i)

	virtualpatch, err := ws.Get(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] virtual patch not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("action", virtualpatch.Mode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", virtualpatch.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("virtual_patch_id", virtualpatch.ID); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceFastlyNGWAFVirtualPatchUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.UpdateInput{
		WorkspaceID:    gofastly.ToPointer(d.Get("workspace_id").(string)),
		VirtualPatchID: gofastly.ToPointer(d.Get("virtual_patch_id").(string)),
		Mode:           gofastly.ToPointer(d.Get("action").(string)),
		Enabled:        gofastly.ToPointer(d.Get("enabled").(bool)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF virtualpatch input: %#v", i)

	virtualPatch, err := ws.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(virtualPatch.ID)

	return resourceFastlyNGWAFVirtualPatchRead(ctx, d, meta)
}

func resourceFastlyNGWAFVirtualPatchDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.UpdateInput{
		WorkspaceID:    gofastly.ToPointer(d.Get("workspace_id").(string)),
		VirtualPatchID: gofastly.ToPointer(d.Get("virtual_patch_id").(string)),
		Mode:           gofastly.ToPointer(d.Get("action").(string)),
		// Disable virtual patch on delete
		Enabled: gofastly.ToPointer(false),
	}

	log.Printf("[DEBUG] DELETE: NGWAF virtual patch input: %#v", i)

	_, err := ws.Update(gofastly.NewContextForResourceID(ctx, d.Get("workspace_id").(string)), conn, &i)
	return diag.FromErr(err)
}

func resourceFastlyNGWAFVirtualPatchImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	// The import ID should be in format: workspaceID/virtualPatchID
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format, expected workspaceID/virtualPatchID, got: %s", d.Id())
	}

	workspaceID := parts[0]
	virtualPatchID := parts[1]

	log.Printf("[DEBUG] IMPORT: workspaceID = %s, virtualPatchID = %s", workspaceID, virtualPatchID)

	// Set the individual attributes
	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id: %w", err)
	}
	if err := d.Set("virtual_patch_id", virtualPatchID); err != nil {
		return nil, fmt.Errorf("error setting virtual_patch_id: %w", err)
	}

	d.SetId(virtualPatchID)

	// Call the read function to populate the rest of the attributes
	diags := resourceFastlyNGWAFVirtualPatchRead(ctx, d, meta)
	if diags.HasError() {
		return nil, fmt.Errorf("error reading virtual patch during import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
