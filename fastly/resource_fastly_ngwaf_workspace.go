package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	ws "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces"
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
			"attack_signal_thresholds": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"immediate": {
							Type:        schema.TypeBool,
							Description: "Ignore thresholds and block immediately when at least one attack signal is detected",
							Optional:    true,
						},
						"one_hour": {
							Type:             schema.TypeInt,
							Description:      "The one-hour interval threshold. Minimum 1 and maximum 10,000",
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10000)),
						},
						"one_minute": {
							Type:             schema.TypeInt,
							Description:      "The one-minute interval threshold. Minimum 1 and maximum 10,000",
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10000)),
						},
						"ten_minutes": {
							Type:             schema.TypeInt,
							Description:      "The ten-minute interval threshold. Minimum 1 and maximum 10,000",
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 10000)),
						},
					},
				},
				MaxItems:    1,
				Description: "Attack threshold parameters for system site alerts. Each threshold value is the number of attack signals per IP address that must be detected during the interval before the related IP address is flagged",
			},
			"client_ip_headers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Specifies the request headers containing the client IP address. Maximum of 10 header names",
				Elem:        &schema.Schema{Type: schema.TypeString},
				MaxItems:    10,
			},
			"default_blocking_response_code": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          406,
				Description:      "The status code returned when a request is blocked. This configuration is applied at the workspace but can be overwritten in rules. Accepted values are [`301`, `302`, `400..599`]. Default value `406`",
				ValidateDiagFunc: validation.ToDiagFunc(validation.Any(validation.IntBetween(400, 599), validation.IntInSlice([]int{301, 302}))),
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User-submitted description of the workspace",
			},
			"ip_anonymization": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Agents will anonymize IP addresses according to the option selected. Accepted value is `hashed`",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"hashed"}, false)),
			},
			"mode": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User-configured mode of the workspace. Accepted values are [`off`, `block`, `log`]",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{"off", "log", "block"},
					false,
				)),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User-submitted display name of the workspace",
			},
		},
	}
}

func resourceFastlyNGWAFWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	log.Printf("[DEBUG] CREATE: NGWAF workspace input: %#v", i)

	workspace, err := ws.Create(ctx, conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(workspace.WorkspaceID)

	return resourceFastlyNGWAFWorkspaceRead(ctx, d, meta)
}

func resourceFastlyNGWAFWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.GetInput{
		WorkspaceID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] REFRESH: NGWAF workspace input: %#v", i)

	workspace, err := ws.Get(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] workspace not found '%s'", d.Id())
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

	log.Printf("[DEBUG] UPDATE: NGWAF workspace attack thresholds get: %#v", workspace.AttackSignalThresholds)

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

	log.Printf("[DEBUG] UPDATE: NGWAF workspace input: %#v", i)

	_, err := ws.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyNGWAFWorkspaceRead(ctx, d, meta)
}

func resourceFastlyNGWAFWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := ws.DeleteInput{
		WorkspaceID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] DELETE: NGWAF workspace input: %#v", i)

	if err := ws.Delete(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
