package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
)

// OpenstackServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type OpenstackServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingOpenstack returns a new resource.
func NewServiceLoggingOpenstack(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&OpenstackServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_openstack",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *OpenstackServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *OpenstackServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"access_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   !DisplaySensitiveFields,
			Description: "Your OpenStack account access key",
		},
		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of your OpenStack container",
		},
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
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
			Description: "The unique name of the OpenStack logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
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
		"processing_region": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "none",
			Description:  "Region where logs will be processed before streaming to BigQuery. Valid values are 'none', 'us' and 'eu'.",
			ValidateFunc: validation.StringInSlice([]string{"none", "us", "eu"}, false),
		},
		"public_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "A PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: TimestampFormatDescription,
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Your OpenStack auth url",
		},
		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The username for your OpenStack account",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     LoggingOpenStackDefaultFormat,
			Description: "Apache style log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed. Can be `none` or `none`.",
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
func (h *OpenstackServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly OpenStack logging addition opts: %#v", opts)

	_, err := conn.CreateOpenstack(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	return err
}

// Read refreshes the resource.
func (h *OpenstackServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing OpenStack logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListOpenstack(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListOpenstackInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up OpenStack logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenOpenstack(remoteState, localState)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting OpenStack logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *OpenstackServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateOpenstackInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["access_key"]; ok {
		opts.AccessKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["bucket_name"]; ok {
		opts.BucketName = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		// This condition prevents users on old provider versions from having
		// compatibility issues with the default 'gzip_level' value of `-1` when upgrading to more recent versions.
		if gl := v.(int); gl != -1 {
			opts.GzipLevel = gofastly.ToPointer(gl)
		}
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
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update OpenStack Opts: %#v", opts)
	_, err := conn.UpdateOpenstack(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *OpenstackServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly OpenStack logging endpoint removal opts: %#v", opts)

	err := conn.DeleteOpenstack(gofastly.NewContextForResourceID(ctx, d.Id()), opts)

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

// pruneVCLLoggingAttributes removes VCL-only attributes from Compute service data.
// For OpenStack logging, period is not VCL-only, so we preserve it.
func (h *OpenstackServiceAttributeHandler) pruneVCLLoggingAttributes(data map[string]any) {
	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		delete(data, "format")
		delete(data, "format_version")
		delete(data, "placement")
		delete(data, "response_condition")
		// Note: period is not deleted for OpenStack logging as it's available for both VCL and Compute
	}
}

// flattenOpenstack models data into format suitable for saving to Terraform state.
func flattenOpenstack(remoteState []*gofastly.Openstack, localState []any) []map[string]any {
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
		if resource.URL != nil {
			data["url"] = *resource.URL
		}
		if resource.User != nil {
			data["user"] = *resource.User
		}
		if resource.BucketName != nil {
			data["bucket_name"] = *resource.BucketName
		}
		if resource.AccessKey != nil {
			data["access_key"] = *resource.AccessKey
		}
		if resource.PublicKey != nil {
			data["public_key"] = *resource.PublicKey
		}
		if resource.GzipLevel != nil {
			data["gzip_level"] = *resource.GzipLevel
		}
		if resource.MessageType != nil {
			data["message_type"] = *resource.MessageType
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
		if resource.CompressionCodec != nil {
			data["compression_codec"] = *resource.CompressionCodec
		}
		if resource.ProcessingRegion != nil {
			data["processing_region"] = *resource.ProcessingRegion
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

func (h *OpenstackServiceAttributeHandler) buildCreate(openstackMap any, serviceID string, serviceVersion int) *gofastly.CreateOpenstackInput {
	resource := openstackMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateOpenstackInput{
		AccessKey:        gofastly.ToPointer(resource["access_key"].(string)),
		BucketName:       gofastly.ToPointer(resource["bucket_name"].(string)),
		CompressionCodec: gofastly.ToPointer(resource["compression_codec"].(string)),
		Format:           gofastly.ToPointer(vla.format),
		FormatVersion:    vla.formatVersion,
		MessageType:      gofastly.ToPointer(resource["message_type"].(string)),
		Name:             gofastly.ToPointer(resource["name"].(string)),
		Path:             gofastly.ToPointer(resource["path"].(string)),
		Period:           gofastly.ToPointer(resource["period"].(int)),
		PublicKey:        gofastly.ToPointer(resource["public_key"].(string)),
		ServiceID:        serviceID,
		ServiceVersion:   serviceVersion,
		TimestampFormat:  gofastly.ToPointer(resource["timestamp_format"].(string)),
		URL:              gofastly.ToPointer(resource["url"].(string)),
		User:             gofastly.ToPointer(resource["user"].(string)),
		ProcessingRegion: gofastly.ToPointer(resource["processing_region"].(string)),
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

func (h *OpenstackServiceAttributeHandler) buildDelete(openstackMap any, serviceID string, serviceVersion int) *gofastly.DeleteOpenstackInput {
	resource := openstackMap.(map[string]any)

	return &gofastly.DeleteOpenstackInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
