package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SplunkServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type SplunkServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingSplunk returns a new resource.
func NewServiceLoggingSplunk(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&SplunkServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_splunk",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *SplunkServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *SplunkServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A unique name to identify the Splunk endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"tls_ca_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SPLUNK_CA_CERT", ""),
			Description: "A secure certificate to authenticate the server with. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SPLUNK_CA_CERT`",
		},
		"tls_client_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SPLUNK_CLIENT_CERT", ""),
			Description: "The client certificate used to make authenticated requests. Must be in PEM format.",
		},
		"tls_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SPLUNK_CLIENT_KEY", ""),
			Description: "The client private key used to make authenticated requests. Must be in PEM format.",
			Sensitive:   true,
		},
		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN)",
		},
		"token": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SPLUNK_TOKEN", nil),
			Description: "The Splunk token to be used for authentication",
			Sensitive:   true,
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Splunk URL to stream logs to",
		},
		"use_tls": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to use TLS for secure logging. Default: `false`",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%h %l %u %t \"%r\" %>s %b",
			Description: "Apache-style string or VCL variables to use for log formatting (default: `%h %l %u %t \"%r\" %>s %b`)",
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
func (h *SplunkServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateSplunkInput{
		Format:         gofastly.ToPointer(vla.format),
		FormatVersion:  vla.formatVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		TLSCACert:      gofastly.ToPointer(resource["tls_ca_cert"].(string)),
		TLSClientCert:  gofastly.ToPointer(resource["tls_client_cert"].(string)),
		TLSClientKey:   gofastly.ToPointer(resource["tls_client_key"].(string)),
		TLSHostname:    gofastly.ToPointer(resource["tls_hostname"].(string)),
		Token:          gofastly.ToPointer(resource["token"].(string)),
		URL:            gofastly.ToPointer(resource["url"].(string)),
		UseTLS:         gofastly.ToPointer(gofastly.Compatibool(resource["use_tls"].(bool))),
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

	log.Printf("[DEBUG] Splunk create opts: %#v", opts)
	_, err := conn.CreateSplunk(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *SplunkServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Splunks for (%s)", d.Id())
		remoteState, err := conn.ListSplunks(&gofastly.ListSplunksInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Splunks for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		spl := flattenSplunks(remoteState)

		for _, element := range spl {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), spl); err != nil {
			log.Printf("[WARN] Error setting Splunks for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *SplunkServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateSplunkInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["request_max_entries"]; ok {
		opts.RequestMaxEntries = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["request_max_bytes"]; ok {
		opts.RequestMaxBytes = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_ca_cert"]; ok {
		opts.TLSCACert = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_hostname"]; ok {
		opts.TLSHostname = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_client_cert"]; ok {
		opts.TLSClientCert = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["tls_client_key"]; ok {
		opts.TLSClientKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["use_tls"]; ok {
		opts.UseTLS = gofastly.ToPointer(gofastly.Compatibool(v.(bool)))
	}

	log.Printf("[DEBUG] Update Splunk Opts: %#v", opts)
	_, err := conn.UpdateSplunk(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *SplunkServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteSplunkInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Splunk removal opts: %#v", opts)
	err := conn.DeleteSplunk(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenSplunks models data into format suitable for saving to Terraform state.
func flattenSplunks(remoteState []*gofastly.Splunk) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.URL != nil {
			data["url"] = *resource.URL
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.FormatVersion != nil {
			data["format_version"] = *resource.FormatVersion
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.Token != nil {
			data["token"] = *resource.Token
		}
		if resource.UseTLS != nil {
			data["use_tls"] = *resource.UseTLS
		}
		if resource.TLSHostname != nil {
			data["tls_hostname"] = *resource.TLSHostname
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
