package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// RateLimiterAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type RateLimiterAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceRateLimiter returns a new resource.
func NewServiceRateLimiter(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&RateLimiterAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "rate_limiter",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *RateLimiterAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
//
// API REFERENCE:
// https://developer.fastly.com/reference/api/vcl-services/rate-limiter/
func (h *RateLimiterAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"action": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The action to take when a rate limiter violation is detected (one of: log_only, response, response_object)",
			ValidateDiagFunc: func(i any, _ cty.Path) diag.Diagnostics {
				for _, a := range gofastly.ERLActions {
					if i.(string) == string(a) {
						return nil
					}
				}
				return diag.Errorf("invalid action, should be one of: %+v", gofastly.ERLActions)
			},
		},
		"client_key": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Comma-separated list of VCL variables used to generate a counter key to identify a client",
			ValidateDiagFunc: func(i any, _ cty.Path) diag.Diagnostics {
				if strings.Contains(i.(string), " ") {
					return diag.Errorf("invalid client_key: should contain no spaces")
				}
				return nil
			},
		},
		"feature_revision": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
			Description: "Revision number of the rate limiting feature implementation",
		},
		"http_methods": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Comma-separated list of HTTP methods to apply rate limiting to",
			ValidateDiagFunc: func(i any, _ cty.Path) diag.Diagnostics {
				v := i.(string)
				if strings.Contains(v, " ") {
					return diag.Errorf("invalid http_methods: should contain no spaces")
				}
				if v != strings.ToUpper(v) {
					return diag.Errorf("invalid http_methods: each method should be UPPERCASE (e.g. POST,PUT,PATCH,DELETE)")
				}
				return nil
			},
		},
		"logger_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the type of logging endpoint to be used when action is log_only (one of: azureblob, bigquery, cloudfiles, datadog, digitalocean, elasticsearch, ftp, gcs, googleanalytics, heroku, honeycomb, http, https, kafka, kinesis, logentries, loggly, logshuttle, newrelic, openstack, papertrail, pubsub, s3, scalyr, sftp, splunk, stackdriver, sumologic, syslog)",
			ValidateDiagFunc: func(i any, _ cty.Path) diag.Diagnostics {
				for _, l := range gofastly.ERLLoggers {
					if i.(string) == string(l) {
						return nil
					}
				}
				return diag.Errorf("invalid logger_type, should be one of: %+v", gofastly.ERLLoggers)
			},
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A unique human readable name for the rate limiting rule",
		},
		"penalty_box_duration": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Length of time in minutes that the rate limiter is in effect after the initial violation is detected",
		},
		"ratelimiter_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Alphanumeric string identifying the rate limiter",
		},
		"response": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Custom response to be sent when the rate limit is exceeded. Required if action is response",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"content": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "HTTP response body data",
					},
					"content_type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "HTTP Content-Type (e.g. application/json)",
					},
					"status": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "HTTP response status code (e.g. 429)",
					},
				},
			},
		},
		"response_object_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of existing response object. Required if action is response_object",
		},
		"rps_limit": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Upper limit of requests per second allowed by the rate limiter",
		},
		"uri_dictionary_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an Edge Dictionary containing URIs as keys. If not defined or null, all origin URIs will be rate limited",
		},
		"window_size": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Number of seconds during which the RPS limit must be exceeded in order to trigger a violation (one of: 1, 10, 60)",
			ValidateDiagFunc: func(i any, _ cty.Path) diag.Diagnostics {
				for _, w := range gofastly.ERLWindowSizes {
					if i.(int) == int(w) {
						return nil
					}
				}
				return diag.Errorf("invalid window_size, should be one of: %+v", gofastly.ERLWindowSizes)
			},
		},
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
}

// Create creates the resource.
func (h *RateLimiterAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreateERLInput(d.Id(), serviceVersion, resource)

	log.Printf("[DEBUG] Create Rate Limiter: %#v", opts)
	_, err := conn.CreateERL(&opts)
	if err != nil {
		return err
	}

	return nil
}

