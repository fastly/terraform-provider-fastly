package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		"content_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Value of the `Content-Type` header sent with the request",
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
			Sensitive:        true,
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
func (h *HTTPSLoggingServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly HTTPS logging addition opts: %#v", opts)

	return createHTTPS(conn, opts)
}

// Read refreshes the resource.
func (h *HTTPSLoggingServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing HTTPS logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListHTTPS(&gofastly.ListHTTPSInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up HTTPS logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		hll := flattenHTTPS(remoteState)

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
func (h *HTTPSLoggingServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
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

	log.Printf("[DEBUG] Update HTTPS Opts: %#v", opts)
	_, err := conn.UpdateHTTPS(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *HTTPSLoggingServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly HTTPS logging endpoint removal opts: %#v", opts)

	return deleteHTTPS(conn, opts)
}

func createHTTPS(conn *gofastly.Client, i *gofastly.CreateHTTPSInput) error {
	_, err := conn.CreateHTTPS(i)
	if err != nil {
		return err
	}
	return nil
}

func deleteHTTPS(conn *gofastly.Client, i *gofastly.DeleteHTTPSInput) error {
	err := conn.DeleteHTTPS(i)

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
func flattenHTTPS(remoteState []*gofastly.HTTPS) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
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
		if resource.ContentType != nil {
			data["content_type"] = *resource.ContentType
		}
		if resource.HeaderName != nil {
			data["header_name"] = *resource.HeaderName
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
		ContentType:       gofastly.ToPointer(resource["content_type"].(string)),
		Format:            gofastly.ToPointer(vla.format),
		FormatVersion:     vla.formatVersion,
		HeaderName:        gofastly.ToPointer(resource["header_name"].(string)),
		HeaderValue:       gofastly.ToPointer(resource["header_value"].(string)),
		JSONFormat:        gofastly.ToPointer(resource["json_format"].(string)),
		MessageType:       gofastly.ToPointer(resource["message_type"].(string)),
		Method:            gofastly.ToPointer(resource["method"].(string)),
		Name:              gofastly.ToPointer(resource["name"].(string)),
		RequestMaxBytes:   gofastly.ToPointer(resource["request_max_bytes"].(int)),
		RequestMaxEntries: gofastly.ToPointer(resource["request_max_entries"].(int)),
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		TLSCACert:         gofastly.ToPointer(resource["tls_ca_cert"].(string)),
		TLSClientCert:     gofastly.ToPointer(resource["tls_client_cert"].(string)),
		TLSClientKey:      gofastly.ToPointer(resource["tls_client_key"].(string)),
		TLSHostname:       gofastly.ToPointer(resource["tls_hostname"].(string)),
		URL:               gofastly.ToPointer(resource["url"].(string)),
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
