package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	ws "github.com/fastly/go-fastly/v10/fastly/ngwaf/v1/workspaces"
)

func resourceFastlyNGWAFWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyNGWAFWorkspaceCreate,
		ReadContext:   resourceFastlyNGWAFWorkspaceRead,
		UpdateContext: resourceFastlyNGWAFWorkspaceUpdate,
		DeleteContext: resourceFastlyNGWAFWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the NGWAF Workspace.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of the NGWAF Workspace.",
			},
			"mode": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The mode of the NGWAF Workspace.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{"off", "log", "block"},
					false,
				)),
			},
			"ip_anonymization": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Whether IPs should be anonymized in the NGWAF Workspace.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"hashed"}, false)),
			},
			"client_ip_headers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of headers used to determine the client IP address.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				MaxItems:    10,
			},
			"default_blocking_response_code": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     406,
				Description: "Default HTTP response code for blocking actions.",
			},
			"attack_signal_thresholds": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"one_minute": {
							Type:             schema.TypeInt,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10000)),
						},
						"ten_minutes": {
							Type:             schema.TypeInt,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10000)),
						},
						"one_hour": {
							Type:             schema.TypeInt,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10000)),
						},
						"immediate": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				MaxItems:    1,
				Description: "Attack signal thresholds configuration.",
			},
		},
	}
}

func resourceFastlyNGWAFWorkspaceCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.CreateInput{
		Name:            gofastly.ToPointer(d.Get("name").(string)),
		Description:     gofastly.ToPointer(d.Get("description").(string)),
		Mode:            gofastly.ToPointer(d.Get("mode").(string)),
		IPAnonymization: gofastly.ToPointer(d.Get("ip_anonymization").(string)),
	}

	if v, ok := d.GetOk("client_ip_headers"); ok {
		rawList := v.([]any)
		clientHeaders := make([]string, len(rawList))
		for i, item := range rawList {
			clientHeaders[i] = item.(string)
		}
		i.ClientIPHeaders = clientHeaders
	}

	if d.HasChange("default_blocking_response_code") || d.Get("default_blocking_response_code") != 406 {
		code := d.Get("default_blocking_response_code").(int)
		i.DefaultBlockingResponseCode = &code
	}

	if v, ok := d.GetOk("attack_signal_thresholds"); ok && len(v.([]any)) > 0 {
		th := v.([]any)[0].(map[string]any)
		i.AttackSignalThresholds = &ws.AttackSignalThresholdsCreateInput{
			OneMinute:  gofastly.ToPointer(th["one_minute"].(int)),
			TenMinutes: gofastly.ToPointer(th["ten_minutes"].(int)),
			OneHour:    gofastly.ToPointer(th["one_hour"].(int)),
			Immediate:  gofastly.ToPointer(th["immediate"].(bool)),
		}
	}

	log.Printf("[DEBUG] CREATE: NGWAF Workspace input: %#v", i)

	workspace, err := ws.Create(conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(workspace.WorkspaceID)

	return nil
}

func resourceFastlyNGWAFWorkspaceRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.GetInput{
		WorkspaceID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF Workspace input: %#v", i)

	workspace, err := ws.Get(conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] NGWAF Workspace not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("name", workspace.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", workspace.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mode", workspace.Mode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ip_anonymization", workspace.IPAnonymization); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_ip_headers", workspace.ClientIPHeaders); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_blocking_response_code", workspace.DefaultBlockingResponseCode); err != nil {
		return diag.FromErr(err)
	}

	thresholds := []map[string]any{
		{
			"one_minute":  workspace.AttackSignalThresholds.OneMinute,
			"ten_minutes": workspace.AttackSignalThresholds.TenMinutes,
			"one_hour":    workspace.AttackSignalThresholds.OneHour,
			"immediate":   workspace.AttackSignalThresholds.Immediate,
		},
	}
	if err := d.Set("attack_signal_thresholds", thresholds); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyNGWAFWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.UpdateInput{
		WorkspaceID:     gofastly.ToPointer(d.Id()),
		Name:            gofastly.ToPointer(d.Get("name").(string)),
		Description:     gofastly.ToPointer(d.Get("description").(string)),
		Mode:            gofastly.ToPointer(d.Get("mode").(string)),
		IPAnonymization: gofastly.ToPointer(d.Get("ip_anonymization").(string)),
	}

	if v, ok := d.GetOk("client_ip_headers"); ok {
		rawList := v.([]any)
		clientHeaders := make([]string, len(rawList))
		for i, item := range rawList {
			clientHeaders[i] = item.(string)
		}
		i.ClientIPHeaders = clientHeaders
	}

	if d.HasChange("default_blocking_response_code") || d.Get("default_blocking_response_code") != 406 {
		code := d.Get("default_blocking_response_code").(int)
		i.DefaultBlockingResponseCode = &code
	}

	if v, ok := d.GetOk("attack_signal_thresholds"); ok && len(v.([]any)) > 0 {
		th := v.([]any)[0].(map[string]any)
		i.AttackSignalThresholds = &ws.AttackSignalThresholdsUpdateInput{
			OneMinute:  gofastly.ToPointer(th["one_minute"].(int)),
			TenMinutes: gofastly.ToPointer(th["ten_minutes"].(int)),
			OneHour:    gofastly.ToPointer(th["one_hour"].(int)),
			Immediate:  gofastly.ToPointer(th["immediate"].(bool)),
		}
	}

	log.Printf("[DEBUG] UPDATE: NGWAF Workspace input: %#v", i)

	_, err := ws.Update(conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFWorkspaceRead(ctx, d, meta)
}

func resourceFastlyNGWAFWorkspaceDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.DeleteInput{
		WorkspaceID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] DELETE: NGWAF Workspace input: %#v", i)

	if err := ws.Delete(conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
