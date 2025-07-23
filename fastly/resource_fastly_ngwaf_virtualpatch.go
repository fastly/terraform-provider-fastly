package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	ws "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/virtualpatches"
)

func resourceFastlyNGWAFVirtualPatch() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFVirtualPatchUpdate,
		DeleteContext: resourceFastlyNGWAFVirtualPatchDelete,
		ReadContext:   resourceFastlyNGWAFVirtualPatchRead,
		UpdateContext: resourceFastlyNGWAFVirtualPatchUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"action": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User-configured action of the virtual patch. Accepted values are [`block`, `log`]",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{"log", "block"},
					false,
				)),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "User-configured control for enabling or disabling virtual patches",
			},
			"virtual_patch_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "User-submitted identifier (ID) of the virtual patch",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
			"workspace_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "User-submitted identifier (ID) of the workspace",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotWhiteSpace),
			},
		},
	}
}

func resourceFastlyNGWAFVirtualPatchRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.GetInput{
		WorkspaceID:    gofastly.ToPointer(d.Get("workspace_id").(string)),
		VirtualPatchID: gofastly.ToPointer(d.Get("virtual_patch_id").(string)),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF virtual patch input: %#v", i)

	virtualpatch, err := ws.Get(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
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

	_, err := ws.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set ID for the create operation
	if d.Id() == "" {
		d.SetId(fmt.Sprintf("%s/%s", d.Get("workspace_id").(string), d.Get("virtual_patch_id").(string)))
	}

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

	_, err := ws.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	return diag.FromErr(err)

}
