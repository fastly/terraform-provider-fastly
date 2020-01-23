package fastly

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceServiceWAFConfigurationV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceWAFConfigurationV1Create,
		Read:   resourceServiceWAFConfigurationV1Read,
		Update: resourceServiceWAFConfigurationV1Update,
		Delete: resourceServiceWAFConfigurationV1Delete,
		Importer: &schema.ResourceImporter{
			State: resourceServiceWAFConfigurationV1Import,
		},

		Schema: map[string]*schema.Schema{
			"waf_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The service the WAF belongs to.",
			},
			"allowed_http_versions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allowed HTTP versions (default HTTP/1.0 HTTP/1.1 HTTP/2).",
			},
			"allowed_methods": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A space-separated list of HTTP method names (default GET HEAD POST OPTIONS PUT PATCH DELETE).",
			},
			"allowed_request_content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allowed request content types (default application/x-www-form-urlencoded|multipart/form-data|text/xml|application/xml|application/x-amf|application/json|text/plain).",
			},
			"allowed_request_content_type_charset": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allowed request content type charset (default utf-8|iso-8859-1|iso-8859-15|windows-1252).",
			},
			"arg_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum number of arguments allowed (default 400).",
			},
			"arg_name_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum allowed argument name length (default 100).",
			},
			"combined_file_sizes": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum allowed size of all files (in bytes, default 10000000).",
			},
			"critical_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Score value to add for critical anomalies (default 6).",
			},
			"crs_validate_utf8_encoding": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "CRS validate UTF8 encoding.",
			},
			"error_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Score value to add for error anomalies (default 5).",
			},
			"high_risk_country_codes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A space-separated list of country codes in ISO 3166-1 (two-letter) format.",
			},
			"http_violation_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "HTTP violation threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"inbound_anomaly_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Inbound anomaly threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"lfi_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Local file inclusion attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"max_file_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum allowed file size, in bytes (default 10000000).",
			},
			"max_num_args": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum number of arguments allowed (default 255).",
			},
			"notice_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Score value to add for notice anomalies (default 4).",
			},
			"paranoia_level": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The configured paranoia level (default 1).",
			},
			"php_injection_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "PHP injection threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"rce_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Remote code execution threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"restricted_extensions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A space-separated list of allowed file extensions (default .asa/ .asax/ .ascx/ .axd/ .backup/ .bak/ .bat/ .cdx/ .cer/ .cfg/ .cmd/ .com/ .config/ .conf/ .cs/ .csproj/ .csr/ .dat/ .db/ .dbf/ .dll/ .dos/ .htr/ .htw/ .ida/ .idc/ .idq/ .inc/ .ini/ .key/ .licx/ .lnk/ .log/ .mdb/ .old/ .pass/ .pdb/ .pol/ .printer/ .pwd/ .resources/ .resx/ .sql/ .sys/ .vb/ .vbs/ .vbproj/ .vsdisco/ .webinfo/ .xsd/ .xsx).",
			},
			"restricted_headers": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A space-separated list of allowed header names (default /proxy/ /lock-token/ /content-range/ /translate/ /if/).",
			},
			"rfi_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Remote file inclusion attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"session_fixation_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Session fixation attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"sql_injection_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "SQL injection attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"total_arg_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum size of argument names and values (default 6400).",
			},
			"warning_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Score value to add for warning anomalies.",
			},
			"xss_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "XSS attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"rule": activeRule,
		},
	}
}

// this method calls update because the creation of the waf (within the service resource) automatically creates
// the first waf version, and this makes both a create and an updating exactly the same operation.
func resourceServiceWAFConfigurationV1Create(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] creating configuration for WAF: %s", d.Get("waf_id").(string))
	d.SetId(d.Get("waf_id").(string))
	return resourceServiceWAFConfigurationV1Update(d, meta)
}

