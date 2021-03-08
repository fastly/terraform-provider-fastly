package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type HTTPSLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceHTTPSLogging(sa ServiceMetadata) ServiceAttributeDefinition {
	return &HTTPSLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "httpslogging",
			serviceMetadata: sa,
		},
	}
}

func (h *HTTPSLoggingServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	oh, nh := d.GetChange(h.GetKey())

	if oh == nil {
		oh = new(schema.Set)
	}
	if nh == nil {
		nh = new(schema.Set)
	}

	oldSet := oh.(*schema.Set)
	newSet := nh.(*schema.Set)

	setDiff := NewSetDiff(func(resource interface{}) (interface{}, error) {
		t, ok := resource.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource failed to be type asserted: %+v", resource)
		}
		return t["name"], nil
	})

	diffResult, err := setDiff.Diff(oldSet, newSet)
	if err != nil {
		return err
	}

	// DELETE removed resources
	for _, resource := range diffResult.Deleted {
		resource := resource.(map[string]interface{})
		opts := h.buildDelete(resource, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly HTTPS logging endpoint removal opts: %#v", opts)

		if err := deleteHTTPS(conn, opts); err != nil {
			return err
		}
	}

	// CREATE new resources
	for _, nRaw := range diffResult.Added {
		hf := nRaw.(map[string]interface{})
		opts := h.buildCreate(hf, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly HTTPS logging addition opts: %#v", opts)

		if err := createHTTPS(conn, opts); err != nil {
			return err
		}
	}

	// UPDATE modified resources
	//
	// NOTE: although the go-fastly API client enables updating of a resource by
	// its 'name' attribute, this isn't possible within terraform due to
	// constraints in the data model/schema of the resources not having a uid.
	for _, resource := range diffResult.Modified {
		resource := resource.(map[string]interface{})

		opts := gofastly.UpdateHTTPSInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
		}

		// only attempt to update attributes that have changed
		modified := setDiff.Filter(resource, oldSet)

		// NOTE: where we transition between interface{} we lose the ability to
		// infer the underlying type being either a uint vs an int. This
		// materializes as a panic (yay) and so it's only at runtime we discover
		// this and so we've updated the below code to convert the type asserted
		// int into a uint before passing the value to gofastly.Uint().
		if v, ok := modified["response_condition"]; ok {
			opts.ResponseCondition = gofastly.String(v.(string))
		}
		if v, ok := modified["format"]; ok {
			opts.Format = gofastly.String(v.(string))
		}
		if v, ok := modified["url"]; ok {
			opts.URL = gofastly.String(v.(string))
		}
		if v, ok := modified["request_max_entries"]; ok {
			opts.RequestMaxEntries = gofastly.Uint(uint(v.(int)))
		}
		if v, ok := modified["request_max_bytes"]; ok {
			opts.RequestMaxBytes = gofastly.Uint(uint(v.(int)))
		}
		if v, ok := modified["content_type"]; ok {
			opts.ContentType = gofastly.String(v.(string))
		}
		if v, ok := modified["header_name"]; ok {
			opts.HeaderName = gofastly.String(v.(string))
		}
		if v, ok := modified["header_value"]; ok {
			opts.HeaderValue = gofastly.String(v.(string))
		}
		if v, ok := modified["method"]; ok {
			opts.Method = gofastly.String(v.(string))
		}
		if v, ok := modified["json_format"]; ok {
			opts.JSONFormat = gofastly.String(v.(string))
		}
		if v, ok := modified["placement"]; ok {
			opts.Placement = gofastly.String(v.(string))
		}
		if v, ok := modified["tls_ca_cert"]; ok {
			opts.TLSCACert = gofastly.String(v.(string))
		}
		if v, ok := modified["tls_client_cert"]; ok {
			opts.TLSClientCert = gofastly.String(v.(string))
		}
		if v, ok := modified["tls_client_key"]; ok {
			opts.TLSClientKey = gofastly.String(v.(string))
		}
		if v, ok := modified["tls_hostname"]; ok {
			opts.TLSHostname = gofastly.String(v.(string))
		}
		if v, ok := modified["message_type"]; ok {
			opts.MessageType = gofastly.String(v.(string))
		}
		if v, ok := modified["format_version"]; ok {
			opts.FormatVersion = gofastly.Uint(uint(v.(int)))
		}

		log.Printf("[DEBUG] Update HTTPS Opts: %#v", opts)
		_, err := conn.UpdateHTTPS(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *HTTPSLoggingServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// refresh HTTPS
	log.Printf("[DEBUG] Refreshing HTTPS logging endpoints for (%s)", d.Id())
	httpsList, err := conn.ListHTTPS(&gofastly.ListHTTPSInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up HTTPS logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	hll := flattenHTTPS(httpsList)

	for _, element := range hll {
		element = h.pruneVCLLoggingAttributes(element)
	}

	if err := d.Set(h.GetKey(), hll); err != nil {
		log.Printf("[WARN] Error setting HTTPS logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func (h *HTTPSLoggingServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the HTTPS logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"url": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "URL that log data will be sent to. Must use the https protocol",
			ValidateFunc: validateHTTPSURL(),
		},

		// Optional fields
		"request_max_entries": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of logs sent in one request",
		},

		"request_max_bytes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of bytes sent in one request",
		},

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

		"method": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "POST",
			Description:  "HTTP method used for request. Can be either `POST` or `PUT`. Default `POST`",
			ValidateFunc: validation.StringInSlice([]string{"POST", "PUT"}, false),
		},

		// NOTE: The `json_format` field's documented type is string, but it should likely be an integer.
		"json_format": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "0",
			Description:  "Formats log entries as JSON. Can be either disabled (`0`), array of json (`1`), or newline delimited json (`2`)",
			ValidateFunc: validation.StringInSlice([]string{"0", "1", "2"}, false),
		},

		"tls_ca_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A secure certificate to authenticate the server with. Must be in PEM format",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_client_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The client certificate used to make authenticated requests. Must be in PEM format",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The client private key used to make authenticated requests. Must be in PEM format",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Used during the TLS handshake to validate the certificate",
		},

		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "blank",
			Description:  "How the message should be formatted; one of: `classic`, `loggly`, `logplex` or `blank`. Default `blank`",
			ValidateFunc: validateLoggingMessageType(),
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache-style string or VCL variables to use for log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed",
			ValidateFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply",
		}
	}

	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
	return nil
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

