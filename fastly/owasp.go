package fastly

import (
	"github.com/hashicorp/terraform/helper/schema"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

var owaspSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Importer: &schema.ResourceImporter{},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID assigned to this OWASP.",
			},
			"allowed_http_versions": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"allowed_methods": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"allowed_request_content_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arg_length": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"arg_name_length": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"combined_file_sizes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"critical_anomaly_score": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"crs_validate_utf8_encoding": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"error_anomaly_score": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"high_risk_country_codes": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"http_violation_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"inbound_anomaly_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"lfi_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"max_file_size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"max_num_args": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"notice_anomaly_score": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"paranoia_level": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"php_injection_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"rce_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"restricted_extensions": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"restricted_headers": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rfi_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"session_fixation_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"sql_injection_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"total_arg_length": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"warning_anomaly_score": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"xss_score_threshold": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	},
}

// expandOWASP converts a machine schema to a gofastly OWASP.
func expandOWASP(raw map[string]interface{}) *gofastly.OWASP {
	return &gofastly.OWASP{
		ID:                            raw["id"].(string),
		AllowedHTTPVersions:           raw["allowed_http_versions"].(string),
		AllowedMethods:                raw["allowed_methods"].(string),
		AllowedRequestContentType:     raw["allowed_request_content_type"].(string),
		ArgLength:                     raw["arg_length"].(int),
		ArgNameLength:                 raw["arg_name_length"].(int),
		CombinedFileSizes:             raw["combined_file_sizes"].(int),
		CreatedAt:                     raw["created_at"].(string),
		CriticalAnomalyScore:          raw["critical_anomaly_score"].(int),
		CRSValidateUTF8Encoding:       raw["crs_validate_utf8_encoding"].(bool),
		ErrorAnomalyScore:             raw["error_anomaly_score"].(int),
		HighRiskCountryCodes:          raw["high_risk_country_codes"].(string),
		HTTPViolationScoreThreshold:   raw["http_violation_score_threshold"].(int),
		InboundAnomalyScoreThreshold:  raw["inbound_anomaly_score_threshold"].(int),
		LFIScoreThreshold:             raw["lfi_score_threshold"].(int),
		MaxFileSize:                   raw["max_file_size"].(int),
		MaxNumArgs:                    raw["max_num_args"].(int),
		NoticeAnomalyScore:            raw["notice_anomaly_score"].(int),
		ParanoiaLevel:                 raw["paranoia_level"].(int),
		PHPInjectionScoreThreshold:    raw["php_injection_score_threshold"].(int),
		RCEScoreThreshold:             raw["rce_score_threshold"].(int),
		RestrictedExtensions:          raw["restricted_extensions"].(string),
		RestrictedHeaders:             raw["restricted_headers"].(string),
		RFIScoreThreshold:             raw["rfi_score_threshold"].(int),
		SessionFixationScoreThreshold: raw["session_fixation_score_threshold"].(int),
		SQLInjectionScoreThreshold:    raw["sql_injection_score_threshold"].(int),
		TotalArgLength:                raw["total_arg_length"].(int),
		UpdatedAt:                     raw["updated_at"].(string),
		WarningAnomalyScore:           raw["warning_anomaly_score"].(int),
		XDDScoreThreshold:             raw["xss_score_threshold"].(int),
	}
}

