package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		// required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name for this Backend. Must be unique to this Service. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "An IPv4, hostname, or IPv6 address for the Backend",
		},
		// Optional fields, defaults where they exist
		"auto_loadbalance": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Denotes if this Backend should be included in the pool of backends that requests are load balanced against. Default `false`",
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
		"max_conn": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     200,
			Description: "Maximum number of connections for this Backend. Default `200`",
		},
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     80,
			Description: "The port number on which the Backend responds. Default `80`",
		},
		"override_host": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "The hostname to override the Host header",
		},
		"shield": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "The POP of the shield designated to reduce inbound load. Valid values for `shield` are included in the `GET /datacenters` API response",
		},
		"use_ssl": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether or not to use SSL to reach the Backend. Default `false`",
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
		"ssl_ciphers": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Cipher list consisting of one or more cipher strings separated by colons. Commas or spaces are also acceptable separators but colons are normally used.",
		},
		"ssl_check_cert": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Be strict about checking SSL certs. Default `true`",
		},
		"ssl_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Used for both SNI during the TLS handshake and to validate the cert",
			Deprecated:  "Use ssl_cert_hostname and ssl_sni_hostname instead.",
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
			Description: "Overrides ssl_hostname, but only for cert verification. Does not affect SNI at all",
		},
		"ssl_sni_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Overrides ssl_hostname, but only for SNI in the handshake. Does not affect cert validation at all",
		},
		"ssl_client_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Client certificate attached to origin. Used when connecting to the backend",
			Sensitive:   true,
		},
		"ssl_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Client key attached to origin. Used when connecting to the backend",
			Sensitive:   true,
		},
		"weight": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     100,
			Description: "The [portion of traffic](https://docs.fastly.com/en/guides/load-balancing-configuration#how-weight-affects-load-balancing) to send to this Backend. Each Backend receives weight / total of the traffic. Default `100`",
		},
	}

	// backend is optional in both VCL and C@E
	required := false

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["request_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of a condition, which if met, will select this backend during a request.",
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
func (h *BackendServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreateBackendInput(d.Id(), serviceVersion, resource)

	log.Printf("[DEBUG] Create Backend Opts: %#v", opts)
	_, err := conn.CreateBackend(&opts)
	if err != nil {
		return err
	}

	return nil
}

// Read refreshes the resource.
func (h *BackendServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Backends for (%s)", d.Id())
	backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
	})
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%v): %s", d.Id(), serviceVersion, err)
	}

	bl := flattenBackend(backendList, h.GetServiceMetadata())
	if err := d.Set(h.GetKey(), bl); err != nil {
		log.Printf("[WARN] Error setting Backends for (%s): %s", d.Id(), err)
	}
	return nil
}

