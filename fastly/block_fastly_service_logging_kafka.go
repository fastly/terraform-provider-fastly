package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// KafkaServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type KafkaServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingKafka returns a new resource.
func NewServiceLoggingKafka(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&KafkaServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_kafka",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *KafkaServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *KafkaServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"auth_method": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "SASL authentication method. One of: plain, scram-sha-256, scram-sha-512",
		},
		"brokers": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A comma-separated list of IP addresses or hostnames of Kafka brokers",
		},
		"compression_codec": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The codec used for compression of your logs. One of: `gzip`, `snappy`, `lz4`",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Kafka logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"parse_log_keyvals": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enables parsing of key=value tuples from the beginning of a logline, turning them into record headers",
		},
		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "SASL Pass",
			Sensitive:   true,
		},
		"request_max_bytes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum size of log batch, if non-zero. Defaults to 0 for unbounded",
		},
		"required_acks": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Number of acknowledgements a leader must receive before a write is considered successful. One of: `1` (default) One server needs to respond. `0` No servers need to respond. `-1` Wait for all in-sync replicas to respond",
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
			Description: "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN)",
		},
		"topic": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Kafka topic to send logs to",
		},
		"use_tls": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to use TLS for secure logging. Can be either `true` or `false`",
		},
		"user": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "SASL User",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed.",
			ValidateDiagFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
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
func (h *KafkaServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Kafka logging addition opts: %#v", opts)

	return createKafka(conn, opts)
}

// Read refreshes the resource.
func (h *KafkaServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Kafka logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListKafkas(&gofastly.ListKafkasInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Kafka logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		kafkaLogList := flattenKafka(remoteState)

		for _, element := range kafkaLogList {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), kafkaLogList); err != nil {
			log.Printf("[WARN] Error setting Kafka logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *KafkaServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateKafkaInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["brokers"]; ok {
		opts.Brokers = gofastly.String(v.(string))
	}
	if v, ok := modified["topic"]; ok {
		opts.Topic = gofastly.String(v.(string))
	}
	if v, ok := modified["required_acks"]; ok {
		opts.RequiredACKs = gofastly.String(v.(string))
	}
	if v, ok := modified["use_tls"]; ok {
		opts.UseTLS = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.String(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}
	if v, ok := modified["tls_ca_cert"]; ok {
		opts.TLSCACert = gofastly.String(v.(string))
	}
	if v, ok := modified["tls_hostname"]; ok {
		opts.TLSHostname = gofastly.String(v.(string))
	}
	if v, ok := modified["tls_client_cert"]; ok {
		opts.TLSClientCert = gofastly.String(v.(string))
	}
	if v, ok := modified["tls_client_key"]; ok {
		opts.TLSClientKey = gofastly.String(v.(string))
	}
	if v, ok := modified["parse_log_keyvals"]; ok {
		opts.ParseLogKeyvals = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["request_max_bytes"]; ok {
		opts.RequestMaxBytes = gofastly.Int(v.(int))
	}
	if v, ok := modified["auth_method"]; ok {
		opts.AuthMethod = gofastly.String(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.String(v.(string))
	}
	if v, ok := modified["password"]; ok {
		opts.Password = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Kafka Opts: %#v", opts)
	_, err := conn.UpdateKafka(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *KafkaServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Kafka logging endpoint removal opts: %#v", opts)

	return deleteKafka(conn, opts)
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

// flattenKafka models data into format suitable for saving to Terraform state.
func flattenKafka(remoteState []*gofastly.Kafka) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"name":               resource.Name,
			"topic":              resource.Topic,
			"brokers":            resource.Brokers,
			"compression_codec":  resource.CompressionCodec,
			"required_acks":      resource.RequiredACKs,
			"use_tls":            resource.UseTLS,
			"tls_ca_cert":        resource.TLSCACert,
			"tls_client_cert":    resource.TLSClientCert,
			"tls_client_key":     resource.TLSClientKey,
			"tls_hostname":       resource.TLSHostname,
			"format":             resource.Format,
			"format_version":     resource.FormatVersion,
			"placement":          resource.Placement,
			"response_condition": resource.ResponseCondition,
			"parse_log_keyvals":  resource.ParseLogKeyvals,
			"request_max_bytes":  resource.RequestMaxBytes,
			"auth_method":        resource.AuthMethod,
			"user":               resource.User,
			"password":           resource.Password,
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

func (h *KafkaServiceAttributeHandler) buildCreate(kafkaMap any, serviceID string, serviceVersion int) *gofastly.CreateKafkaInput {
	resource := kafkaMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateKafkaInput{
		AuthMethod:       gofastly.String(resource["auth_method"].(string)),
		Brokers:          gofastly.String(resource["brokers"].(string)),
		CompressionCodec: gofastly.String(resource["compression_codec"].(string)),
		Format:           gofastly.String(vla.format),
		FormatVersion:    vla.formatVersion,
		Name:             gofastly.String(resource["name"].(string)),
		ParseLogKeyvals:  gofastly.CBool(resource["parse_log_keyvals"].(bool)),
		Password:         gofastly.String(resource["password"].(string)),
		RequestMaxBytes:  gofastly.Int(resource["request_max_bytes"].(int)),
		RequiredACKs:     gofastly.String(resource["required_acks"].(string)),
		ServiceID:        serviceID,
		ServiceVersion:   serviceVersion,
		TLSCACert:        gofastly.String(resource["tls_ca_cert"].(string)),
		TLSClientCert:    gofastly.String(resource["tls_client_cert"].(string)),
		TLSClientKey:     gofastly.String(resource["tls_client_key"].(string)),
		TLSHostname:      gofastly.String(resource["tls_hostname"].(string)),
		Topic:            gofastly.String(resource["topic"].(string)),
		UseTLS:           gofastly.CBool(resource["use_tls"].(bool)),
		User:             gofastly.String(resource["user"].(string)),
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.placement != "" {
		opts.Placement = gofastly.String(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.String(vla.responseCondition)
	}

	return opts
}

func (h *KafkaServiceAttributeHandler) buildDelete(kafkaMap any, serviceID string, serviceVersion int) *gofastly.DeleteKafkaInput {
	resource := kafkaMap.(map[string]any)

	return &gofastly.DeleteKafkaInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