func flattenHTTPS(httpsList []*gofastly.HTTPS) []map[string]interface{} {
	var hsl []map[string]interface{}
	for _, hl := range httpsList {
		// Convert HTTP logging to a map for saving to state.
		nhl := map[string]interface{}{
			"name":                hl.Name,
			"response_condition":  hl.ResponseCondition,
			"format":              hl.Format,
			"url":                 hl.URL,
			"request_max_entries": hl.RequestMaxEntries,
			"request_max_bytes":   hl.RequestMaxBytes,
			"content_type":        hl.ContentType,
			"header_name":         hl.HeaderName,
			"header_value":        hl.HeaderValue,
			"method":              hl.Method,
			"json_format":         hl.JSONFormat,
			"placement":           hl.Placement,
			"tls_ca_cert":         hl.TLSCACert,
			"tls_client_cert":     hl.TLSClientCert,
			"tls_client_key":      hl.TLSClientKey,
			"tls_hostname":        hl.TLSHostname,
			"message_type":        hl.MessageType,
			"format_version":      hl.FormatVersion,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nhl {
			if v == "" {
				delete(nhl, k)
			}
		}

		hsl = append(hsl, nhl)
	}

	return hsl
}

func (h *HTTPSLoggingServiceAttributeHandler) buildCreate(httpsMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateHTTPSInput {
	df := httpsMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	opts := gofastly.CreateHTTPSInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              df["name"].(string),
		URL:               df["url"].(string),
		RequestMaxEntries: uint(df["request_max_entries"].(int)),
		RequestMaxBytes:   uint(df["request_max_bytes"].(int)),
		ContentType:       df["content_type"].(string),
		HeaderName:        df["header_name"].(string),
		HeaderValue:       df["header_value"].(string),
		Method:            df["method"].(string),
		JSONFormat:        df["json_format"].(string),
		TLSCACert:         df["tls_ca_cert"].(string),
		TLSClientCert:     df["tls_client_cert"].(string),
		TLSClientKey:      df["tls_client_key"].(string),
		TLSHostname:       df["tls_hostname"].(string),
		MessageType:       df["message_type"].(string),
		Format:            vla.format,
		FormatVersion:     uintOrDefault(vla.formatVersion),
		ResponseCondition: vla.responseCondition,
		Placement:         vla.placement,
	}

	return &opts
}

func (h *HTTPSLoggingServiceAttributeHandler) buildDelete(httpsMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteHTTPSInput {
	df := httpsMap.(map[string]interface{})

	opts := gofastly.DeleteHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}

	return &opts
}
