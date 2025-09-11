package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

// HTTPSLoggingServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type HTTPSLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingHTTPS returns a new resource.
func NewServiceLoggingHTTPS(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&HTTPSLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_https",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *HTTPSLoggingServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *HTTPSLoggingServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
		},
		"content_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Value of the `Content-Type` header sent with the request",
		},
		"gzip_level": {
			Type:     schema.TypeInt,
			Optional: true,
			// NOTE: The default represents an unset value
			// We use this instead of zero because the zero value for an int type is
			// actually a valid value for the API. The API will attempt to default to
			// zero if nothing is set by the user in their TF configuration.
			Default:     -1,
			Description: GzipLevelDescription,
		},
		"header_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Custom header sent with the request",
		},
		"header_value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Value of the custom header sent with the request",
		},
		// NOTE: The `json_format` field's documented type is string, but it should likely be an integer.
		"json_format": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "0",
			Description:  "Formats log entries as JSON. Can be either disabled (`0`), array of json (`1`), or newline delimited json (`2`)",
			ValidateFunc: validation.StringInSlice([]string{"0", "1", "2"}, false),
		},
		"message_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "blank",
			Description:      MessageTypeDescription,
			ValidateDiagFunc: validateLoggingMessageType(),
		},
		"method": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "POST",
			Description:  "HTTP method used for request. Can be either `POST` or `PUT`. Default `POST`",
			ValidateFunc: validation.StringInSlice([]string{"POST", "PUT"}, false),
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the HTTPS logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently the logs should be transferred, in seconds (Default 3600)",
		},
		"processing_region": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "none",
			Description:  "Region where logs will be processed before streaming to BigQuery. Valid values are 'none', 'us' and 'eu'.",
			ValidateFunc: validation.StringInSlice([]string{"none", "us", "eu"}, false),
		},
		"request_max_bytes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of bytes sent in one request",
		},
		"request_max_entries": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of logs sent in one request",
		},
		"tls_ca_cert": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "A secure certificate to authenticate the server with. Must be in PEM format",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"tls_client_cert": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The client certificate used to make authenticated requests. Must be in PEM format",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"tls_client_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The client private key used to make authenticated requests. Must be in PEM format",
			Sensitive:        !DisplaySensitiveFields,
			ValidateDiagFunc: validateStringTrimmed,
		},
		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Used during the TLS handshake to validate the certificate",
		},
		"url": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "URL that log data will be sent to. Must use the https protocol",
			ValidateFunc: validation.IsURLWithHTTPS,
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     LoggingHTTPSDefaultFormat,
			Description: "Apache-style string or VCL variables to use for log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed",
			ValidateDiagFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply",
		}
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
func (h *HTTPSLoggingServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly HTTPS logging addition opts: %#v", opts)

	_, err := conn.CreateHTTPS(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *HTTPSLoggingServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing HTTPS logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListHTTPS(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListHTTPSInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up HTTPS logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		hll := flattenHTTPS(remoteState, localState)

		for _, element := range hll {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), hll); err != nil {
			log.Printf("[WARN] Error setting HTTPS logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *HTTPSLoggingServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateHTTPSInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["request_max_entries"]; ok {
		opts.RequestMaxEntries = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["request_max_bytes"]; ok {
		opts.RequestMaxBytes = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["content_type"]; ok {
		opts.ContentType = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		opts.GzipLevel = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["header_name"]; ok {
		opts.HeaderName = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["header_value"]; ok {
		opts.HeaderValue = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["method"]; ok {
		opts.Method = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["json_format"]; ok {
		opts.JSONFormat = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_ca_cert"]; ok {
		opts.TLSCACert = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_client_cert"]; ok {
		opts.TLSClientCert = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_client_key"]; ok {
		opts.TLSClientKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_hostname"]; ok {
		opts.TLSHostname = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update HTTPS Opts: %#v", opts)
	_, err := conn.UpdateHTTPS(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *HTTPSLoggingServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly HTTPS logging endpoint removal opts: %#v", opts)

	err := conn.DeleteHTTPS(gofastly.NewContextForResourceID(ctx, d.Id()), opts)

	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// flattenHTTPS models data into format suitable for saving to Terraform state.
func flattenHTTPS(remoteState []*gofastly.HTTPS, localState []any) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		// Avoid setting gzip_level to the API default of zero if originally unset.
		// This avoids an unnecessary diff where the local state would have been
		// updated to zero and so would be different from the -1 default set.
		// As the user never set the attribute we don't want to show a diff to say
		// it should be zero according to the API.
		//
		// NOTE: Ideally the local state would be updated when .Create() is called.
		// e.g. we'd check if the value is -1 for gzip_level and set it in state as
		// zero instead. This way we could avoid having to do this check here.
		// The reason that's not possible (or not ideal at least) is because Create
		// is called multiple times (once for each block defined in configuration)
		// while the setting of the state must be done holistically, and so what
		// that means is, if we did the above suggestion we would be resetting the
		// entire state object multiple times, where as here we're only ever setting
		// it once.
		for _, s := range localState {
			v := s.(map[string]any)
			if resource.Name != nil && v["name"].(string) == *resource.Name && v["gzip_level"].(int) == -1 {
				resource.GzipLevel = gofastly.ToPointer(v["gzip_level"].(int))
				break
			}
		}

		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.URL != nil {
			data["url"] = *resource.URL
		}
		if resource.RequestMaxEntries != nil {
			data["request_max_entries"] = *resource.RequestMaxEntries
		}
		if resource.RequestMaxBytes != nil {
			data["request_max_bytes"] = *resource.RequestMaxBytes
		}
		if resource.CompressionCodec != nil {
			data["compression_codec"] = *resource.CompressionCodec
		}
		if resource.ContentType != nil {
			data["content_type"] = *resource.ContentType
		}
		if resource.HeaderName != nil {
			data["header_name"] = *resource.HeaderName
		}
		if resource.GzipLevel != nil {
			data["gzip_level"] = *resource.GzipLevel
		}
		if resource.HeaderValue != nil {
			data["header_value"] = *resource.HeaderValue
		}
		if resource.Method != nil {
			data["method"] = *resource.Method
		}
		if resource.JSONFormat != nil {
			data["json_format"] = *resource.JSONFormat
		}
		if resource.Period != nil {
			data["period"] = *resource.Period
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.TLSCACert != nil {
			data["tls_ca_cert"] = *resource.TLSCACert
		}
		if resource.TLSClientCert != nil {
			data["tls_client_cert"] = *resource.TLSClientCert
		}
		if resource.TLSClientKey != nil {
			data["tls_client_key"] = *resource.TLSClientKey
		}
		if resource.TLSHostname != nil {
			data["tls_hostname"] = *resource.TLSHostname
		}
		if resource.MessageType != nil {
			data["message_type"] = *resource.MessageType
		}
		if resource.FormatVersion != nil {
			data["format_version"] = *resource.FormatVersion
		}
		if resource.ProcessingRegion != nil {
			data["processing_region"] = *resource.ProcessingRegion
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
}

func (h *HTTPSLoggingServiceAttributeHandler) buildCreate(httpsMap any, serviceID string, serviceVersion int) *gofastly.CreateHTTPSInput {
	resource := httpsMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateHTTPSInput{
		CompressionCodec:  gofastly.ToPointer(resource["compression_codec"].(string)),
		ContentType:       gofastly.ToPointer(resource["content_type"].(string)),
		Format:            gofastly.ToPointer(vla.format),
		FormatVersion:     vla.formatVersion,
		HeaderName:        gofastly.ToPointer(resource["header_name"].(string)),
		HeaderValue:       gofastly.ToPointer(resource["header_value"].(string)),
		JSONFormat:        gofastly.ToPointer(resource["json_format"].(string)),
		MessageType:       gofastly.ToPointer(resource["message_type"].(string)),
		Method:            gofastly.ToPointer(resource["method"].(string)),
		Name:              gofastly.ToPointer(resource["name"].(string)),
		Period:            gofastly.ToPointer(resource["period"].(int)),
		RequestMaxBytes:   gofastly.ToPointer(resource["request_max_bytes"].(int)),
		RequestMaxEntries: gofastly.ToPointer(resource["request_max_entries"].(int)),
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		TLSCACert:         gofastly.ToPointer(resource["tls_ca_cert"].(string)),
		TLSClientCert:     gofastly.ToPointer(resource["tls_client_cert"].(string)),
		TLSClientKey:      gofastly.ToPointer(resource["tls_client_key"].(string)),
		TLSHostname:       gofastly.ToPointer(resource["tls_hostname"].(string)),
		URL:               gofastly.ToPointer(resource["url"].(string)),
		ProcessingRegion:  gofastly.ToPointer(resource["processing_region"].(string)),
	}

	// NOTE: go-fastly v7+ expects a pointer, so TF can't set the zero type value.
	// If we set a default value for an attribute, then it will be sent to the API.
	// In some scenarios this can cause the API to reject the request.
	// For example, configuring compression_codec + gzip_level is invalid.
	if gl, ok := resource["gzip_level"].(int); ok && gl != -1 {
		opts.GzipLevel = gofastly.ToPointer(gl)
	}

	if p, ok := resource["period"].(int); ok && p != 0 {
		opts.Period = gofastly.ToPointer(p)
	}
	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.placement != "" {
		opts.Placement = gofastly.ToPointer(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.ToPointer(vla.responseCondition)
	}

	return &opts
}

func (h *HTTPSLoggingServiceAttributeHandler) buildDelete(httpsMap any, serviceID string, serviceVersion int) *gofastly.DeleteHTTPSInput {
	resource := httpsMap.(map[string]any)

	opts := gofastly.DeleteHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	return &opts
}
