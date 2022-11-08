package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
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
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "An IPv4, hostname, or IPv6 address for the Backend",
		},
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
			Description: "Overrides ssl_hostname, but only for cert verification. Does not affect SNI at all",
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
			Sensitive:   true,
		},
		"ssl_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Client key attached to origin. Used when connecting to the backend",
			Sensitive:   true,
		},
		"ssl_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Used for both SNI during the TLS handshake and to validate the cert",
			Deprecated:  "Use ssl_cert_hostname and ssl_sni_hostname instead.",
		},
		"ssl_sni_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Overrides ssl_hostname, but only for SNI in the handshake. Does not affect cert validation at all",
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
func (h *BackendServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreateBackendInput(d.Id(), serviceVersion, resource)

	log.Printf("[DEBUG] Create Backend Opts: %#v", opts)
	_, err := conn.CreateBackend(&opts)
	if err != nil {
		return err
	}

	return nil
}

// Read refreshes the resource.
func (h *BackendServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Backends for (%s)", d.Id())
		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Backends for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		bl := flattenBackend(backendList, h.GetServiceMetadata())
		if err := d.Set(h.GetKey(), bl); err != nil {
			log.Printf("[WARN] Error setting Backends for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *BackendServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildUpdateBackendInput(d.Id(), serviceVersion, resource, modified)

	log.Printf("[DEBUG] Update Backend Opts: %#v", opts)
	_, err := conn.UpdateBackend(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *BackendServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
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

func (h *BackendServiceAttributeHandler) createDeleteBackendInput(service string, latestVersion int, bf map[string]any) gofastly.DeleteBackendInput {
	return gofastly.DeleteBackendInput{
		ServiceID:      service,
		ServiceVersion: latestVersion,
		Name:           bf["name"].(string),
	}
}

func (h *BackendServiceAttributeHandler) buildCreateBackendInput(service string, latestVersion int, df map[string]any) gofastly.CreateBackendInput {
	opts := gofastly.CreateBackendInput{
		ServiceID:           service,
		ServiceVersion:      latestVersion,
		Name:                gofastly.String(df["name"].(string)),
		Address:             gofastly.String(df["address"].(string)),
		OverrideHost:        gofastly.String(df["override_host"].(string)),
		AutoLoadbalance:     gofastly.CBool(df["auto_loadbalance"].(bool)),
		SSLCheckCert:        gofastly.CBool(df["ssl_check_cert"].(bool)),
		SSLHostname:         gofastly.String(df["ssl_hostname"].(string)),
		SSLCACert:           gofastly.String(df["ssl_ca_cert"].(string)),
		SSLCertHostname:     gofastly.String(df["ssl_cert_hostname"].(string)),
		SSLSNIHostname:      gofastly.String(df["ssl_sni_hostname"].(string)),
		UseSSL:              gofastly.CBool(df["use_ssl"].(bool)),
		SSLClientKey:        gofastly.String(df["ssl_client_key"].(string)),
		SSLClientCert:       gofastly.String(df["ssl_client_cert"].(string)),
		MaxTLSVersion:       gofastly.String(df["max_tls_version"].(string)),
		MinTLSVersion:       gofastly.String(df["min_tls_version"].(string)),
		SSLCiphers:          gofastly.String(df["ssl_ciphers"].(string)),
		Shield:              gofastly.String(df["shield"].(string)),
		Port:                gofastly.Int(df["port"].(int)),
		BetweenBytesTimeout: gofastly.Int(df["between_bytes_timeout"].(int)),
		ConnectTimeout:      gofastly.Int(df["connect_timeout"].(int)),
		ErrorThreshold:      gofastly.Int(df["error_threshold"].(int)),
		FirstByteTimeout:    gofastly.Int(df["first_byte_timeout"].(int)),
		MaxConn:             gofastly.Int(df["max_conn"].(int)),
		Weight:              gofastly.Int(df["weight"].(int)),
		HealthCheck:         gofastly.String(df["healthcheck"].(string)),
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		opts.RequestCondition = gofastly.String(df["request_condition"].(string))
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
		opts.Address = gofastly.String(v.(string))
	}
	if v, ok := modified["port"]; ok {
		opts.Port = gofastly.Int(v.(int))
	}
	if v, ok := modified["override_host"]; ok {
		opts.OverrideHost = gofastly.String(v.(string))
	}
	if v, ok := modified["connect_timeout"]; ok {
		opts.ConnectTimeout = gofastly.Int(v.(int))
	}
	if v, ok := modified["max_conn"]; ok {
		opts.MaxConn = gofastly.Int(v.(int))
	}
	if v, ok := modified["error_threshold"]; ok {
		opts.ErrorThreshold = gofastly.Int(v.(int))
	}
	if v, ok := modified["first_byte_timeout"]; ok {
		opts.FirstByteTimeout = gofastly.Int(v.(int))
	}
	if v, ok := modified["between_bytes_timeout"]; ok {
		opts.BetweenBytesTimeout = gofastly.Int(v.(int))
	}
	if v, ok := modified["auto_loadbalance"]; ok {
		opts.AutoLoadbalance = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["weight"]; ok {
		opts.Weight = gofastly.Int(v.(int))
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
		opts.SSLCiphers = gofastly.String(v.(string))
	}

	return opts
}

func flattenBackend(backendList []*gofastly.Backend, sa ServiceMetadata) []map[string]any {
	bl := make([]map[string]any, 0, len(backendList))

	for _, b := range backendList {
		backend := map[string]any{
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