func resourceServiceWAFConfigurationV1Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	latestVersion, err := getLatestVersion(d, meta)
	if err != nil {
		return err
	}

	wafID := d.Get("waf_id").(string)
	log.Printf("[INFO] updating configuration for WAF: %s", wafID)
	if latestVersion.Locked {
		latestVersion, err = conn.CloneWAFVersion(&gofastly.CloneWAFVersionInput{
			WAFID:            wafID,
			WAFVersionNumber: latestVersion.Number,
		})
		if err != nil {
			return err
		}
	}

	input := buildUpdateInput(d, latestVersion.ID, latestVersion.Number)
	if input.HasChanges() {
		latestVersion, err = conn.UpdateWAFVersion(input)
		if err != nil {
			return err
		}
	}

	if d.HasChange("rule") {
		if err := updateRules(d, meta, wafID, latestVersion.Number); err != nil {
			return err
		}
	}

	err = conn.DeployWAFVersion(&gofastly.DeployWAFVersionInput{
		WAFID:            wafID,
		WAFVersionNumber: latestVersion.Number,
	})
	if err != nil {
		return err
	}

	return resourceServiceWAFConfigurationV1Read(d, meta)
}

func resourceServiceWAFConfigurationV1Read(d *schema.ResourceData, meta interface{}) error {

	latestVersion, err := getLatestVersion(d, meta)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode == 404 {
			log.Printf("[DEBUG] WAF (%s) was not found - removing from state", d.Get("waf_id").(string))
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[INFO] retrieving WAF version number: %d", latestVersion.Number)
	refreshWAFConfig(d, latestVersion)

	if err := readWAFRules(meta, d, latestVersion.Number); err != nil {
		return err
	}

	return nil
}

func resourceServiceWAFConfigurationV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	wafID := d.Get("waf_id").(string)
	log.Printf("[INFO] destroying configuration by creating empty version of WAF: %s", wafID)
	emptyVersion, err := conn.CreateEmptyWAFVersion(&gofastly.CreateEmptyWAFVersionInput{
		WAFID: wafID,
	})
	if err != nil {
		return err
	}

	err = conn.DeployWAFVersion(&gofastly.DeployWAFVersionInput{
		WAFID:            wafID,
		WAFVersionNumber: emptyVersion.Number,
	})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceServiceWAFConfigurationV1Import(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

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
		input.WAFID = strToPtr(v.(string))
	}
	if v, ok := d.GetOk("allowed_http_versions"); ok {
		input.AllowedHTTPVersions = strToPtr(v.(string))
	}
	if v, ok := d.GetOk("allowed_methods"); ok {
		input.AllowedMethods = strToPtr(v.(string))
	}
	if v, ok := d.GetOk("allowed_request_content_type"); ok {
		input.AllowedRequestContentType = strToPtr(v.(string))
	}
	if v, ok := d.GetOk("allowed_request_content_type_charset"); ok {
		input.AllowedRequestContentTypeCharset = strToPtr(v.(string))
	}
	if v, ok := d.GetOk("arg_length"); ok {
		input.ArgLength = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("arg_name_length"); ok {
		input.ArgNameLength = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("combined_file_sizes"); ok {
		input.CombinedFileSizes = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("critical_anomaly_score"); ok {
		input.CriticalAnomalyScore = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("crs_validate_utf8_encoding"); ok {
		input.CRSValidateUTF8Encoding = boolToPtr(v.(bool))
	}
	if v, ok := d.GetOk("error_anomaly_score"); ok {
		input.ErrorAnomalyScore = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("high_risk_country_codes"); ok {
		input.HighRiskCountryCodes = strToPtr(v.(string))
	}
	if v, ok := d.GetOk("http_violation_score_threshold"); ok {
		input.HTTPViolationScoreThreshold = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("inbound_anomaly_score_threshold"); ok {
		input.InboundAnomalyScoreThreshold = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("lfi_score_threshold"); ok {
		input.LFIScoreThreshold = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("max_file_size"); ok {
		input.MaxFileSize = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("max_num_args"); ok {
		input.MaxNumArgs = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("notice_anomaly_score"); ok {
		input.NoticeAnomalyScore = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("paranoia_level"); ok {
		input.ParanoiaLevel = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("php_injection_score_threshold"); ok {
		input.PHPInjectionScoreThreshold = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("rce_score_threshold"); ok {
		input.RCEScoreThreshold = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("restricted_extensions"); ok {
		input.RestrictedExtensions = strToPtr(v.(string))
	}
	if v, ok := d.GetOk("restricted_headers"); ok {
		input.RestrictedHeaders = strToPtr(v.(string))
	}
	if v, ok := d.GetOk("rfi_score_threshold"); ok {
		input.RFIScoreThreshold = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("session_fixation_score_threshold"); ok {
		input.SessionFixationScoreThreshold = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("sql_injection_score_threshold"); ok {
		input.SQLInjectionScoreThreshold = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("total_arg_length"); ok {
		input.TotalArgLength = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("warning_anomaly_score"); ok {
		input.WarningAnomalyScore = intToPtr(v.(int))
	}
	if v, ok := d.GetOk("xss_score_threshold"); ok {
		input.XSSScoreThreshold = intToPtr(v.(int))
	}
	return input
}

func refreshWAFConfig(d *schema.ResourceData, version *gofastly.WAFVersion) {

	pairings := composePairings(version)
	for k, v := range pairings {
		var ok bool
		switch t := reflect.TypeOf(v).String(); t {
		case "string":
			if _, ok := d.GetOk(k); !ok || v.(string) == "" {
				continue
			}
		case "int":
			if _, ok := d.GetOk(k); !ok || v.(int) == 0 {
				continue
			}
		case "bool":
			if _, ok := d.GetOkExists(k); !ok {
				continue
			}
		}
		d.Set(k, v)
		log.Printf("[DEBUG] GetOk for %v is %v \n", k, ok)
	}
}

func composePairings(version *gofastly.WAFVersion) map[string]interface{} {
	return map[string]interface{}{
		"allowed_http_versions":                version.AllowedHTTPVersions,
		"allowed_methods":                      version.AllowedMethods,
		"allowed_request_content_type":         version.AllowedRequestContentType,
		"allowed_request_content_type_charset": version.AllowedRequestContentTypeCharset,
		"arg_length":                           version.ArgLength,
		"arg_name_length":                      version.ArgNameLength,
		"combined_file_sizes":                  version.CombinedFileSizes,
		"critical_anomaly_score":               version.CriticalAnomalyScore,
		"crs_validate_utf8_encoding":           version.CRSValidateUTF8Encoding,
		"error_anomaly_score":                  version.ErrorAnomalyScore,
		"high_risk_country_codes":              version.HighRiskCountryCodes,
		"http_violation_score_threshold":       version.HTTPViolationScoreThreshold,
		"inbound_anomaly_score_threshold":      version.InboundAnomalyScoreThreshold,
		"lfi_score_threshold":                  version.LFIScoreThreshold,
		"max_file_size":                        version.MaxFileSize,
		"max_num_args":                         version.MaxNumArgs,
		"notice_anomaly_score":                 version.NoticeAnomalyScore,
		"paranoia_level":                       version.ParanoiaLevel,
		"php_injection_score_threshold":        version.PHPInjectionScoreThreshold,
		"rce_score_threshold":                  version.RCEScoreThreshold,
		"restricted_extensions":                version.RestrictedExtensions,
		"restricted_headers":                   version.RestrictedHeaders,
		"rfi_score_threshold":                  version.RFIScoreThreshold,
		"session_fixation_score_threshold":     version.SessionFixationScoreThreshold,
		"sql_injection_score_threshold":        version.SQLInjectionScoreThreshold,
		"total_arg_length":                     version.TotalArgLength,
		"warning_anomaly_score":                version.WarningAnomalyScore,
		"xss_score_threshold":                  version.XSSScoreThreshold,
	}
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
