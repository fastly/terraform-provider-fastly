package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GCSLoggingServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type GCSLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingGCS returns a new resource.
func NewServiceLoggingGCS(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&GCSLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_gcs",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *GCSLoggingServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *GCSLoggingServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the bucket in which to store the logs",
		},
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
		},
		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: GzipLevelDescription,
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
			Description: "A unique name to identify this GCS endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Path to store the files. Must end with a trailing slash. If this field is left empty, the files will be saved in the bucket's root path",
		},
		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently the logs should be transferred, in seconds (Default 3600)",
		},
		"secret_key": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_SECRET_KEY", ""),
			Description: "The secret key associated with the target gcs bucket on your account. You may optionally provide this secret via an environment variable, `FASTLY_GCS_SECRET_KEY`. A typical format for the key is PEM format, containing actual newline characters where required",
			Sensitive:   true,
		},
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: TimestampFormatDescription,
		},
		"user": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_EMAIL", ""),
			Description: "Your Google Cloud Platform service account email address. The `client_email` field in your service account authentication JSON. You may optionally provide this via an environment variable, `FASTLY_GCS_EMAIL`.",
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
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 2)",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of a condition to apply this logging.",
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed.",
			ValidateDiagFunc: validateLoggingPlacement(),
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
func (h *GCSLoggingServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateGCSInput{
		ServiceID:         d.Id(),
		ServiceVersion:    serviceVersion,
		Name:              resource["name"].(string),
		User:              resource["user"].(string),
		Bucket:            resource["bucket_name"].(string),
		SecretKey:         resource["secret_key"].(string),
		Path:              resource["path"].(string),
		Period:            uint(resource["period"].(int)),
		GzipLevel:         uint8(resource["gzip_level"].(int)),
		TimestampFormat:   resource["timestamp_format"].(string),
		MessageType:       resource["message_type"].(string),
		CompressionCodec:  resource["compression_codec"].(string),
		Format:            vla.format,
		ResponseCondition: vla.responseCondition,
		Placement:         vla.placement,
	}

	log.Printf("[DEBUG] Create GCS Opts: %#v", opts)
	_, err := conn.CreateGCS(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *GCSLoggingServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing GCS for (%s)", d.Id())
		gs, err := conn.ListGCSs(&gofastly.ListGCSsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up GCS for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		gcsl := flattenGCS(gs)

		for _, element := range gcsl {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), gcsl); err != nil {
			log.Printf("[WARN] Error setting gcs for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *GCSLoggingServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateGCSInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: where we transition between any we lose the ability to
	// infer the underlying type being either a uint vs an int. This
	// materializes as a panic (yay) and so it's only at runtime we discover
	// this and so we've updated the below code to convert the type asserted
	// int into a uint before passing the value to gofastly.Uint().
	if v, ok := modified["bucket_name"]; ok {
		opts.Bucket = gofastly.String(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.String(v.(string))
	}
	if v, ok := modified["secret_key"]; ok {
		opts.SecretKey = gofastly.String(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.String(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.String(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		opts.GzipLevel = gofastly.Uint8(uint8(v.(int)))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update GCS Opts: %#v", opts)
	_, err := conn.UpdateGCS(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *GCSLoggingServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteGCSInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly GCS removal opts: %#v", opts)
	err := conn.DeleteGCS(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func flattenGCS(gcsList []*gofastly.GCS) []map[string]any {
	var sm []map[string]any
	for _, currentGCS := range gcsList {
		// Convert gcs to a map for saving to state.
		m := map[string]any{
			"name":               currentGCS.Name,
			"user":               currentGCS.User,
			"bucket_name":        currentGCS.Bucket,
			"secret_key":         currentGCS.SecretKey,
			"path":               currentGCS.Path,
			"period":             int(currentGCS.Period),
			"gzip_level":         int(currentGCS.GzipLevel),
			"response_condition": currentGCS.ResponseCondition,
			"message_type":       currentGCS.MessageType,
			"format":             currentGCS.Format,
			"format_version":     currentGCS.FormatVersion,
			"timestamp_format":   currentGCS.TimestampFormat,
			"placement":          currentGCS.Placement,
			"compression_codec":  currentGCS.CompressionCodec,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range m {
			if v == "" {
				delete(m, k)
			}
		}

		sm = append(sm, m)
	}

	return sm
}
