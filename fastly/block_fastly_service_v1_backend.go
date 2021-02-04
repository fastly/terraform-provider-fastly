package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type BackendServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceBackend(sa ServiceMetadata) ServiceAttributeDefinition {
	return &BackendServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "backend",
			serviceMetadata: sa,
		},
	}
}

func (h *BackendServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	ob, nb := d.GetChange(h.GetKey())
	if ob == nil {
		ob = new(schema.Set)
	}
	if nb == nil {
		nb = new(schema.Set)
	}

	obs := ob.(*schema.Set)
	nbs := nb.(*schema.Set)

	setDiff := NewSetDiff(func(backend interface{}) (interface{}, error) {
		// Use the backend name as key
		return backend.(map[string]interface{})["name"], nil
	})

	diffResult, err := setDiff.Diff(obs, nbs)
	if err != nil {
		return err
	}

	// DELETE old Backends
	for _, bRaw := range diffResult.Deleted {
		bf := bRaw.(map[string]interface{})
		opts := h.createDeleteBackendInput(d.Id(), latestVersion, bf)

		log.Printf("[DEBUG] Fastly Backend removal opts: %#v", opts)
		err := conn.DeleteBackend(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// ADD new Backends
	for _, dRaw := range diffResult.Added {
		df := dRaw.(map[string]interface{})
		opts := h.buildCreateBackendInput(d.Id(), latestVersion, df)

		log.Printf("[DEBUG] Create Backend Opts: %#v", opts)
		_, err := conn.CreateBackend(&opts)
		if err != nil {
			return err
		}
	}

	// UPDATE existing backends
	for _, dRaw := range diffResult.Modified {
		backend := dRaw.(map[string]interface{})

		// Each element in the .Modified slice is a backend with nested fields.
		// Although we know the backend has been modified, we have to manually
		// filter out any unmodified nested fields by comparing them against the
		// original backend set data structure using .Filter.
		//
		modified := setDiff.Filter(backend, obs)
		opts := h.buildUpdateBackendInput(d.Id(), latestVersion, backend, modified)

		log.Printf("[DEBUG] Update Backend Opts: %#v", opts)
		_, err := conn.UpdateBackend(&opts)
		if err != nil {
			return err
		}
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
		SSLCiphers:          strings.Split(df["ssl_ciphers"].(string), ","),
		Shield:              df["shield"].(string),
		Port:                uint(df["port"].(int)),
		BetweenBytesTimeout: uint(df["between_bytes_timeout"].(int)),
		ConnectTimeout:      uint(df["connect_timeout"].(int)),
		ErrorThreshold:      uint(df["error_threshold"].(int)),
		FirstByteTimeout:    uint(df["first_byte_timeout"].(int)),
		MaxConn:             uint(df["max_conn"].(int)),
		Weight:              uint(df["weight"].(int)),
		HealthCheck:         df["healthcheck"].(string),
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		opts.RequestCondition = df["request_condition"].(string)
	}
	return opts
}

func (h *BackendServiceAttributeHandler) buildUpdateBackendInput(service string, latestVersion int, backend, modified map[string]interface{}) gofastly.UpdateBackendInput {
	input := gofastly.UpdateBackendInput{
		ServiceID:      service,
		ServiceVersion: latestVersion,
		Name:           backend["name"].(string),
	}

	if v, ok := modified["address"]; ok {
		input.Address = gofastly.String(v.(string))
	}
	if v, ok := modified["override_host"]; ok {
		input.OverrideHost = gofastly.String(v.(string))
	}
	if v, ok := modified["auto_loadbalance"]; ok {
		input.AutoLoadbalance = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["ssl_check_cert"]; ok {
		input.SSLCheckCert = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["ssl_hostname"]; ok {
		input.SSLHostname = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_ca_cert"]; ok {
		input.SSLCACert = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_cert_hostname"]; ok {
		input.SSLCertHostname = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_sni_hostname"]; ok {
		input.SSLSNIHostname = gofastly.String(v.(string))
	}
	if v, ok := modified["use_ssl"]; ok {
		input.UseSSL = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["ssl_client_key"]; ok {
		input.SSLClientKey = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_client_cert"]; ok {
		input.SSLClientCert = gofastly.String(v.(string))
	}
	if v, ok := modified["max_tls_version"]; ok {
		input.MaxTLSVersion = gofastly.String(v.(string))
	}
	if v, ok := modified["min_tls_version"]; ok {
		input.MinTLSVersion = gofastly.String(v.(string))
	}
	if v, ok := modified["ssl_ciphers"]; ok {
		input.SSLCiphers = strings.Split(v.(string), ",")
	}
	if v, ok := modified["shield"]; ok {
		input.Shield = gofastly.String(v.(string))
	}
	if v, ok := modified["port"]; ok {
		input.Port = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["between_bytes_timeout"]; ok {
		input.BetweenBytesTimeout = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["connect_timeout"]; ok {
		input.ConnectTimeout = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["error_threshold"]; ok {
		input.ErrorThreshold = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["first_byte_timeout"]; ok {
		input.FirstByteTimeout = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["max_conn"]; ok {
		input.MaxConn = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["weight"]; ok {
		input.Weight = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["healthcheck"]; ok {
		input.HealthCheck = gofastly.String(v.(string))
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		input.RequestCondition = gofastly.String(backend["request_condition"].(string))
	}
	return input
}

func (h *BackendServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Backends for (%s)", d.Id())
	backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	bl := flattenBackend(backendList, h.GetServiceMetadata())
	if err := d.Set(h.GetKey(), bl); err != nil {
		log.Printf("[WARN] Error setting Backends for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *BackendServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name for this Backend. Must be unique to this Service",
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
			Default:     true,
			Description: "Denotes if this Backend should be included in the pool of backends that requests are load balanced against. Default `true`",
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
			Description: "Comma separated list of OpenSSL Ciphers to try when negotiating to the backend",
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

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["request_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of a condition, which if met, will select this backend during a request.",
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
			"ssl_ciphers":           strings.Join(b.SSLCiphers, ","),
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
