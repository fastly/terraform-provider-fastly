package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SyslogServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type SyslogServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingSyslog returns a new resource.
func NewServiceLoggingSyslog(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&SyslogServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_syslog",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *SyslogServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *SyslogServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A hostname or IPv4 address of the Syslog endpoint",
		},
		"message_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "classic",
			Description:      MessageTypeDescription,
			ValidateDiagFunc: validateLoggingMessageType(),
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A unique name to identify this Syslog endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     514,
			Description: "The port associated with the address where the Syslog endpoint can be accessed. Default `514`",
		},
		"tls_ca_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CA_CERT", ""),
			Description: "A secure certificate to authenticate the server with. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SYSLOG_CA_CERT`",
		},
		"tls_client_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CLIENT_CERT", ""),
			Description: "The client certificate used to make authenticated requests. Must be in PEM format. You can provide this certificate via an environment variable, `FASTLY_SYSLOG_CLIENT_CERT`",
		},
		"tls_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CLIENT_KEY", ""),
			Description: "The client private key used to make authenticated requests. Must be in PEM format. You can provide this key via an environment variable, `FASTLY_SYSLOG_CLIENT_KEY`",
			Sensitive:   true,
		},
		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Used during the TLS handshake to validate the certificate",
		},
		"token": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Whether to prepend each message with a specific token",
		},
		"use_tls": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to use TLS for secure logging. Default `false`",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     `%h %l %u %t "%r" %>s %b`,
			Description: "Apache-style string or VCL variables to use for log formatting",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format. Can be either 1 or 2. (Default: 2)",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of blockAttributes condition to apply this logging.",
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Where in the generated VCL the logging call should be placed (ignored).",
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
func (h *SyslogServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateSyslogInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		Address:        gofastly.ToPointer(resource["address"].(string)),
		Port:           gofastly.ToPointer(resource["port"].(int)),
		Token:          gofastly.ToPointer(resource["token"].(string)),
		UseTLS:         gofastly.ToPointer(gofastly.Compatibool(resource["use_tls"].(bool))),
		TLSHostname:    gofastly.ToPointer(resource["tls_hostname"].(string)),
		TLSCACert:      gofastly.ToPointer(resource["tls_ca_cert"].(string)),
		TLSClientCert:  gofastly.ToPointer(resource["tls_client_cert"].(string)),
		TLSClientKey:   gofastly.ToPointer(resource["tls_client_key"].(string)),
		MessageType:    gofastly.ToPointer(resource["message_type"].(string)),
		Format:         gofastly.ToPointer(vla.format),
		FormatVersion:  vla.formatVersion,
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.ToPointer(vla.responseCondition)
	}
	if vla.placement != "" {
		opts.Placement = gofastly.ToPointer(vla.placement)
	}

	log.Printf("[DEBUG] Create Syslog Opts: %#v", opts)
	_, err := conn.CreateSyslog(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *SyslogServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Syslog for (%s)", d.Id())
		remoteState, err := conn.ListSyslogs(&gofastly.ListSyslogsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Syslog for (%s), version (%d): %s", d.Id(), serviceVersion, err)
		}

		sll := flattenSyslogs(remoteState)

		for _, element := range sll {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), sll); err != nil {
			log.Printf("[WARN] Error setting Syslog for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *SyslogServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateSyslogInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["address"]; ok {
		opts.Address = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["hostname"]; ok {
		opts.Hostname = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["port"]; ok {
		opts.Port = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["use_tls"]; ok {
		opts.UseTLS = gofastly.ToPointer(gofastly.Compatibool(v.(bool)))
	}
	if v, ok := modified["ipv4"]; ok {
		opts.IPV4 = gofastly.ToPointer(v.(string))
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
	if v, ok := modified["token"]; ok {
		opts.Token = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Syslog Opts: %#v", opts)
	_, err := conn.UpdateSyslog(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *SyslogServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteSyslogInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Syslog removal opts: %#v", opts)
	err := conn.DeleteSyslog(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenSyslogs models data into format suitable for saving to Terraform state.
func flattenSyslogs(remoteState []*gofastly.Syslog) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Address != nil {
			data["address"] = *resource.Address
		}
		if resource.Port != nil {
			data["port"] = *resource.Port
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.FormatVersion != nil {
			data["format_version"] = *resource.FormatVersion
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
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.MessageType != nil {
			data["message_type"] = *resource.MessageType
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
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
