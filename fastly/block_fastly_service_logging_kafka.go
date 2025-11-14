package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
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
			Sensitive:   !DisplaySensitiveFields,
		},
		"processing_region": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "none",
			Description:  "Region where logs will be processed before streaming to BigQuery. Valid values are 'none', 'us' and 'eu'.",
			ValidateFunc: validation.StringInSlice([]string{"none", "us", "eu"}, false),
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
			Sensitive:        !DisplaySensitiveFields,
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
			Default:     LoggingKafkaDefaultFormat,
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
func (h *KafkaServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Kafka logging addition opts: %#v", opts)

	_, err := conn.CreateKafka(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	return err
}

// Read refreshes the resource.
func (h *KafkaServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Kafka logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListKafkas(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListKafkasInput{
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
func (h *KafkaServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateKafkaInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// Always preserve optional bool values to prevent drift
	opts.UseTLS = gofastly.ToPointer(gofastly.Compatibool(resource["use_tls"].(bool)))
	opts.ParseLogKeyvals = gofastly.ToPointer(gofastly.Compatibool(resource["parse_log_keyvals"].(bool)))

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["brokers"]; ok {
		opts.Brokers = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["topic"]; ok {
		opts.Topic = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["required_acks"]; ok {
		opts.RequiredACKs = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.ToPointer(v.(string))
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
	if v, ok := modified["request_max_bytes"]; ok {
		opts.RequestMaxBytes = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["auth_method"]; ok {
		opts.AuthMethod = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["password"]; ok {
		opts.Password = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Kafka Opts: %#v", opts)

	_, err := conn.UpdateKafka(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	return err
}

// Delete deletes the resource.
func (h *KafkaServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Kafka logging endpoint removal opts: %#v", opts)

	err := conn.DeleteKafka(gofastly.NewContextForResourceID(ctx, d.Id()), opts)

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
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Topic != nil {
			data["topic"] = *resource.Topic
		}
		if resource.Brokers != nil {
			data["brokers"] = *resource.Brokers
		}
		if resource.CompressionCodec != nil {
			data["compression_codec"] = *resource.CompressionCodec
		}
		if resource.RequiredACKs != nil {
			data["required_acks"] = *resource.RequiredACKs
		}
		if resource.UseTLS != nil {
			data["use_tls"] = *resource.UseTLS
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
		if resource.TLSHostname != nil {
			data["tls_hostname"] = *resource.TLSHostname
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.FormatVersion != nil {
			data["format_version"] = *resource.FormatVersion
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.ParseLogKeyvals != nil {
			data["parse_log_keyvals"] = *resource.ParseLogKeyvals
		}
		if resource.RequestMaxBytes != nil {
			data["request_max_bytes"] = *resource.RequestMaxBytes
		}
		if resource.AuthMethod != nil {
			data["auth_method"] = *resource.AuthMethod
		}
		if resource.User != nil {
			data["user"] = *resource.User
		}
		if resource.Password != nil {
			data["password"] = *resource.Password
		}
		if resource.ProcessingRegion != nil {
			data["processing_region"] = *resource.ProcessingRegion
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
		AuthMethod:       gofastly.ToPointer(resource["auth_method"].(string)),
		Brokers:          gofastly.ToPointer(resource["brokers"].(string)),
		CompressionCodec: gofastly.ToPointer(resource["compression_codec"].(string)),
		Format:           gofastly.ToPointer(vla.format),
		FormatVersion:    vla.formatVersion,
		Name:             gofastly.ToPointer(resource["name"].(string)),
		ParseLogKeyvals:  gofastly.ToPointer(gofastly.Compatibool(resource["parse_log_keyvals"].(bool))),
		Password:         gofastly.ToPointer(resource["password"].(string)),
		RequestMaxBytes:  gofastly.ToPointer(resource["request_max_bytes"].(int)),
		RequiredACKs:     gofastly.ToPointer(resource["required_acks"].(string)),
		ServiceID:        serviceID,
		ServiceVersion:   serviceVersion,
		TLSCACert:        gofastly.ToPointer(resource["tls_ca_cert"].(string)),
		TLSClientCert:    gofastly.ToPointer(resource["tls_client_cert"].(string)),
		TLSClientKey:     gofastly.ToPointer(resource["tls_client_key"].(string)),
		TLSHostname:      gofastly.ToPointer(resource["tls_hostname"].(string)),
		Topic:            gofastly.ToPointer(resource["topic"].(string)),
		UseTLS:           gofastly.ToPointer(gofastly.Compatibool(resource["use_tls"].(bool))),
		User:             gofastly.ToPointer(resource["user"].(string)),
		ProcessingRegion: gofastly.ToPointer(resource["processing_region"].(string)),
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
