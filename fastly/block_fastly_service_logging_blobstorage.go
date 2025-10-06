package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
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
		"compression": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      CompressionDescription,
			ValidateDiagFunc: validateLoggingCompression(),
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
		"sas_token": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", ""),
			Description: "The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work",
			Sensitive:   !DisplaySensitiveFields,
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
			Default:     LoggingBlobStorageDefaultFormat,
			Description: "Apache-style string or VCL variables to use for log formatting.",
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
func (h *BlobStorageLoggingServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	vla := h.getVCLLoggingAttributes(resource)

	// Convert the compression field to API fields
	var compressionCodec *string
	var gzipLevel *int
	if compression, ok := resource["compression"].(string); ok && compression != "" {
		compressionCodec, gzipLevel = CompressionToAPIFields(compression)
	}

	opts := gofastly.CreateBlobStorageInput{
		AccountName:      gofastly.ToPointer(resource["account_name"].(string)),
		CompressionCodec: compressionCodec,
		Container:        gofastly.ToPointer(resource["container"].(string)),
		FileMaxBytes:     gofastly.ToPointer(resource["file_max_bytes"].(int)),
		Format:           gofastly.ToPointer(vla.format),
		FormatVersion:    vla.formatVersion,
		GzipLevel:        gzipLevel,
		MessageType:      gofastly.ToPointer(resource["message_type"].(string)),
		Name:             gofastly.ToPointer(resource["name"].(string)),
		Path:             gofastly.ToPointer(resource["path"].(string)),
		Period:           gofastly.ToPointer(resource["period"].(int)),
		PublicKey:        gofastly.ToPointer(resource["public_key"].(string)),
		SASToken:         gofastly.ToPointer(resource["sas_token"].(string)),
		ServiceID:        d.Id(),
		ServiceVersion:   serviceVersion,
		TimestampFormat:  gofastly.ToPointer(resource["timestamp_format"].(string)),
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

	log.Printf("[DEBUG] Blob Storage logging create opts: %#v", opts)
	_, err := conn.CreateBlobStorage(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *BlobStorageLoggingServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Blob Storages for (%s)", d.Id())
		remoteState, err := conn.ListBlobStorages(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListBlobStoragesInput{
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
func (h *BlobStorageLoggingServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
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
	if v, ok := modified["compression"]; ok {
		compressionCodec, gzipLevel := CompressionToAPIFields(v.(string))
		opts.CompressionCodec = compressionCodec
		opts.GzipLevel = gzipLevel
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
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Blob Storage Opts: %#v", opts)
	_, err := conn.UpdateBlobStorage(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *BlobStorageLoggingServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteBlobStorageInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Blob Storage logging removal opts: %#v", opts)
	err := conn.DeleteBlobStorage(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// pruneVCLLoggingAttributes removes VCL-only attributes from Compute service data.
// For BlobStorage logging, period is not VCL-only, so we preserve it.
func (h *BlobStorageLoggingServiceAttributeHandler) pruneVCLLoggingAttributes(data map[string]any) {
	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		delete(data, "format")
		delete(data, "format_version")
		delete(data, "placement")
		delete(data, "response_condition")
		// Note: period is not deleted for BlobStorage logging as it's available for both VCL and Compute
	}
}

// flattenBlobStorages models data into format suitable for saving to Terraform state.
func flattenBlobStorages(remoteState []*gofastly.BlobStorage, localState []any) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.AccountName != nil {
			data["account_name"] = *resource.AccountName
		}

		// Check if compression was originally set in the config by looking at localState
		var compressionSetInConfig bool
		for _, s := range localState {
			v := s.(map[string]any)
			if resource.Name != nil && v["name"].(string) == *resource.Name {
				_, compressionSetInConfig = v["compression"]
				break
			}
		}

		// compression represents the combined value of the compression_codec and gzip_level
		// attributes that we will need to parse to the API accordingly
		// Only set it in state if it was originally specified in the config
		if compressionSetInConfig {
			compression := APIFieldsToCompression(resource.CompressionCodec, resource.GzipLevel)
			if compression != "" {
				data["compression"] = compression
			}
		}

		if resource.Container != nil {
			data["container"] = *resource.Container
		}
		if resource.FileMaxBytes != nil {
			data["file_max_bytes"] = *resource.FileMaxBytes
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
		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Path != nil {
			data["path"] = *resource.Path
		}
		if resource.Period != nil {
			data["period"] = *resource.Period
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.PublicKey != nil {
			data["public_key"] = *resource.PublicKey
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}
		if resource.SASToken != nil {
			data["sas_token"] = *resource.SASToken
		}
		if resource.TimestampFormat != nil {
			data["timestamp_format"] = *resource.TimestampFormat
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
