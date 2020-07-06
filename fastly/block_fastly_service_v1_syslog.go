package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type SyslogServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceSyslog(sa ServiceAttributes) ServiceAttributeDefinition {
	return &SyslogServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:               "syslog",
			serviceAttributes: sa,
		},
	}
}

func (h *SyslogServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	os, ns := d.GetChange(h.GetKey())
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)
	removeSyslog := oss.Difference(nss).List()
	addSyslog := nss.Difference(oss).List()

	// DELETE old syslog configurations
	for _, pRaw := range removeSyslog {
		slf := pRaw.(map[string]interface{})
		opts := gofastly.DeleteSyslogInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    slf["name"].(string),
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
	}

	// POST new/updated Syslog
	for _, pRaw := range addSyslog {
		slf := pRaw.(map[string]interface{})

		var vla = h.getVCLLoggingAttributes(slf)
		opts := gofastly.CreateSyslogInput{
			Service:           d.Id(),
			Version:           latestVersion,
			Name:              slf["name"].(string),
			Address:           slf["address"].(string),
			Port:              uint(slf["port"].(int)),
			Token:             slf["token"].(string),
			UseTLS:            gofastly.CBool(slf["use_tls"].(bool)),
			TLSHostname:       slf["tls_hostname"].(string),
			TLSCACert:         slf["tls_ca_cert"].(string),
			TLSClientCert:     slf["tls_client_cert"].(string),
			TLSClientKey:      slf["tls_client_key"].(string),
			MessageType:       slf["message_type"].(string),
			Format:            vla.format,
			FormatVersion:     vla.formatVersion,
			ResponseCondition: vla.responseCondition,
			Placement:         vla.placement,
		}

		log.Printf("[DEBUG] Create Syslog Opts: %#v", opts)
		_, err := conn.CreateSyslog(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *SyslogServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Syslog for (%s)", d.Id())
	syslogList, err := conn.ListSyslogs(&gofastly.ListSyslogsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Syslog for (%s), version (%d): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	sll := flattenSyslogs(syslogList)

	if err := d.Set(h.GetKey(), sll); err != nil {
		log.Printf("[WARN] Error setting Syslog for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *SyslogServiceAttributeHandler) Register(s *schema.Resource) error {
	var a = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Unique name to refer to this logging setup",
		},
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The address of the syslog service",
		},
		// Optional
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     514,
			Description: "The port of the syslog service",
		},
		"token": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Authentication token",
		},
		"use_tls": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Use TLS for secure logging",
		},
		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Used during the TLS handshake to validate the certificate.",
		},
		"tls_ca_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CA_CERT", ""),
			Description: "A secure certificate to authenticate the server with. Must be in PEM format.",
		},
		"tls_client_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CLIENT_CERT", ""),
			Description: "The client certificate used to make authenticated requests. Must be in PEM format.",
		},
		"tls_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_SYSLOG_CLIENT_KEY", ""),
			Description: "The client private key used to make authenticated requests. Must be in PEM format.",
			Sensitive:   true,
		},
		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted.",
			ValidateFunc: validateLoggingMessageType(),
		},
	}

	if h.GetServiceAttributes().serviceType == ServiceTypeVCL {
		a["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%h %l %u %t \"%r\" %>s %b",
			Description: "Apache-style string or VCL variables to use for log formatting",
		}
		a["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      1,
			Description:  "The version of the custom logging format. Can be either 1 or 2. (Default: 1)",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		a["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of a condition to apply this logging.",
		}
		a["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed.",
			ValidateFunc: validateLoggingPlacement(),
		}
	}

	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: a,
		},
	}
	return nil
}

func flattenSyslogs(syslogList []*gofastly.Syslog) []map[string]interface{} {
	var pl []map[string]interface{}
	for _, p := range syslogList {
		// Convert Syslog to a map for saving to state.
		ns := map[string]interface{}{
			"name":               p.Name,
			"address":            p.Address,
			"port":               p.Port,
			"format":             p.Format,
			"format_version":     p.FormatVersion,
			"token":              p.Token,
			"use_tls":            p.UseTLS,
			"tls_hostname":       p.TLSHostname,
			"tls_ca_cert":        p.TLSCACert,
			"tls_client_cert":    p.TLSClientCert,
			"tls_client_key":     p.TLSClientKey,
			"response_condition": p.ResponseCondition,
			"message_type":       p.MessageType,
			"placement":          p.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		pl = append(pl, ns)
	}

	return pl
}
