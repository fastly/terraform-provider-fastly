package fastly

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"sort"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceServiceWAFConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceWAFConfigurationCreate,
		ReadContext:   resourceServiceWAFConfigurationRead,
		UpdateContext: resourceServiceWAFConfigurationUpdate,
		DeleteContext: resourceServiceWAFConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceWAFConfigurationImport,
		},
		CustomizeDiff: validateWAFConfigurationResource,
		Schema: map[string]*schema.Schema{
			"waf_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Web Application Firewall that the configuration belongs to",
			},
			"allowed_http_versions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Allowed HTTP versions",
			},
			"allowed_methods": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "A space-separated list of HTTP method names",
			},
			"allowed_request_content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Allowed request content types",
			},
			"allowed_request_content_type_charset": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Allowed request content type charset",
			},
			"arg_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum number of arguments allowed",
			},
			"arg_name_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum allowed argument name length",
			},
			"combined_file_sizes": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum allowed size of all files",
			},
			"critical_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Score value to add for critical anomalies",
			},
			"crs_validate_utf8_encoding": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "CRS validate UTF8 encoding",
			},
			"error_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Score value to add for error anomalies",
			},
			"high_risk_country_codes": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "A space-separated list of country codes in ISO 3166-1 (two-letter) format",
			},
			"http_violation_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "HTTP violation threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"inbound_anomaly_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Inbound anomaly threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"lfi_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Local file inclusion attack threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"max_file_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum allowed file size, in bytes",
			},
			"max_num_args": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum number of arguments allowed",
			},
			"notice_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Score value to add for notice anomalies",
			},
			"paranoia_level": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The configured paranoia level",
			},
			"php_injection_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "PHP injection threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"rce_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Remote code execution threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"restricted_extensions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "A space-separated list of allowed file extensions",
			},
			"restricted_headers": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "A space-separated list of allowed header names",
			},
			"rfi_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Remote file inclusion attack threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"session_fixation_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Session fixation attack threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"sql_injection_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "SQL injection attack threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"total_arg_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum size of argument names and values",
			},
			"warning_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Score value to add for warning anomalies",
			},
			"xss_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "XSS attack threshold",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"rule":           activeRule,
			"rule_exclusion": wafRuleExclusion,
		},
	}
}

// this method calls update because the creation of the waf (within the service resource) automatically creates
// the first waf version, and this makes both a create and an updating exactly the same operation.
func resourceServiceWAFConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] creating configuration for WAF: %s", d.Get("waf_id").(string))
	d.SetId(d.Get("waf_id").(string))
	return resourceServiceWAFConfigurationUpdate(ctx, d, meta)
}

func resourceServiceWAFConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	latestVersion, err := getLatestVersion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	wafID := d.Get("waf_id").(string)
	log.Printf("[INFO] updating configuration for WAF: %s", wafID)
	if latestVersion.Locked {
		latestVersion, err = conn.CloneWAFVersion(&gofastly.CloneWAFVersionInput{
			WAFID:            wafID,
			WAFVersionNumber: latestVersion.Number,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	input := buildUpdateInput(d, latestVersion.ID, latestVersion.Number)
	if input.HasChanges() {
		latestVersion, err = conn.UpdateWAFVersion(input)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("rule") {
		if err := updateRules(d, meta, wafID, latestVersion.Number); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("rule_exclusion") {
		if err := updateWAFRuleExclusions(d, meta, wafID, latestVersion.Number); err != nil {
			return diag.FromErr(err)
		}
	}

	err = conn.DeployWAFVersion(&gofastly.DeployWAFVersionInput{
		WAFID:            wafID,
		WAFVersionNumber: latestVersion.Number,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	statusCheck := &WAFDeploymentChecker{
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      WAFStatusCheckDelay,
		MinTimeout: WAFStatusCheckMinTimeout,
		Check:      DefaultWAFDeploymentChecker(conn),
	}
	err = statusCheck.waitForDeployment(ctx, wafID, latestVersion)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceServiceWAFConfigurationRead(ctx, d, meta)
}

func resourceServiceWAFConfigurationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	latestVersion, err := getLatestVersion(d, meta)
	if err != nil {
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode == 404 {
				d.SetId("")
				return diag.Diagnostics{
					diag.Diagnostic{
						Severity:      diag.Warning,
						Summary:       fmt.Sprintf("WAF (%s) not found - removing from state", d.Get("waf_id")),
						AttributePath: cty.Path{cty.GetAttrStep{Name: "waf_id"}},
					},
				}
			}
		}
		return diag.FromErr(err)
	}

	log.Printf("[INFO] retrieving WAF version number: %d", latestVersion.Number)
	refreshWAFConfig(d, latestVersion)

	if err := readWAFRules(meta, d, latestVersion.Number); err != nil {
		return diag.FromErr(err)
	}

	// As the rule exclusion API is still behind a beta feature
	// flag, ensure we only read if the Set is non-empty.
	//
	// TODO(phamann): Remove d.GetOk() guard once in limited availability.
	if _, ok := d.GetOk("rule_exclusion"); ok {
		if err := readWAFRuleExclusions(meta, d, latestVersion.Number); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceServiceWAFConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	wafID := d.Get("waf_id").(string)
	log.Printf("[INFO] destroying configuration by creating empty version of WAF: %s", wafID)
	emptyVersion, err := conn.CreateEmptyWAFVersion(&gofastly.CreateEmptyWAFVersionInput{
		WAFID: wafID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = conn.DeployWAFVersion(&gofastly.DeployWAFVersionInput{
		WAFID:            wafID,
		WAFVersionNumber: emptyVersion.Number,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	statusCheck := &WAFDeploymentChecker{
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      WAFStatusCheckDelay,
		MinTimeout: WAFStatusCheckMinTimeout,
		Check:      DefaultWAFDeploymentChecker(conn),
	}
	err = statusCheck.waitForDeployment(ctx, wafID, emptyVersion)
	if err != nil {
		return diag.Errorf("Error waiting for WAF Version (%s) to be deleted: %s", d.Id(), err)
	}
	d.SetId("")
	return nil
}

func resourceServiceWAFConfigurationImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {

	wafID := d.Id()
	err := d.Set("waf_id", wafID)
	if err != nil {
		return nil, fmt.Errorf("error importing WAF configuration: WAF %s, %s", wafID, err)
	}
	return []*schema.ResourceData{d}, nil
}

func getLatestVersion(d *schema.ResourceData, meta interface{}) (*gofastly.WAFVersion, error) {
	conn := meta.(*FastlyClient).conn

	wafID := d.Get("waf_id").(string)
	resp, err := conn.ListAllWAFVersions(&gofastly.ListAllWAFVersionsInput{
		WAFID: wafID,
	})
	if err != nil {
		return nil, err
	}

	latest, err := determineLatestVersion(resp.Items)
	if err != nil {
		return nil, fmt.Errorf("[ERR] Error looking up WAF id: %s, with error %s", wafID, err)
	}
	return latest, nil
}

func buildUpdateInput(d *schema.ResourceData, id string, number int) *gofastly.UpdateWAFVersionInput {
	input := &gofastly.UpdateWAFVersionInput{
		WAFVersionID:     &id,
		WAFVersionNumber: &number,
	}
	if v, ok := d.GetOk("waf_id"); ok {
		input.WAFID = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("allowed_http_versions"); ok {
		input.AllowedHTTPVersions = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("allowed_methods"); ok {
		input.AllowedMethods = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("allowed_request_content_type"); ok {
		input.AllowedRequestContentType = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("allowed_request_content_type_charset"); ok {
		input.AllowedRequestContentTypeCharset = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("arg_length"); ok {
		input.ArgLength = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("arg_name_length"); ok {
		input.ArgNameLength = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("combined_file_sizes"); ok {
		input.CombinedFileSizes = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("critical_anomaly_score"); ok {
		input.CriticalAnomalyScore = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("crs_validate_utf8_encoding"); ok {
		input.CRSValidateUTF8Encoding = gofastly.Bool(v.(bool))
	}
	if v, ok := d.GetOk("error_anomaly_score"); ok {
		input.ErrorAnomalyScore = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("high_risk_country_codes"); ok {
		input.HighRiskCountryCodes = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("http_violation_score_threshold"); ok {
		input.HTTPViolationScoreThreshold = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("inbound_anomaly_score_threshold"); ok {
		input.InboundAnomalyScoreThreshold = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("lfi_score_threshold"); ok {
		input.LFIScoreThreshold = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("max_file_size"); ok {
		input.MaxFileSize = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("max_num_args"); ok {
		input.MaxNumArgs = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("notice_anomaly_score"); ok {
		input.NoticeAnomalyScore = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("paranoia_level"); ok {
		input.ParanoiaLevel = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("php_injection_score_threshold"); ok {
		input.PHPInjectionScoreThreshold = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("rce_score_threshold"); ok {
		input.RCEScoreThreshold = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("restricted_extensions"); ok {
		input.RestrictedExtensions = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("restricted_headers"); ok {
		input.RestrictedHeaders = gofastly.String(v.(string))
	}
	if v, ok := d.GetOk("rfi_score_threshold"); ok {
		input.RFIScoreThreshold = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("session_fixation_score_threshold"); ok {
		input.SessionFixationScoreThreshold = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("sql_injection_score_threshold"); ok {
		input.SQLInjectionScoreThreshold = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("total_arg_length"); ok {
		input.TotalArgLength = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("warning_anomaly_score"); ok {
		input.WarningAnomalyScore = gofastly.Int(v.(int))
	}
	if v, ok := d.GetOk("xss_score_threshold"); ok {
		input.XSSScoreThreshold = gofastly.Int(v.(int))
	}
	return input
}

func refreshWAFConfig(d *schema.ResourceData, version *gofastly.WAFVersion) {
	d.Set("allowed_http_versions", version.AllowedHTTPVersions)
	d.Set("allowed_methods", version.AllowedMethods)
	d.Set("allowed_request_content_type", version.AllowedRequestContentType)
	d.Set("allowed_request_content_type_charset", version.AllowedRequestContentTypeCharset)
	d.Set("arg_length", version.ArgLength)
	d.Set("arg_name_length", version.ArgNameLength)
	d.Set("combined_file_sizes", version.CombinedFileSizes)
	d.Set("critical_anomaly_score", version.CriticalAnomalyScore)
	d.Set("crs_validate_utf8_encoding", version.CRSValidateUTF8Encoding)
	d.Set("error_anomaly_score", version.ErrorAnomalyScore)
	d.Set("high_risk_country_codes", version.HighRiskCountryCodes)
	d.Set("http_violation_score_threshold", version.HTTPViolationScoreThreshold)
	d.Set("inbound_anomaly_score_threshold", version.InboundAnomalyScoreThreshold)
	d.Set("lfi_score_threshold", version.LFIScoreThreshold)
	d.Set("max_file_size", version.MaxFileSize)
	d.Set("max_num_args", version.MaxNumArgs)
	d.Set("notice_anomaly_score", version.NoticeAnomalyScore)
	d.Set("paranoia_level", version.ParanoiaLevel)
	d.Set("php_injection_score_threshold", version.PHPInjectionScoreThreshold)
	d.Set("rce_score_threshold", version.RCEScoreThreshold)
	d.Set("restricted_extensions", version.RestrictedExtensions)
	d.Set("restricted_headers", version.RestrictedHeaders)
	d.Set("rfi_score_threshold", version.RFIScoreThreshold)
	d.Set("session_fixation_score_threshold", version.SessionFixationScoreThreshold)
	d.Set("sql_injection_score_threshold", version.SQLInjectionScoreThreshold)
	d.Set("total_arg_length", version.TotalArgLength)
	d.Set("warning_anomaly_score", version.WarningAnomalyScore)
	d.Set("xss_score_threshold", version.XSSScoreThreshold)
}

func determineLatestVersion(versions []*gofastly.WAFVersion) (*gofastly.WAFVersion, error) {

	if len(versions) == 0 {
		return nil, errors.New("the list of WAFVersions cannot be empty")
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Number > versions[j].Number
	})

	return versions[0], nil
}

func validateWAFConfigurationResource(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	err := validateWAFRuleExclusion(d)
	return err
}
