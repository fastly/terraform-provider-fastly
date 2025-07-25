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
	wsr "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/redactions"
)

func resourceFastlyNGWAFRedaction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFRedactionCreate,
		ReadContext:   resourceFastlyNGWAFRedactionRead,
		UpdateContext: resourceFastlyNGWAFRedactionUpdate,
		DeleteContext: resourceFastlyNGWAFRedactionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFRedactionsImport,
		},
		Schema: map[string]*schema.Schema{
			"field": {
				Description: "The name of the field that should be redacted.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Type:             schema.TypeString,
				Description:      "The type of field that is being redacted. One of `request_parameter`, `request_header`, or `response_header`.",
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"request_parameter", "request_header", "response_header"}, false)),
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Description: "The id of the workspace this redaction belongs to.",
				Required:    true,
			},
		},
	}
}

func resourceFastlyNGWAFRedactionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := wsr.CreateInput{
		Field:       gofastly.ToPointer(d.Get("field").(string)),
		Type:        gofastly.ToPointer(d.Get("type").(string)),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF redaction input: %#v", i)

	redaction, err := wsr.Create(ctx, conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF redaction result: %#v", redaction)

	d.SetId(redaction.RedactionID)

	return resourceFastlyNGWAFRedactionRead(ctx, d, meta)
}

func resourceFastlyNGWAFRedactionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := wsr.GetInput{
		RedactionID: gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF redaction input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	redaction, err := wsr.Get(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] redaction not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("field", redaction.Field); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", redaction.Type); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFRedactionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := wsr.UpdateInput{
		Field:       gofastly.ToPointer(d.Get("field").(string)),
		RedactionID: gofastly.ToPointer(d.Id()),
		Type:        gofastly.ToPointer(d.Get("type").(string)),
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF redaction input: %#v", i)

	_, err := wsr.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFRedactionRead(ctx, d, meta)
}

func resourceFastlyNGWAFRedactionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := wsr.DeleteInput{
		RedactionID: gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF redaction input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := wsr.Delete(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFRedactionsImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF redaction ID: %s", d.Id())

	workspaceID, redactionID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<redactionID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(redactionID)

	return []*schema.ResourceData{d}, nil
}