// Read refreshes the resource.
func (h *RateLimiterAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Rate Limiters for (%s)", d.Id())
		remoteState, err := conn.ListERLs(&gofastly.ListERLsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Rate Limiters for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		data := flattenRateLimiter(remoteState, h.GetServiceMetadata())
		if err := d.Set(h.GetKey(), data); err != nil {
			log.Printf("[WARN] Error setting Rate Limiters for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *RateLimiterAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	var rateLimiterID string

	// IMPORTANT: Cloning a Service will result in new Rate Limiter IDs.
	//
	// This means, to update a Rate Limiter we have to first get a list of all
	// available Rate Limiters on the cloned service version, then identify the
	// one we need by its 'name' (which the API doesn't treat as unique, so
	// multiple Rate Limiters in theory can have the same name, but in Terraform
	// we enforce that the name must be unique otherwise we can't safely determine
	// the Rate Limiter ID).

	erls, err := conn.ListERLs(&gofastly.ListERLsInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
	})
	if err != nil {
		return err
	}

	for _, e := range erls {
		if e.Name == resource["name"].(string) {
			rateLimiterID = e.ID
			break
		}
	}

	input := h.buildUpdateERLInput(rateLimiterID, d.Id(), serviceVersion, resource, modified)

	log.Printf("[DEBUG] Update Rate Limiter: %#v", input)
	_, err = conn.UpdateERL(&input)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *RateLimiterAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	var rateLimiterID string

	// IMPORTANT: Cloning a Service will result in new Rate Limiter IDs.
	//
	// This means, to update a Rate Limiter we have to first get a list of all
	// available Rate Limiters on the cloned service version, then identify the
	// one we need by its 'name' (which the API doesn't treat as unique, so
	// multiple Rate Limiters in theory can have the same name, but in Terraform
	// we enforce that the name must be unique otherwise we can't safely determine
	// the Rate Limiter ID).

	erls, err := conn.ListERLs(&gofastly.ListERLsInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
	})
	if err != nil {
		return err
	}

	for _, e := range erls {
		if e.Name == resource["name"].(string) {
			rateLimiterID = e.ID
			break
		}
	}
	input := h.createDeleteERLInput(rateLimiterID)

	log.Printf("[DEBUG] Delete Rate Limiter: %#v", input)
	err = conn.DeleteERL(&input)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (h *RateLimiterAttributeHandler) createDeleteERLInput(rateLimiterID string) gofastly.DeleteERLInput {
	return gofastly.DeleteERLInput{
		ERLID: rateLimiterID,
	}
}

func (h *RateLimiterAttributeHandler) buildCreateERLInput(service string, latestVersion int, resource map[string]any) gofastly.CreateERLInput {
	input := gofastly.CreateERLInput{
		Name:               gofastly.ToPointer(resource["name"].(string)),
		PenaltyBoxDuration: gofastly.ToPointer(resource["penalty_box_duration"].(int)),
		RpsLimit:           gofastly.ToPointer(resource["rps_limit"].(int)),
		ServiceID:          service,
		ServiceVersion:     latestVersion,
	}

	action := resource["action"].(string)
	for _, a := range gofastly.ERLActions {
		if action == string(a) {
			input.Action = gofastly.ToPointer(a)
			break
		}
	}

	clientKeys := strings.Split(strings.ReplaceAll(resource["client_key"].(string), " ", ""), ",")
	input.ClientKey = &clientKeys

	featRevision := resource["feature_revision"].(int)
	if featRevision > 0 {
		input.FeatureRevision = gofastly.ToPointer(featRevision)
	}

	httpMethods := strings.Split(strings.ReplaceAll(resource["http_methods"].(string), " ", ""), ",")
	input.HTTPMethods = &httpMethods

	loggerType := resource["logger_type"].(string)
	for _, l := range gofastly.ERLLoggers {
		if loggerType == string(l) {
			input.LoggerType = gofastly.ToPointer(l)
			break
		}
	}

	response := resource["response"].([]any)
	if len(response) > 0 {
		for _, v := range response {
			m := v.(map[string]any)
			input.Response = &gofastly.ERLResponseType{
				ERLContent:     m["content"].(string),
				ERLContentType: m["content_type"].(string),
				ERLStatus:      int(m["status"].(int)),
			}
		}
	}

	respObjName := resource["response_object_name"].(string)
	if respObjName != "" {
		input.ResponseObjectName = gofastly.ToPointer(respObjName)
	}

	uriDictName := resource["uri_dictionary_name"].(string)
	if uriDictName != "" {
		input.URIDictionaryName = gofastly.ToPointer(uriDictName)
	}

	windowSize := resource["window_size"].(int)
	for _, w := range gofastly.ERLWindowSizes {
		if windowSize == int(w) {
			input.WindowSize = gofastly.ToPointer(w)
			break
		}
	}

	return input
}

func (h *RateLimiterAttributeHandler) buildUpdateERLInput(rateLimiterID, serviceID string, latestVersion int, resource, modified map[string]any) gofastly.UpdateERLInput {
	input := gofastly.UpdateERLInput{
		ERLID: rateLimiterID,
	}

	// NOTE: When converting from an `any` type, we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.

	if v, ok := modified["action"]; ok {
		for _, a := range gofastly.ERLActions {
			if v.(string) == string(a) {
				input.Action = gofastly.ToPointer(a)
				break
			}
		}
	}

	if v, ok := modified["client_key"]; ok {
		clientKeys := strings.Split(strings.ReplaceAll(v.(string), " ", ""), ",")
		input.ClientKey = &clientKeys
	}

	if v, ok := modified["feature_revision"]; ok {
		input.FeatureRevision = gofastly.ToPointer(v.(int))
	}

	if v, ok := modified["http_methods"]; ok {
		httpMethods := strings.Split(strings.ReplaceAll(v.(string), " ", ""), ",")
		input.HTTPMethods = &httpMethods
	}

	if v, ok := modified["logger_type"]; ok {
		for _, l := range gofastly.ERLLoggers {
			if v.(string) == string(l) {
				input.LoggerType = gofastly.ToPointer(l)
				break
			}
		}
	}

	if v, ok := modified["name"]; ok {
		input.Name = gofastly.ToPointer(v.(string))
	}

	if v, ok := modified["penalty_box_duration"]; ok {
		input.PenaltyBoxDuration = gofastly.ToPointer(v.(int))
	}

	if v, ok := modified["response"]; ok {
		s := v.([]any)
		if len(s) > 0 {
			m := s[0].(map[string]any)
			input.Response = &gofastly.ERLResponseType{
				ERLContent:     m["content"].(string),
				ERLContentType: m["content_type"].(string),
				ERLStatus:      m["status"].(int),
			}
		}
	}

	if v, ok := modified["response_object_name"]; ok {
		input.ResponseObjectName = gofastly.ToPointer(v.(string))
	}

	if v, ok := modified["rps_limit"]; ok {
		input.RpsLimit = gofastly.ToPointer(v.(int))
	}

	if v, ok := modified["uri_dictionary_name"]; ok {
		input.URIDictionaryName = gofastly.ToPointer(v.(string))
	}

	if v, ok := modified["window_size"]; ok {
		for _, w := range gofastly.ERLWindowSizes {
			if v.(int) == int(w) {
				input.WindowSize = gofastly.ToPointer(w)
				break
			}
		}
	}

	return input
}

// flattenRateLimiter models data into format suitable for saving to Terraform state.
func flattenRateLimiter(remoteState []*gofastly.ERL, _ ServiceMetadata) []map[string]any {
	result := make([]map[string]any, 0, len(remoteState))

	for _, o := range remoteState {
		data := map[string]any{
			"action":               string(o.Action),
			"client_key":           strings.Join(o.ClientKey, ","),
			"feature_revision":     o.FeatureRevision,
			"http_methods":         strings.Join(o.HTTPMethods, ","),
			"logger_type":          string(o.LoggerType),
			"name":                 o.Name,
			"penalty_box_duration": o.PenaltyBoxDuration,
			"ratelimiter_id":       o.ID,
			"response_object_name": o.ResponseObjectName,
			"rps_limit":            o.RpsLimit,
			"window_size":          int(o.WindowSize),
		}

		if o.Response != nil {
			data["response"] = []map[string]any{
				{
					"content":      o.Response.ERLContent,
					"content_type": o.Response.ERLContentType,
					"status":       o.Response.ERLStatus,
				},
			}
		}

		result = append(result, data)
	}

	return result
}
