package fastly

import (
	"fmt"
	"log"

	"github.com/fastly/go-fastly/fastly"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type KafkaServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingKafka() ServiceAttributeDefinition {
	return &KafkaServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key: "logging_kafka",
		},
	}
}

func (h *KafkaServiceAttributeHandler) Register(s *schema.Resource, serviceType string) error {
	var a = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Kafka logging endpoint.",
		},

		"topic": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Kafka topic to send logs to.",
		},

		"brokers": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A comma-separated list of IP addresses or hostnames of Kafka brokers.",
		},

		// Optional
		"compression_codec": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The codec used for compression of your logs. One of: gzip, snappy, lz4.",
		},

		"required_acks": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "The Number of acknowledgements a leader must receive before a write is considered successful. One of: 1 (default) One server needs to respond. 0 No servers need to respond. -1	Wait for all in-sync replicas to respond.",
		},

		"use_tls": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to use TLS for secure logging. Can be either true or false.",
		},

		"tls_ca_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A secure certificate to authenticate the server with. Must be in PEM format.",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_client_cert": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The client certificate used to make authenticated requests. Must be in PEM format.",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_client_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The client private key used to make authenticated requests. Must be in PEM format.",
			Sensitive:   true,
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"tls_hostname": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN).",
		},
	}

	if serviceType == ServiceTypeVCL {
		a["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting.",
		}
		a["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		a["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed.",
			ValidateFunc: validateLoggingPlacement(),
		}
		a["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
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

func (h *KafkaServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client, serviceType string) error {
	serviceID := d.Id()
	oldLogCfg, newLogCfg := d.GetChange(h.GetKey())

	if oldLogCfg == nil {
		oldLogCfg = new(schema.Set)
	}
	if newLogCfg == nil {
		newLogCfg = new(schema.Set)
	}

	oldLogSet := oldLogCfg.(*schema.Set)
	newLogSet := newLogCfg.(*schema.Set)

	removeKafkaLogging := oldLogSet.Difference(newLogSet).List()
	addKafkaLogging := newLogSet.Difference(oldLogSet).List()

	// DELETE old Kafka logging endpoints
	for _, oRaw := range removeKafkaLogging {
		of := oRaw.(map[string]interface{})
		opts := buildDeleteKafka(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Kafka logging endpoint removal opts: %#v", opts)

		if err := deleteKafka(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Kafka logging endponts
	for _, nRaw := range addKafkaLogging {
		cfg := nRaw.(map[string]interface{})

		// @HACK for a TF SDK Issue.
		//
		// This ensures that the required, `name`, field is present.
		//
		// If we have made it this far and `name` is not present, it is most-likely due
		// to a defunct diff as noted here - https://github.com/hashicorp/terraform-plugin-sdk/issues/160#issuecomment-522935697.
		//
		// This is caused by using a StateFunc in a nested TypeSet. While the StateFunc
		// properly handles setting state with the StateFunc, it returns extra entries
		// during state Gets, specifically `GetChange("logging_kafka")` in this case.
		if v, ok := cfg["name"]; !ok || v.(string) == "" {
			continue
		}

		opts := buildCreateKafka(cfg, serviceID, latestVersion, serviceType)

		log.Printf("[DEBUG] Fastly Kafka logging addition opts: %#v", opts)

		if err := createKafka(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *KafkaServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client, serviceType string) error {
	// refresh Kafka
	log.Printf("[DEBUG] Refreshing Kafka logging endpoints for (%s)", d.Id())
	kafkaList, err := conn.ListKafkas(&gofastly.ListKafkasInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Kafka logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	kafkaLogList := flattenKafka(kafkaList)

	if err := d.Set(h.GetKey(), kafkaLogList); err != nil {
		log.Printf("[WARN] Error setting Kafka logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createKafka(conn *gofastly.Client, i *gofastly.CreateKafkaInput) error {
	_, err := conn.CreateKafka(i)
	return err
}

func deleteKafka(conn *gofastly.Client, i *gofastly.DeleteKafkaInput) error {
	err := conn.DeleteKafka(i)

	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func flattenKafka(kafkaList []*gofastly.Kafka) []map[string]interface{} {
	var flattened []map[string]interface{}
	for _, s := range kafkaList {
		// Convert logging to a map for saving to state.
		flatKafka := map[string]interface{}{
			"name":               s.Name,
			"topic":              s.Topic,
			"brokers":            s.Brokers,
			"compression_codec":  s.CompressionCodec,
			"required_acks":      s.RequiredACKs,
			"use_tls":            s.UseTLS,
			"tls_ca_cert":        s.TLSCACert,
			"tls_client_cert":    s.TLSClientCert,
			"tls_client_key":     s.TLSClientKey,
			"tls_hostname":       s.TLSHostname,
			"format":             s.Format,
			"format_version":     s.FormatVersion,
			"placement":          s.Placement,
			"response_condition": s.ResponseCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range flatKafka {
			if v == "" {
				delete(flatKafka, k)
			}
		}

		flattened = append(flattened, flatKafka)
	}

	return flattened
}

func buildCreateKafka(kafkaMap interface{}, serviceID string, serviceVersion int, serviceType string) *gofastly.CreateKafkaInput {
	df := kafkaMap.(map[string]interface{})

	var vla = NewVCLLoggingAttributes()
	if serviceType == ServiceTypeVCL {
		vla.format = df["format"].(string)
		vla.formatVersion = uint(df["format_version"].(int))
		vla.placement = df["placement"].(string)
		vla.responseCondition = df["response_condition"].(string)
	}

	return &gofastly.CreateKafkaInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              fastly.NullString(df["name"].(string)),
		Brokers:           fastly.NullString(df["brokers"].(string)),
		Topic:             fastly.NullString(df["topic"].(string)),
		RequiredACKs:      fastly.NullString(df["required_acks"].(string)),
		UseTLS:            fastly.CBool(df["use_tls"].(bool)),
		CompressionCodec:  fastly.NullString(df["compression_codec"].(string)),
		TLSCACert:         fastly.NullString(df["tls_ca_cert"].(string)),
		TLSClientCert:     fastly.NullString(df["tls_client_cert"].(string)),
		TLSClientKey:      fastly.NullString(df["tls_client_key"].(string)),
		TLSHostname:       fastly.NullString(df["tls_hostname"].(string)),
		Format:            gofastly.NullString(vla.format),
		FormatVersion:     gofastly.Uint(vla.formatVersion),
		Placement:         gofastly.NullString(vla.placement),
		ResponseCondition: gofastly.NullString(vla.responseCondition),
	}
}

func buildDeleteKafka(kafkaMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteKafkaInput {
	df := kafkaMap.(map[string]interface{})

	return &gofastly.DeleteKafkaInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
