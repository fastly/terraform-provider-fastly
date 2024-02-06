package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DigitalOceanServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type DigitalOceanServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingDigitalOcean returns a new resource.
func NewServiceLoggingDigitalOcean(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&DigitalOceanServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_digitalocean",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *DigitalOceanServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *DigitalOceanServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"access_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Your DigitalOcean Spaces account access key",
		},
		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the DigitalOcean Space",
		},
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
		},
		"domain": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The domain of the DigitalOcean Spaces endpoint (default `nyc3.digitaloceanspaces.com`)",
			Default:     "nyc3.digitaloceanspaces.com",
		},
		"gzip_level": {
			Type:     schema.TypeInt,
			Optional: true,
			// NOTE: The default represents an unset value
			// We use this instead of zero because the zero value for an int type is
			// actually a valid value for the API. The API will attempt to default to
			// zero if nothing is set by the user in their TF configuration.
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
			Description: "The unique name of the DigitalOcean Spaces logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The path to upload logs to",
		},
		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "How frequently log files are finalized so they can be available for reading (in seconds, default `3600`)",
		},
		"public_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "A PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"secret_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Your DigitalOcean Spaces account secret key",
		},
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: TimestampFormatDescription,
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
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.",
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
func (h *DigitalOceanServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly DigitalOcean Spaces logging addition opts: %#v", opts)

	return createDigitalOcean(conn, opts)
}

// Read refreshes the resource.
func (h *DigitalOceanServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing DigitalOcean Spaces logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListDigitalOceans(&gofastly.ListDigitalOceansInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up DigitalOcean Spaces logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenDigitalOcean(remoteState, localState)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting DigitalOcean Spaces logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *DigitalOceanServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateDigitalOceanInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["bucket_name"]; ok {
		opts.BucketName = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["domain"]; ok {
		opts.Domain = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["access_key"]; ok {
		opts.AccessKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["secret_key"]; ok {
		opts.SecretKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["gzip_level"]; ok {
		opts.GzipLevel = gofastly.ToPointer(v.(int))
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
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update DigitalOcean Opts: %#v", opts)
	_, err := conn.UpdateDigitalOcean(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *DigitalOceanServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly DigitalOcean Spaces logging endpoint removal opts: %#v", opts)

	return deleteDigitalOcean(conn, opts)
}

func createDigitalOcean(conn *gofastly.Client, i *gofastly.CreateDigitalOceanInput) error {
	_, err := conn.CreateDigitalOcean(i)
	return err
}

func deleteDigitalOcean(conn *gofastly.Client, i *gofastly.DeleteDigitalOceanInput) error {
	err := conn.DeleteDigitalOcean(i)

	errRes, ok := err.(*gofastly.HTTPError)
	if !ok {
		return err
	}

	// 404 response codes don't result in an error propagating because a 404 could
	// indicate that a resource was deleted elsewhere.
	if !errRes.IsNotFound() {
		return err
	}

	return nil
}

// flattenDigitalOcean models data into format suitable for saving to Terraform state.
func flattenDigitalOcean(remoteState []*gofastly.DigitalOcean, localState []any) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		// Avoid setting gzip_level to the API default of zero if originally unset.
		// This avoids an unnecessary diff where the local state would have been
		// updated to zero and so would be different from the -1 default set.
		// As the user never set the attribute we don't want to show a diff to say
		// it should be zero according to the API.
		//
		// NOTE: Ideally the local state would be updated when .Create() is called.
		// e.g. we'd check if the value is -1 for gzip_level and set it in state as
		// zero instead. This way we could avoid having to do this check here.
		// The reason that's not possible (or not ideal at least) is because Create
		// is called multiple times (once for each block defined in configuration)
		// while the setting of the state must be done holistically, and so what
		// that means is, if we did the above suggestion we would be resetting the
		// entire state object multiple times, where as here we're only ever setting
		// it once.
		for _, s := range localState {
			v := s.(map[string]any)
			if resource.Name != nil && v["name"].(string) == *resource.Name && v["gzip_level"].(int) == -1 {
				resource.GzipLevel = gofastly.ToPointer(v["gzip_level"].(int))
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
		if resource.Domain != nil {
			data["domain"] = *resource.Domain
		}
		if resource.AccessKey != nil {
			data["access_key"] = *resource.AccessKey
		}
		if resource.SecretKey != nil {
			data["secret_key"] = *resource.SecretKey
		}
		if resource.PublicKey != nil {
			data["public_key"] = *resource.PublicKey
		}
		if resource.Path != nil {
			data["path"] = *resource.Path
		}
		if resource.Period != nil {
			data["period"] = *resource.Period
		}
		if resource.TimestampFormat != nil {
			data["timestamp_format"] = *resource.TimestampFormat
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
		if resource.MessageType != nil {
			data["message_type"] = *resource.MessageType
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
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

func (h *DigitalOceanServiceAttributeHandler) buildCreate(digitaloceanMap any, serviceID string, serviceVersion int) *gofastly.CreateDigitalOceanInput {
	resource := digitaloceanMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateDigitalOceanInput{
		AccessKey:        gofastly.ToPointer(resource["access_key"].(string)),
		BucketName:       gofastly.ToPointer(resource["bucket_name"].(string)),
		CompressionCodec: gofastly.ToPointer(resource["compression_codec"].(string)),
		Domain:           gofastly.ToPointer(resource["domain"].(string)),
		Format:           gofastly.ToPointer(vla.format),
		FormatVersion:    vla.formatVersion,
		MessageType:      gofastly.ToPointer(resource["message_type"].(string)),
		Name:             gofastly.ToPointer(resource["name"].(string)),
		Path:             gofastly.ToPointer(resource["path"].(string)),
		Period:           gofastly.ToPointer(resource["period"].(int)),
		PublicKey:        gofastly.ToPointer(resource["public_key"].(string)),
		SecretKey:        gofastly.ToPointer(resource["secret_key"].(string)),
		ServiceID:        serviceID,
		ServiceVersion:   serviceVersion,
		TimestampFormat:  gofastly.ToPointer(resource["timestamp_format"].(string)),
	}

	// NOTE: go-fastly v7+ expects a pointer, so TF can't set the zero type value.
	// If we set a default value for an attribute, then it will be sent to the API.
	// In some scenarios this can cause the API to reject the request.
	// For example, configuring compression_codec + gzip_level is invalid.
	if gl, ok := resource["gzip_level"].(int); ok && gl != -1 {
		opts.GzipLevel = gofastly.ToPointer(gl)
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

func (h *DigitalOceanServiceAttributeHandler) buildDelete(digitaloceanMap any, serviceID string, serviceVersion int) *gofastly.DeleteDigitalOceanInput {
	resource := digitaloceanMap.(map[string]any)

	return &gofastly.DeleteDigitalOceanInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
