package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

// FTPServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type FTPServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingFTP returns a new resource.
func NewServiceLoggingFTP(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&FTPServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_ftp",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *FTPServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *FTPServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The FTP address to stream logs to",
		},
		"compression": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      CompressionDescription,
			ValidateDiagFunc: validateLoggingCompression(),
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
			Description: "The unique name of the FTP logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"password": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The password for the server (for anonymous use an email address)",
			Sensitive:   !DisplaySensitiveFields,
		},
		"path": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The path to upload log files to. If the path ends in `/` then it is treated as a directory",
		},
		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently the logs should be transferred, in seconds (Default `3600`)",
		},
		"port": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     21,
			Description: "The port number. Default: `21`",
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
			Description:      "The PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: TimestampFormatDescription,
		},
		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The username for the server (can be `anonymous`)",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache-style string or VCL variables to use for log formatting.",
			Default:     LoggingFTPDefaultFormat,
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
			Description: "The name of the condition to apply.",
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
func (h *FTPServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly FTP logging addition opts: %#v", opts)

	_, err := conn.CreateFTP(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	return err
}

// Read refreshes the resource.
func (h *FTPServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing FTP logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListFTPs(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListFTPsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up FTP logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenFTP(remoteState, localState)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting FTP logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *FTPServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateFTPInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["address"]; ok {
		opts.Address = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["port"]; ok {
		opts.Port = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.Username = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["password"]; ok {
		opts.Password = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.ToPointer(v.(int))
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
	if v, ok := modified["compression"]; ok {
		compressionCodec, gzipLevel := CompressionToAPIFields(v.(string))
		opts.CompressionCodec = compressionCodec
		opts.GzipLevel = gzipLevel
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update FTP Opts: %#v", opts)
	_, err := conn.UpdateFTP(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *FTPServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly FTP logging endpoint removal opts: %#v", opts)

	err := conn.DeleteFTP(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
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

// flattenFTP models data into format suitable for saving to Terraform state.
func flattenFTP(remoteState []*gofastly.FTP, _ []any) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Address != nil {
			data["address"] = *resource.Address
		}
		if resource.Username != nil {
			data["user"] = *resource.Username
		}
		if resource.Password != nil {
			data["password"] = *resource.Password
		}
		if resource.Path != nil {
			data["path"] = *resource.Path
		}
		if resource.Port != nil {
			data["port"] = *resource.Port
		}
		if resource.Period != nil {
			data["period"] = *resource.Period
		}
		if resource.PublicKey != nil {
			data["public_key"] = *resource.PublicKey
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
		if resource.MessageType != nil {
			data["message_type"] = *resource.MessageType
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
		}

		// compression represents the combined value of the compression_codec and gzip_level
		// attributes that we will need to parse to the API accordingly
		compression := APIFieldsToCompression(resource.CompressionCodec, resource.GzipLevel)
		if compression != "" {
			data["compression"] = compression
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

func (h *FTPServiceAttributeHandler) buildCreate(ftpMap any, serviceID string, serviceVersion int) *gofastly.CreateFTPInput {
	resource := ftpMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)

	// Convert the compression field to API fields
	var compressionCodec *string
	var gzipLevel *int
	if compression, ok := resource["compression"].(string); ok && compression != "" {
		compressionCodec, gzipLevel = CompressionToAPIFields(compression)
	}

	opts := &gofastly.CreateFTPInput{
		Address:          gofastly.ToPointer(resource["address"].(string)),
		CompressionCodec: compressionCodec,
		Format:           gofastly.ToPointer(vla.format),
		FormatVersion:    vla.formatVersion,
		GzipLevel:        gzipLevel,
		MessageType:      gofastly.ToPointer(resource["message_type"].(string)),
		Name:             gofastly.ToPointer(resource["name"].(string)),
		Password:         gofastly.ToPointer(resource["password"].(string)),
		Path:             gofastly.ToPointer(resource["path"].(string)),
		Period:           gofastly.ToPointer(resource["period"].(int)),
		Port:             gofastly.ToPointer(resource["port"].(int)),
		PublicKey:        gofastly.ToPointer(resource["public_key"].(string)),
		ServiceID:        serviceID,
		ServiceVersion:   serviceVersion,
		TimestampFormat:  gofastly.ToPointer(resource["timestamp_format"].(string)),
		Username:         gofastly.ToPointer(resource["user"].(string)),
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

func (h *FTPServiceAttributeHandler) buildDelete(ftpMap any, serviceID string, serviceVersion int) *gofastly.DeleteFTPInput {
	resource := ftpMap.(map[string]any)

	return &gofastly.DeleteFTPInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
