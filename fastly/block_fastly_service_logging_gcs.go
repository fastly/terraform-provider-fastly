package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
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
		"account_name": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_ACCOUNT_NAME", ""),
			Description: "The google account name used to obtain temporary credentials (default none). You may optionally provide this via an environment variable, `FASTLY_GCS_ACCOUNT_NAME`.",
		},
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
		"processing_region": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "none",
			Description:  "Region where logs will be processed before streaming to BigQuery. Valid values are 'none', 'us' and 'eu'.",
			ValidateFunc: validation.StringInSlice([]string{"none", "us", "eu"}, false),
		},
		"project_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID of your Google Cloud Platform project",
		},
		"secret_key": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_SECRET_KEY", ""),
			Description: "The secret key associated with the target gcs bucket on your account. You may optionally provide this secret via an environment variable, `FASTLY_GCS_SECRET_KEY`. A typical format for the key is PEM format, containing actual newline characters where required",
			Sensitive:   !DisplaySensitiveFields,
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
			Default:     LoggingGCSDefaultFormat,
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
func (h *GCSLoggingServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateGCSInput{
		Bucket:           gofastly.ToPointer(resource["bucket_name"].(string)),
		CompressionCodec: gofastly.ToPointer(resource["compression_codec"].(string)),
		Format:           gofastly.ToPointer(vla.format),
		MessageType:      gofastly.ToPointer(resource["message_type"].(string)),
		Name:             gofastly.ToPointer(resource["name"].(string)),
		Path:             gofastly.ToPointer(resource["path"].(string)),
		Period:           gofastly.ToPointer(resource["period"].(int)),
		ProjectID:        gofastly.ToPointer(resource["project_id"].(string)),
		SecretKey:        gofastly.ToPointer(resource["secret_key"].(string)),
		ServiceID:        d.Id(),
		ServiceVersion:   serviceVersion,
		TimestampFormat:  gofastly.ToPointer(resource["timestamp_format"].(string)),
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
	if v, ok := resource["account_name"].(string); ok && v != "" {
		opts.AccountName = gofastly.ToPointer(v)
	}

	log.Printf("[DEBUG] Create GCS Opts: %#v", opts)
	_, err := conn.CreateGCS(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *GCSLoggingServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing GCS for (%s)", d.Id())
		remoteState, err := conn.ListGCSs(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListGCSsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up GCS for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		gcsl := flattenGCS(remoteState, localState)

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
func (h *GCSLoggingServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateGCSInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["bucket_name"]; ok {
		opts.Bucket = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["account_name"]; ok {
		opts.AccountName = gofastly.ToPointer(v.(string))
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
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		// This condition prevents users on old provider versions from having
		// compatability issues with the default 'gzip_level' value of `-1` when upgrading to more recent versions.
		if gl := v.(int); gl != -1 {
			opts.GzipLevel = gofastly.ToPointer(gl)
		}
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["project_id"]; ok {
		opts.ProjectID = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update GCS Opts: %#v", opts)
	_, err := conn.UpdateGCS(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *GCSLoggingServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteGCSInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly GCS removal opts: %#v", opts)
	err := conn.DeleteGCS(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenGCS models data into format suitable for saving to Terraform state.
func flattenGCS(remoteState []*gofastly.GCS, state []any) []map[string]any {
	var result []map[string]any
	for _, resources := range remoteState {
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
		for _, s := range state {
			v := s.(map[string]any)
			if resources.Name != nil && v["name"].(string) == *resources.Name && v["gzip_level"].(int) == -1 {
				resources.GzipLevel = gofastly.ToPointer(v["gzip_level"].(int))
				break
			}
		}

		data := map[string]any{}

		if resources.Name != nil {
			data["name"] = *resources.Name
		}
		if resources.User != nil {
			data["user"] = *resources.User
		}
		if resources.AccountName != nil {
			data["account_name"] = *resources.AccountName
		}
		if resources.ProjectID != nil {
			data["project_id"] = *resources.ProjectID
		}
		if resources.Bucket != nil {
			data["bucket_name"] = *resources.Bucket
		}
		if resources.SecretKey != nil {
			data["secret_key"] = *resources.SecretKey
		}
		if resources.Path != nil {
			data["path"] = *resources.Path
		}
		if resources.Period != nil {
			data["period"] = *resources.Period
		}
		if resources.GzipLevel != nil {
			data["gzip_level"] = *resources.GzipLevel
		}
		if resources.ResponseCondition != nil {
			data["response_condition"] = *resources.ResponseCondition
		}
		if resources.MessageType != nil {
			data["message_type"] = *resources.MessageType
		}
		if resources.Format != nil {
			data["format"] = *resources.Format
		}
		if resources.FormatVersion != nil {
			data["format_version"] = *resources.FormatVersion
		}
		if resources.TimestampFormat != nil {
			data["timestamp_format"] = *resources.TimestampFormat
		}
		if resources.Placement != nil {
			data["placement"] = *resources.Placement
		}
		if resources.CompressionCodec != nil {
			data["compression_codec"] = *resources.CompressionCodec
		}
		if resources.ProcessingRegion != nil {
			data["processing_region"] = *resources.ProcessingRegion
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