// Update updates the resource.
func (h *BackendServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildUpdateBackendInput(d.Id(), serviceVersion, resource, modified)

	log.Printf("[DEBUG] Update Backend Opts: %#v", opts)
	_, err := conn.UpdateBackend(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *BackendServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := h.createDeleteBackendInput(d.Id(), serviceVersion, resource)

	log.Printf("[DEBUG] Fastly Backend removal opts: %#v", opts)
	err := conn.DeleteBackend(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (h *BackendServiceAttributeHandler) createDeleteBackendInput(service string, latestVersion int, bf map[string]interface{}) gofastly.DeleteBackendInput {
	return gofastly.DeleteBackendInput{
		ServiceID:      service,
		ServiceVersion: latestVersion,
		Name:           bf["name"].(string),
	}
}

func (h *BackendServiceAttributeHandler) buildCreateBackendInput(service string, latestVersion int, df map[string]interface{}) gofastly.CreateBackendInput {
	opts := gofastly.CreateBackendInput{
		ServiceID:           service,
		ServiceVersion:      latestVersion,
		Name:                df["name"].(string),
		Address:             df["address"].(string),
		OverrideHost:        df["override_host"].(string),
		AutoLoadbalance:     gofastly.Compatibool(df["auto_loadbalance"].(bool)),
		SSLCheckCert:        gofastly.Compatibool(df["ssl_check_cert"].(bool)),
		SSLHostname:         df["ssl_hostname"].(string),
		SSLCACert:           df["ssl_ca_cert"].(string),
		SSLCertHostname:     df["ssl_cert_hostname"].(string),
		SSLSNIHostname:      df["ssl_sni_hostname"].(string),
		UseSSL:              gofastly.Compatibool(df["use_ssl"].(bool)),
		SSLClientKey:        df["ssl_client_key"].(string),
		SSLClientCert:       df["ssl_client_cert"].(string),
		MaxTLSVersion:       df["max_tls_version"].(string),
		MinTLSVersion:       df["min_tls_version"].(string),
		SSLCiphers:          df["ssl_ciphers"].(string),
		Shield:              df["shield"].(string),
		Port:                gofastly.Uint(uint(df["port"].(int))),
		BetweenBytesTimeout: gofastly.Uint(uint(df["between_bytes_timeout"].(int))),
		ConnectTimeout:      gofastly.Uint(uint(df["connect_timeout"].(int))),
		ErrorThreshold:      gofastly.Uint(uint(df["error_threshold"].(int))),
		FirstByteTimeout:    gofastly.Uint(uint(df["first_byte_timeout"].(int))),
		MaxConn:             gofastly.Uint(uint(df["max_conn"].(int))),
		Weight:              gofastly.Uint(uint(df["weight"].(int))),
		HealthCheck:         df["healthcheck"].(string),
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		opts.RequestCondition = df["request_condition"].(string)
	}
	return opts
}

func (h *BackendServiceAttributeHandler) buildUpdateBackendInput(serviceID string, latestVersion int, resource, modified map[string]interface{}) gofastly.UpdateBackendInput {
	opts := gofastly.UpdateBackendInput{
		ServiceID:      serviceID,
		ServiceVersion: latestVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: where we transition between interface{} we lose the ability to
	// infer the underlying type being either a uint vs an int. This
	// materializes as a panic (yay) and so it's only at runtime we discover
	// this and so we've updated the below code to convert the type asserted
	// int into a uint before passing the value to gofastly.Uint().
	if v, ok := modified["comment"]; ok {
		opts.Comment = gofastly.String(v.(string))
	}
	if v, ok := modified["address"]; ok {
		opts.Address = gofastly.String(v.(string))
	}
	if v, ok := modified["port"]; ok {
		opts.Port = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["override_host"]; ok {
		opts.OverrideHost = gofastly.String(v.(string))
	}
	if v, ok := modified["connect_timeout"]; ok {
		opts.ConnectTimeout = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["max_conn"]; ok {
		opts.MaxConn = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["error_threshold"]; ok {
		opts.ErrorThreshold = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["first_byte_timeout"]; ok {
		opts.FirstByteTimeout = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["between_bytes_timeout"]; ok {
		opts.BetweenBytesTimeout = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["auto_loadbalance"]; ok {
		opts.AutoLoadbalance = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["weight"]; ok {
		opts.Weight = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["request_condition"]; ok {
		if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
			opts.RequestCondition = gofastly.String(v.(string))
		}
	}
	if v, ok := modified["healthcheck"]; ok {
		opts.HealthCheck = gofastly.String(v.(string))
	}
	if v, ok := modified["shield"]; ok {
		opts.Shield = gofastly.String(v.(string))
	}
	if v, ok := modified["use_ssl"]; ok {
		opts.UseSSL = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["ssl_check_cert"]; ok {
		opts.SSLCheckCert = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["ssl_ca_cert"]; ok {
		opts.SSLCACert = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_client_cert"]; ok {
		opts.SSLClientCert = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_client_key"]; ok {
		opts.SSLClientKey = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_hostname"]; ok {
		opts.SSLHostname = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_cert_hostname"]; ok {
		opts.SSLCertHostname = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_sni_hostname"]; ok {
		opts.SSLSNIHostname = gofastly.String(v.(string))
	}
	if v, ok := modified["min_tls_version"]; ok {
		opts.MinTLSVersion = gofastly.String(v.(string))
	}
	if v, ok := modified["max_tls_version"]; ok {
		opts.MaxTLSVersion = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_ciphers"]; ok {
		opts.SSLCiphers = v.(string)
	}

	return opts
}

func flattenBackend(backendList []*gofastly.Backend, sa ServiceMetadata) []map[string]interface{} {
	bl := make([]map[string]interface{}, 0, len(backendList))

	for _, b := range backendList {

		backend := map[string]interface{}{
			"name":                  b.Name,
			"address":               b.Address,
			"auto_loadbalance":      b.AutoLoadbalance,
			"between_bytes_timeout": int(b.BetweenBytesTimeout),
			"connect_timeout":       int(b.ConnectTimeout),
			"error_threshold":       int(b.ErrorThreshold),
			"first_byte_timeout":    int(b.FirstByteTimeout),
			"max_conn":              int(b.MaxConn),
			"port":                  int(b.Port),
			"override_host":         b.OverrideHost,
			"shield":                b.Shield,
			"ssl_check_cert":        b.SSLCheckCert,
			"ssl_hostname":          b.SSLHostname,
			"ssl_ca_cert":           b.SSLCACert,
			"ssl_client_key":        b.SSLClientKey,
			"ssl_client_cert":       b.SSLClientCert,
			"max_tls_version":       b.MaxTLSVersion,
			"min_tls_version":       b.MinTLSVersion,
			"ssl_ciphers":           b.SSLCiphers,
			"use_ssl":               b.UseSSL,
			"ssl_cert_hostname":     b.SSLCertHostname,
			"ssl_sni_hostname":      b.SSLSNIHostname,
			"weight":                int(b.Weight),
			"healthcheck":           b.HealthCheck,
		}

		if sa.serviceType == ServiceTypeVCL {
			backend["request_condition"] = b.RequestCondition
		}

		bl = append(bl, backend)
	}
	return bl
}
