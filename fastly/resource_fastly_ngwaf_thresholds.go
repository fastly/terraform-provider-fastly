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
	wsr "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/thresholds"
)

func resourceFastlyNGWAFThresholds() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFThresholdsCreate,
		ReadContext:   resourceFastlyNGWAFThresholdsRead,
		UpdateContext: resourceFastlyNGWAFThresholdsUpdate,
		DeleteContext: resourceFastlyNGWAFThresholdsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyNGWAFThresholdssImport,
		},
		Schema: map[string]*schema.Schema{
			"action": {
				Type:             schema.TypeString,
				Description:      "Action to take when threshold is exceeded",
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"block", "log"}, false)),
			},
			"dont_notify": {
				Type:        schema.TypeBool,
				Description: "Whether to silence notifications when action is taken",
				Required:    true,
			},
			"duration": {
				Type:             schema.TypeInt,
				Description:      "Duration the action is in place. Default duration is 86,400 seconds (1 day)",
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 31556900)),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether this threshold is active",
				Required:    true,
			},
			"interval": {
				Type:             schema.TypeInt,
				Description:      "Threshold interval in seconds. Default interval is 3600 seconds (1 hour)",
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntInSlice([]int{60, 600, 3600})),
			},
			"limit": {
				Type:             schema.TypeInt,
				Description:      "Threshold limit. Default limit is 10",
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10000)),
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Threshold name",
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(3, 50)),
			},
			"signal": {
				Type: schema.TypeString,
				// Update to ref custom / system signals.
				Description: "The name of the signal this threshold is acting on",
				Required:    true,
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Description: "The unique identifier of the workspace.",
				Required:    true,
			},
		},
	}
}

func resourceFastlyNGWAFThresholdsCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := wsr.CreateInput{
		Action:      gofastly.ToPointer(d.Get("action").(string)),
		DontNotify:  gofastly.ToPointer(d.Get("dont_notify").(bool)),
		Duration:    gofastly.ToPointer(d.Get("duration").(int)),
		Enabled:     gofastly.ToPointer(d.Get("enabled").(bool)),
		Interval:    gofastly.ToPointer(d.Get("interval").(int)),
		Limit:       gofastly.ToPointer(d.Get("limit").(int)),
		Name:        gofastly.ToPointer(d.Get("name").(string)),
		Signal:      gofastly.ToPointer(d.Get("signal").(string)),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] CREATE: NGWAF threshold input: %#v", i)

	threshold, err := wsr.Create(ctx, conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] CREATE: NGWAF threshold result: %#v", threshold)

	d.SetId(threshold.ThresholdID)

	return resourceFastlyNGWAFThresholdsRead(ctx, d, meta)
}

func resourceFastlyNGWAFThresholdsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := wsr.GetInput{
		ThresholdID: gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF threshold input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	threshold, err := wsr.Get(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] threshold not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("action", threshold.Action); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dont_notify", threshold.DontNotify); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("duration", threshold.Duration); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", threshold.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interval", threshold.Interval); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("limit", threshold.Limit); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", threshold.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("signal", threshold.Signal); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFThresholdsUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := wsr.UpdateInput{
		Action:      gofastly.ToPointer(d.Get("action").(string)),
		DontNotify:  gofastly.ToPointer(d.Get("dont_notify").(bool)),
		Duration:    gofastly.ToPointer(d.Get("duration").(int)),
		Enabled:     gofastly.ToPointer(d.Get("enabled").(bool)),
		Interval:    gofastly.ToPointer(d.Get("interval").(int)),
		Limit:       gofastly.ToPointer(d.Get("limit").(int)),
		Name:        gofastly.ToPointer(d.Get("name").(string)),
		Signal:      gofastly.ToPointer(d.Get("signal").(string)),
		ThresholdID: gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] UPDATE: NGWAF threshold input: %#v", i)

	_, err := wsr.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFThresholdsRead(ctx, d, meta)
}

func resourceFastlyNGWAFThresholdsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	i := wsr.DeleteInput{
		ThresholdID: gofastly.ToPointer(d.Id()),
		WorkspaceID: gofastly.ToPointer(workspaceID),
	}

	log.Printf("[DEBUG] DELETE: NGWAF threshold input: id=%s, workspaceID=%s", d.Id(), workspaceID)

	if err := wsr.Delete(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFThresholdssImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] IMPORT: NGWAF threshold ID: %s", d.Id())

	workspaceID, thresholdID, isInCorrectForm := strings.Cut(d.Id(), "/")
	if !isInCorrectForm {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <workspaceID>/<thresholdID>", d.Id())
	}

	if err := d.Set("workspace_id", workspaceID); err != nil {
		return nil, fmt.Errorf("error setting workspace_id (%s): %w", workspaceID, err)
	}

	d.SetId(thresholdID)

	return []*schema.ResourceData{d}, nil
}
