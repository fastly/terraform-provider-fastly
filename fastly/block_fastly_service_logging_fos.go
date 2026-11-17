package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
)

// FOSLoggingServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type FOSLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingFOS returns a new resource.
func NewServiceLoggingFOS(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&FOSLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_fos",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *FOSLoggingServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *FOSLoggingServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Fastly Object Storage bucket in which to store the logs",
		},
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
		},
		"file_max_bytes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum size of an uploaded log file, if non-zero. The minimum value is 1048576 bytes (1 MiB).",
		},
		"gzip_level": {
			Type:     schema.TypeInt,
			Optional: true,
			// -1 means unset; 0 is a valid API value.
			Default:     -1,
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
			Description: "The unique name of the Fastly Object Storage logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
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
			Description: "How frequently the logs should be transferred, in seconds. Default `3600`",
		},
		"public_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "A PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"region": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The region of the Fastly Object Storage bucket (e.g. `us-east`)",
		},
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: TimestampFormatDescription,
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Default:          LoggingFOSDefaultFormat,
			Description:      "Apache-style string or VCL variables to use for log formatting.",
			ValidateDiagFunc: validateLoggingFormat(),
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 2).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Name of blockAttributes condition to apply this logging.",
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
func (h *FOSLoggingServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly FOS logging addition opts: %#v", opts)

	_, err := conn.CreateFOS(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	return err
}

// Read refreshes the resource.
func (h *FOSLoggingServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing FOS logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListFOSes(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListFOSesInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up FOS logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		fl := flattenFOS(remoteState, localState)

		for _, element := range fl {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), fl); err != nil {
			log.Printf("[WARN] Error setting FOS logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *FOSLoggingServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateFOSInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: type assertions here panic if the type is wrong.
	if v, ok := modified["bucket_name"]; ok {
		opts.BucketName = new(v.(string))
	}
	if v, ok := modified["region"]; ok {
		opts.Region = new(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = new(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = new(v.(int))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = new(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		// Avoid sending the unset sentinel value of `-1` to the API.
		if gl := v.(int); gl != -1 {
			opts.GzipLevel = new(gl)
		}
	}
	if v, ok := modified["format"]; ok {
		opts.Format = new(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = new(v.(int))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = new(v.(string))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = new(v.(string))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = new(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.NewNullable(v.(string))
	}
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = new(v.(string))
	}
	if v, ok := modified["file_max_bytes"]; ok {
		opts.FileMaxBytes = new(v.(int))
	}

	log.Printf("[DEBUG] Update FOS Opts: %#v", opts)
	_, err := conn.UpdateFOS(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	return err
}

// Delete deletes the resource.
func (h *FOSLoggingServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly FOS logging endpoint removal opts: %#v", opts)

	err := conn.DeleteFOS(gofastly.NewContextForResourceID(ctx, d.Id()), opts)

	errRes, ok := err.(*gofastly.HTTPError)
	if !ok {
		return err
	}

	// Ignore 404s: the resource may have been deleted elsewhere.
	if !errRes.IsNotFound() {
		return err
	}

	return nil
}

// pruneVCLLoggingAttributes removes VCL-only attributes from Compute service data (period is kept).
func (h *FOSLoggingServiceAttributeHandler) pruneVCLLoggingAttributes(data map[string]any) {
	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		delete(data, "format")
		delete(data, "format_version")
		delete(data, "placement")
		delete(data, "response_condition")
	}
}

// flattenFOS models data into format suitable for saving to Terraform state.
func flattenFOS(remoteState []*gofastly.FOS, localState []any) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		// Keep an unset gzip_level (-1) instead of the API's 0 to avoid a perpetual diff.
		for _, s := range localState {
			v := s.(map[string]any)
			if resource.Name != nil && v["name"].(string) == *resource.Name && v["gzip_level"].(int) == -1 {
				resource.GzipLevel = new(v["gzip_level"].(int))
				break
			}
		}

		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.BucketName != nil {
			data["bucket_name"] = *resource.BucketName
		}
		if resource.Region != nil {
			data["region"] = *resource.Region
		}
		if resource.Path != nil {
			data["path"] = *resource.Path
		}
		if resource.Period != nil {
			data["period"] = *resource.Period
		}
		if resource.FileMaxBytes != nil {
			data["file_max_bytes"] = *resource.FileMaxBytes
		}
		if resource.GzipLevel != nil {
			data["gzip_level"] = *resource.GzipLevel
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.FormatVersion != nil {
			data["format_version"] = *resource.FormatVersion
		}
		if resource.TimestampFormat != nil {
			data["timestamp_format"] = *resource.TimestampFormat
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.MessageType != nil {
			data["message_type"] = *resource.MessageType
		}
		if resource.PublicKey != nil {
			data["public_key"] = *resource.PublicKey
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.CompressionCodec != nil {
			data["compression_codec"] = *resource.CompressionCodec
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
}

func (h *FOSLoggingServiceAttributeHandler) buildCreate(fosMap any, serviceID string, serviceVersion int) *gofastly.CreateFOSInput {
	resource := fosMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateFOSInput{
		BucketName:       new(resource["bucket_name"].(string)),
		CompressionCodec: new(resource["compression_codec"].(string)),
		FileMaxBytes:     new(resource["file_max_bytes"].(int)),
		Format:           new(vla.format),
		FormatVersion:    vla.formatVersion,
		MessageType:      new(resource["message_type"].(string)),
		Name:             new(resource["name"].(string)),
		Path:             new(resource["path"].(string)),
		Period:           new(resource["period"].(int)),
		PublicKey:        new(resource["public_key"].(string)),
		Region:           new(resource["region"].(string)),
		ServiceID:        serviceID,
		ServiceVersion:   serviceVersion,
		TimestampFormat:  new(resource["timestamp_format"].(string)),
	}

	// compression_codec and gzip_level are mutually exclusive, so only send gzip_level if set.
	if gl, ok := resource["gzip_level"].(int); ok && gl != -1 {
		opts.GzipLevel = new(gl)
	}

	// The API rejects empty strings for these fields.
	if vla.placement != "" {
		opts.Placement = new(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = new(vla.responseCondition)
	}

	return &opts
}

func (h *FOSLoggingServiceAttributeHandler) buildDelete(fosMap any, serviceID string, serviceVersion int) *gofastly.DeleteFOSInput {
	resource := fosMap.(map[string]any)

	return &gofastly.DeleteFOSInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
