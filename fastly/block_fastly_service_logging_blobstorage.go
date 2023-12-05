package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// BlobStorageLoggingServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type BlobStorageLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingBlobStorage returns a new resource.
func NewServiceLoggingBlobStorage(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&BlobStorageLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_blobstorage",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *BlobStorageLoggingServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *BlobStorageLoggingServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"account_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique Azure Blob Storage namespace in which your data objects are stored",
		},
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
		},
		"container": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Azure Blob Storage container in which to store logs",
		},
		"file_max_bytes": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Maximum size of an uploaded log file, if non-zero.",
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
			Description: "A unique name to identify the Azure Blob Storage endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The path to upload logs to. Must end with a trailing slash. If this field is left empty, the files will be saved in the container's root path",
		},
		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently the logs should be transferred in seconds. Default `3600`",
		},
		"public_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "A PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"sas_token": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", ""),
			Description: "The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work",
			Sensitive:   true,
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
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%h %l %u %t \"%r\" %>s %b",
			Description: "Apache-style string or VCL variables to use for log formatting (default: `%h %l %u %t \"%r\" %>s %b`)",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed",
			ValidateDiagFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply",
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
func (h *BlobStorageLoggingServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateBlobStorageInput{
		AccountName:      gofastly.ToPointer(resource["account_name"].(string)),
		CompressionCodec: gofastly.ToPointer(resource["compression_codec"].(string)),
		Container:        gofastly.ToPointer(resource["container"].(string)),
		FileMaxBytes:     gofastly.ToPointer(resource["file_max_bytes"].(int)),
		Format:           gofastly.ToPointer(vla.format),
		FormatVersion:    vla.formatVersion,
		MessageType:      gofastly.ToPointer(resource["message_type"].(string)),
		Name:             gofastly.ToPointer(resource["name"].(string)),
		Path:             gofastly.ToPointer(resource["path"].(string)),
		Period:           gofastly.ToPointer(resource["period"].(int)),
		PublicKey:        gofastly.ToPointer(resource["public_key"].(string)),
		SASToken:         gofastly.ToPointer(resource["sas_token"].(string)),
		ServiceID:        d.Id(),
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

	log.Printf("[DEBUG] Blob Storage logging create opts: %#v", opts)
	_, err := conn.CreateBlobStorage(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *BlobStorageLoggingServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Blob Storages for (%s)", d.Id())
		remoteState, err := conn.ListBlobStorages(&gofastly.ListBlobStoragesInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Blob Storages for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		bsl := flattenBlobStorages(remoteState, localState)

		for _, element := range bsl {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), bsl); err != nil {
			log.Printf("[WARN] Error setting Blob Storages for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *BlobStorageLoggingServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateBlobStorageInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["account_name"]; ok {
		opts.AccountName = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["container"]; ok {
		opts.Container = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["sas_token"]; ok {
		opts.SASToken = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		opts.GzipLevel = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["file_max_bytes"]; ok {
		opts.FileMaxBytes = gofastly.ToPointer(v.(int))
	}

	log.Printf("[DEBUG] Update Blob Storage Opts: %#v", opts)
	_, err := conn.UpdateBlobStorage(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *BlobStorageLoggingServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteBlobStorageInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Blob Storage logging removal opts: %#v", opts)
	err := conn.DeleteBlobStorage(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenBlobStorages models data into format suitable for saving to Terraform state.
func flattenBlobStorages(remoteState []*gofastly.BlobStorage, localState []any) []map[string]any {
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
			if v["name"].(string) == resource.Name && v["gzip_level"].(int) == -1 {
				resource.GzipLevel = v["gzip_level"].(int)
				break
			}
		}

		data := map[string]any{
			"account_name":       resource.AccountName,
			"compression_codec":  resource.CompressionCodec,
			"container":          resource.Container,
			"file_max_bytes":     resource.FileMaxBytes,
			"format":             resource.Format,
			"format_version":     resource.FormatVersion,
			"gzip_level":         resource.GzipLevel,
			"message_type":       resource.MessageType,
			"name":               resource.Name,
			"path":               resource.Path,
			"period":             resource.Period,
			"placement":          resource.Placement,
			"public_key":         resource.PublicKey,
			"response_condition": resource.ResponseCondition,
			"sas_token":          resource.SASToken,
			"timestamp_format":   resource.TimestampFormat,
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