// flattenOWASP converts an OWASP object to map[string]interface for inclusion into a WAF map.
func flattenOWASP(owasp *gofastly.OWASP) map[string]interface{} {
	return map[string]interface{}{
		"id": owasp.ID,
		"allowed_http_versions":            owasp.AllowedHTTPVersions,
		"allowed_methods":                  owasp.AllowedMethods,
		"allowed_request_content_type":     owasp.AllowedRequestContentType,
		"arg_length":                       owasp.ArgLength,
		"arg_name_length":                  owasp.ArgNameLength,
		"combined_file_sizes":              owasp.CombinedFileSizes,
		"created_at":                       owasp.CreatedAt,
		"critical_anomaly_score":           owasp.CriticalAnomalyScore,
		"crs_validate_utf8_encoding":       owasp.CRSValidateUTF8Encoding,
		"error_anomaly_score":              owasp.ErrorAnomalyScore,
		"high_risk_country_codes":          owasp.HighRiskCountryCodes,
		"http_violation_score_threshold":   owasp.HTTPViolationScoreThreshold,
		"inbound_anomaly_score_threshold":  owasp.InboundAnomalyScoreThreshold,
		"lfi_score_threshold":              owasp.LFIScoreThreshold,
		"max_file_size":                    owasp.MaxFileSize,
		"max_num_args":                     owasp.MaxNumArgs,
		"notice_anomaly_score":             owasp.NoticeAnomalyScore,
		"paranoia_level":                   owasp.ParanoiaLevel,
		"php_injection_score_threshold":    owasp.PHPInjectionScoreThreshold,
		"rce_score_threshold":              owasp.RCEScoreThreshold,
		"restricted_extensions":            owasp.RestrictedExtensions,
		"restricted_headers":               owasp.RestrictedHeaders,
		"rfi_score_threshold":              owasp.RFIScoreThreshold,
		"session_fixation_score_threshold": owasp.SessionFixationScoreThreshold,
		"sql_injection_score_threshold":    owasp.SQLInjectionScoreThreshold,
		"total_arg_length":                 owasp.TotalArgLength,
		"updated_at":                       owasp.UpdatedAt,
		"warning_anomaly_score":            owasp.WarningAnomalyScore,
		"xss_score_threshold":              owasp.XDDScoreThreshold,
	}
}

func updateOWASP(client *gofastly.Client, service string, wafID string, owasp *gofastly.OWASP) (*gofastly.OWASP, error) {
	input := &gofastly.UpdateOWASPInput{
		Service: service,
		ID:      wafID,
		OWASPID: owasp.ID,
		Type:    "owasp",

		AllowedHTTPVersions:           owasp.AllowedHTTPVersions,
		AllowedMethods:                owasp.AllowedMethods,
		AllowedRequestContentType:     owasp.AllowedRequestContentType,
		ArgLength:                     owasp.ArgLength,
		ArgNameLength:                 owasp.ArgNameLength,
		CombinedFileSizes:             owasp.CombinedFileSizes,
		CreatedAt:                     owasp.CreatedAt,
		CriticalAnomalyScore:          owasp.CriticalAnomalyScore,
		CRSValidateUTF8Encoding:       owasp.CRSValidateUTF8Encoding,
		ErrorAnomalyScore:             owasp.ErrorAnomalyScore,
		HighRiskCountryCodes:          owasp.HighRiskCountryCodes,
		HTTPViolationScoreThreshold:   owasp.HTTPViolationScoreThreshold,
		InboundAnomalyScoreThreshold:  owasp.InboundAnomalyScoreThreshold,
		LFIScoreThreshold:             owasp.LFIScoreThreshold,
		MaxFileSize:                   owasp.MaxFileSize,
		MaxNumArgs:                    owasp.MaxNumArgs,
		NoticeAnomalyScore:            owasp.NoticeAnomalyScore,
		ParanoiaLevel:                 owasp.ParanoiaLevel,
		PHPInjectionScoreThreshold:    owasp.PHPInjectionScoreThreshold,
		RCEScoreThreshold:             owasp.RCEScoreThreshold,
		RestrictedExtensions:          owasp.RestrictedExtensions,
		RestrictedHeaders:             owasp.RestrictedHeaders,
		RFIScoreThreshold:             owasp.RFIScoreThreshold,
		SessionFixationScoreThreshold: owasp.SessionFixationScoreThreshold,
		SQLInjectionScoreThreshold:    owasp.SQLInjectionScoreThreshold,
		TotalArgLength:                owasp.TotalArgLength,
		UpdatedAt:                     owasp.UpdatedAt,
		WarningAnomalyScore:           owasp.WarningAnomalyScore,
		XDDScoreThreshold:             owasp.XDDScoreThreshold,
	}

	return client.UpdateOWASP(input)
}
