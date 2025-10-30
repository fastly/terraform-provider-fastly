package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

// BackendServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type BackendServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceBackend returns a new resource.
func NewServiceBackend(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&BackendServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "backend",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *BackendServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *BackendServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "An IPv4, hostname, or IPv6 address for the Backend",
		},
		"between_bytes_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     10000,
			Description: "How long to wait between bytes in milliseconds. Default `10000`",
		},
		"connect_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1000,
			Description: "How long to wait for a timeout in milliseconds. Default `1000`",
		},
		"error_threshold": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "Number of errors to allow before the Backend is marked as down. Default `0`",
		},
		"first_byte_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     15000,
			Description: "How long to wait for the first bytes in milliseconds. Default `15000`",
		},
		"healthcheck": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of a defined `healthcheck` to assign to this backend",
		},
		"keepalive_time": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "How long in seconds to keep a persistent connection to the backend between requests.",
		},
		"max_conn": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     200,
			Description: "Maximum number of connections for this Backend. Default `200`",
		},
		"max_tls_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Maximum allowed TLS version on SSL connections to this backend.",
		},
		"min_tls_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Minimum allowed TLS version on SSL connections to this backend.",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name for this Backend. Must be unique to this Service. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"override_host": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "The hostname to override the Host header",
		},
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     80,
			Description: "The port number on which the Backend responds. Default `80`",
		},
		"share_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Value that when shared across backends will enable those backends to share the same health check.",
		},
		"shield": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "The POP of the shield designated to reduce inbound load. Valid values for `shield` are included in the `GET /datacenters` API response",
		},
		"ssl_ca_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "CA certificate attached to origin.",
		},
		"ssl_cert_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Configure certificate validation. Does not affect SNI at all",
		},
		"ssl_check_cert": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Be strict about checking SSL certs. Default `true`",
		},
		"ssl_ciphers": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Cipher list consisting of one or more cipher strings separated by colons. Commas or spaces are also acceptable separators but colons are normally used.",
		},
		"ssl_client_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Client certificate attached to origin. Used when connecting to the backend",
			Sensitive:   !DisplaySensitiveFields,
		},
		"ssl_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Client key attached to origin. Used when connecting to the backend",
			Sensitive:   !DisplaySensitiveFields,
		},
		"ssl_sni_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Configure SNI in the TLS handshake. Does not affect cert validation at all",
		},
		"use_ssl": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether or not to use SSL to reach the Backend. Default `false`",
		},
		"weight": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     100,
			Description: "The [portion of traffic](https://docs.fastly.com/en/guides/load-balancing-configuration#how-weight-affects-load-balancing) to send to this Backend. Each Backend receives weight / total of the traffic. Default `100`",
		},
	}

	// backend is optional in both VCL and Compute
	required := false

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["auto_loadbalance"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Denotes if this Backend should be included in the pool of backends that requests are load balanced against. Default `false`",
		}
		blockAttributes["request_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of a condition, which if met, will select this backend during a request.",
		}
		blockAttributes["prefer_ipv6"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Prefer IPv6 connections to origins for hostname backends. Default `false`",
		}
	} else {
		blockAttributes["prefer_ipv6"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Prefer IPv6 connections to origins for hostname backends. Default `true`",
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: required,
		Optional: !required,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
}

// Create creates the resource.
func (h *BackendServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreateBackendInput(d.Id(), serviceVersion, resource)

	log.Printf("[DEBUG] Create Backend Opts: %#v", opts)
	_, err := conn.CreateBackend(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}

	return nil
}

// Read refreshes the resource.
func (h *BackendServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Backends for (%s)", d.Id())
		remoteState, err := conn.ListBackends(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListBackendsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Backends for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		bl := flattenBackend(remoteState, h.GetServiceMetadata())
		if err := d.Set(h.GetKey(), bl); err != nil {
			log.Printf("[WARN] Error setting Backends for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *BackendServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildUpdateBackendInput(d.Id(), serviceVersion, resource, modified)

	log.Printf("[DEBUG] Update Backend Opts: %#v", opts)
	_, err := conn.UpdateBackend(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *BackendServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.createDeleteBackendInput(d.Id(), serviceVersion, resource)

	log.Printf("[DEBUG] Fastly Backend removal opts: %#v", opts)
	err := conn.DeleteBackend(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (h *BackendServiceAttributeHandler) createDeleteBackendInput(service string, latestVersion int, bf map[string]any) gofastly.DeleteBackendInput {
	return gofastly.DeleteBackendInput{
		ServiceID:      service,
		ServiceVersion: latestVersion,
		Name:           bf["name"].(string),
	}
}

func (h *BackendServiceAttributeHandler) buildCreateBackendInput(service string, latestVersion int, resource map[string]any) gofastly.CreateBackendInput {
	opts := gofastly.CreateBackendInput{
		Address:             gofastly.ToPointer(resource["address"].(string)),
		BetweenBytesTimeout: gofastly.ToPointer(resource["between_bytes_timeout"].(int)),
		ConnectTimeout:      gofastly.ToPointer(resource["connect_timeout"].(int)),
		ErrorThreshold:      gofastly.ToPointer(resource["error_threshold"].(int)),
		FirstByteTimeout:    gofastly.ToPointer(resource["first_byte_timeout"].(int)),
		HealthCheck:         gofastly.ToPointer(resource["healthcheck"].(string)),
		MaxConn:             gofastly.ToPointer(resource["max_conn"].(int)),
		Name:                gofastly.ToPointer(resource["name"].(string)),
		Port:                gofastly.ToPointer(resource["port"].(int)),
		PreferIPv6:          gofastly.ToPointer(gofastly.Compatibool(resource["prefer_ipv6"].(bool))),
		SSLCheckCert:        gofastly.ToPointer(gofastly.Compatibool(resource["ssl_check_cert"].(bool))),
		ServiceID:           service,
		ServiceVersion:      latestVersion,
		Shield:              gofastly.ToPointer(resource["shield"].(string)),
		UseSSL:              gofastly.ToPointer(gofastly.Compatibool(resource["use_ssl"].(bool))),
		Weight:              gofastly.ToPointer(resource["weight"].(int)),
	}

	if resource["keepalive_time"].(int) > 0 {
		opts.KeepAliveTime = gofastly.ToPointer(resource["keepalive_time"].(int))
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if resource["min_tls_version"].(string) != "" {
		opts.MinTLSVersion = gofastly.ToPointer(resource["min_tls_version"].(string))
	}
	if resource["max_tls_version"].(string) != "" {
		opts.MaxTLSVersion = gofastly.ToPointer(resource["max_tls_version"].(string))
	}
	if resource["override_host"].(string) != "" {
		opts.OverrideHost = gofastly.ToPointer(resource["override_host"].(string))
	}
	if resource["share_key"].(string) != "" {
		opts.ShareKey = gofastly.ToPointer(resource["share_key"].(string))
	}
	if resource["ssl_ca_cert"].(string) != "" {
		opts.SSLCACert = gofastly.ToPointer(resource["ssl_ca_cert"].(string))
	}
	if resource["ssl_cert_hostname"].(string) != "" {
		opts.SSLCertHostname = gofastly.ToPointer(resource["ssl_cert_hostname"].(string))
	}
	if resource["ssl_ciphers"].(string) != "" {
		opts.SSLCiphers = gofastly.ToPointer(resource["ssl_ciphers"].(string))
	}
	if resource["ssl_client_cert"].(string) != "" {
		opts.SSLClientCert = gofastly.ToPointer(resource["ssl_client_cert"].(string))
	}
	if resource["ssl_client_key"].(string) != "" {
		opts.SSLClientKey = gofastly.ToPointer(resource["ssl_client_key"].(string))
	}
	if resource["ssl_sni_hostname"].(string) != "" {
		opts.SSLSNIHostname = gofastly.ToPointer(resource["ssl_sni_hostname"].(string))
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		opts.AutoLoadbalance = gofastly.ToPointer(gofastly.Compatibool(resource["auto_loadbalance"].(bool)))
		opts.RequestCondition = gofastly.ToPointer(resource["request_condition"].(string))
	}
	return opts
}

func (h *BackendServiceAttributeHandler) buildUpdateBackendInput(serviceID string, latestVersion int, resource, modified map[string]any) gofastly.UpdateBackendInput {
	opts := gofastly.UpdateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: latestVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["address"]; ok {
		opts.Address = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["port"]; ok {
		opts.Port = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["override_host"]; ok {
		opts.OverrideHost = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["connect_timeout"]; ok {
		opts.ConnectTimeout = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["keepalive_time"]; ok {
		opts.KeepAliveTime = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["max_conn"]; ok {
		opts.MaxConn = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["error_threshold"]; ok {
		opts.ErrorThreshold = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["first_byte_timeout"]; ok {
		opts.FirstByteTimeout = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["between_bytes_timeout"]; ok {
		opts.BetweenBytesTimeout = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["weight"]; ok {
		opts.Weight = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["request_condition"]; ok {
		if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
			opts.RequestCondition = gofastly.ToPointer(v.(string))
		}
	}
	if v, ok := modified["healthcheck"]; ok {
		opts.HealthCheck = gofastly.ToPointer(v.(string))
	}
	// NOTE: An empty string value will be coerced by Northstar into a null.
	// This will allow the share_key to be unset.
	if v, ok := modified["share_key"]; ok {
		opts.ShareKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["shield"]; ok {
		opts.Shield = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["ssl_ca_cert"]; ok {
		opts.SSLCACert = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["ssl_client_cert"]; ok {
		opts.SSLClientCert = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["ssl_client_key"]; ok {
		opts.SSLClientKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["ssl_cert_hostname"]; ok {
		opts.SSLCertHostname = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["ssl_sni_hostname"]; ok {
		opts.SSLSNIHostname = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["min_tls_version"]; ok {
		opts.MinTLSVersion = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["max_tls_version"]; ok {
		opts.MaxTLSVersion = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["ssl_ciphers"]; ok {
		opts.SSLCiphers = gofastly.ToPointer(v.(string))
	}
	// Always set optional boolean values to preserve state
	opts.UseSSL = gofastly.ToPointer(gofastly.Compatibool(resource["use_ssl"].(bool)))
	opts.SSLCheckCert = gofastly.ToPointer(gofastly.Compatibool(resource["ssl_check_cert"].(bool)))
	opts.PreferIPv6 = gofastly.ToPointer(gofastly.Compatibool(resource["prefer_ipv6"].(bool)))
	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		opts.AutoLoadbalance = gofastly.ToPointer(gofastly.Compatibool(resource["auto_loadbalance"].(bool)))
	}

	return opts
}

// flattenBackend models data into format suitable for saving to Terraform state.
func flattenBackend(remoteState []*gofastly.Backend, sa ServiceMetadata) []map[string]any {
	result := make([]map[string]any, 0, len(remoteState))

	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Address != nil {
			data["address"] = *resource.Address
		}
		if resource.BetweenBytesTimeout != nil {
			data["between_bytes_timeout"] = *resource.BetweenBytesTimeout
		}
		if resource.ConnectTimeout != nil {
			data["connect_timeout"] = *resource.ConnectTimeout
		}
		if resource.ErrorThreshold != nil {
			data["error_threshold"] = *resource.ErrorThreshold
		}
		if resource.FirstByteTimeout != nil {
			data["first_byte_timeout"] = *resource.FirstByteTimeout
		}
		if resource.HealthCheck != nil {
			data["healthcheck"] = *resource.HealthCheck
		}
		if resource.KeepAliveTime != nil {
			data["keepalive_time"] = *resource.KeepAliveTime
		}
		if resource.MaxConn != nil {
			data["max_conn"] = *resource.MaxConn
		}
		if resource.MaxTLSVersion != nil {
			data["max_tls_version"] = *resource.MaxTLSVersion
		}
		if resource.MinTLSVersion != nil {
			data["min_tls_version"] = *resource.MinTLSVersion
		}
		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.OverrideHost != nil {
			data["override_host"] = *resource.OverrideHost
		}
		if resource.Port != nil {
			data["port"] = *resource.Port
		}
		if resource.PreferIPv6 != nil {
			data["prefer_ipv6"] = *resource.PreferIPv6
		}
		if resource.ShareKey != nil {
			data["share_key"] = *resource.ShareKey
		}
		if resource.SSLCACert != nil {
			data["ssl_ca_cert"] = *resource.SSLCACert
		}
		if resource.SSLCertHostname != nil {
			data["ssl_cert_hostname"] = *resource.SSLCertHostname
		}
		if resource.SSLCheckCert != nil {
			data["ssl_check_cert"] = *resource.SSLCheckCert
		}
		if resource.SSLCiphers != nil {
			data["ssl_ciphers"] = *resource.SSLCiphers
		}
		if resource.SSLClientCert != nil {
			data["ssl_client_cert"] = *resource.SSLClientCert
		}
		if resource.SSLClientKey != nil {
			data["ssl_client_key"] = *resource.SSLClientKey
		}
		if resource.SSLSNIHostname != nil {
			data["ssl_sni_hostname"] = *resource.SSLSNIHostname
		}
		if resource.Shield != nil {
			data["shield"] = *resource.Shield
		}
		if resource.UseSSL != nil {
			data["use_ssl"] = *resource.UseSSL
		}
		if resource.Weight != nil {
			data["weight"] = *resource.Weight
		}

		if sa.serviceType == ServiceTypeVCL {
			if resource.AutoLoadbalance != nil {
				data["auto_loadbalance"] = *resource.AutoLoadbalance
			}
			if resource.RequestCondition != nil {
				data["request_condition"] = *resource.RequestCondition
			}
		}

		result = append(result, data)
	}
	return result
}
