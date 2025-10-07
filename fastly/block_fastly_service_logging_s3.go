package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

// S3LoggingServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type S3LoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingS3 returns a new resource.
func NewServiceLoggingS3(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&S3LoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_s3",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *S3LoggingServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *S3LoggingServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"acl": {
			Type:     schema.TypeString,
			Optional: true,
			Description: fmt.Sprintf("The AWS [Canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl) to use for objects uploaded to the S3 bucket. Options are: `%s`, `%s`, `%s`, `%s`, `%s`, `%s`, `%s`",
				gofastly.S3AccessControlListPrivate,
				gofastly.S3AccessControlListPublicRead,
				gofastly.S3AccessControlListPublicReadWrite,
				gofastly.S3AccessControlListAWSExecRead,
				gofastly.S3AccessControlListAuthenticatedRead,
				gofastly.S3AccessControlListBucketOwnerRead,
				gofastly.S3AccessControlListBucketOwnerFullControl,
			),
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
				[]string{
					string(gofastly.S3AccessControlListPrivate),
					string(gofastly.S3AccessControlListPublicRead),
					string(gofastly.S3AccessControlListPublicReadWrite),
					string(gofastly.S3AccessControlListAWSExecRead),
					string(gofastly.S3AccessControlListAuthenticatedRead),
					string(gofastly.S3AccessControlListBucketOwnerRead),
					string(gofastly.S3AccessControlListBucketOwnerFullControl),
				},
				false,
			)),
		},
		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the bucket in which to store the logs",
		},
		"compression": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      CompressionDescription,
			ValidateDiagFunc: validateLoggingCompression(),
		},
		"domain": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "If you created the S3 bucket outside of `us-east-1`, then specify the corresponding bucket endpoint. Example: `s3-us-west-2.amazonaws.com`",
			Default:     "s3.amazonaws.com",
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
			Description: "The unique name of the S3 logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
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
		"redundancy": {
			Type:     schema.TypeString,
			Optional: true,
			Description: fmt.Sprintf("The S3 storage class (redundancy level). Should be one of: `%s`, `%s`, `%s`, `%s`, `%s`, `%s`, `%s`, or `%s`",
				gofastly.S3RedundancyStandard,
				gofastly.S3RedundancyIntelligentTiering,
				gofastly.S3RedundancyStandardIA,
				gofastly.S3RedundancyOneZoneIA,
				gofastly.S3RedundancyGlacierFlexibleRetrieval,
				gofastly.S3RedundancyGlacierInstantRetrieval,
				gofastly.S3RedundancyGlacierDeepArchive,
				gofastly.S3RedundancyReduced),
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
				[]string{
					string(gofastly.S3RedundancyStandard),
					string(gofastly.S3RedundancyIntelligentTiering),
					string(gofastly.S3RedundancyStandardIA),
					string(gofastly.S3RedundancyOneZoneIA),
					string(gofastly.S3RedundancyGlacierFlexibleRetrieval),
					string(gofastly.S3RedundancyGlacierInstantRetrieval),
					string(gofastly.S3RedundancyGlacierDeepArchive),
					string(gofastly.S3RedundancyReduced),
				},
				false,
			)),
		},
		"s3_access_key": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_ACCESS_KEY", ""),
			Description: "AWS Access Key of an account with the required permissions to post logs. It is **strongly** recommended you create a separate IAM user with permissions to only operate on this Bucket. This key will be not be encrypted. Not required if `iam_role` is provided. You can provide this key via an environment variable, `FASTLY_S3_ACCESS_KEY`",
			Sensitive:   !DisplaySensitiveFields,
		},
		"s3_iam_role": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_IAM_ROLE", ""),
			Description: "The Amazon Resource Name (ARN) for the IAM role granting Fastly access to S3. Not required if `access_key` and `secret_key` are provided. You can provide this value via an environment variable, `FASTLY_S3_IAM_ROLE`",
			Sensitive:   false,
		},
		"s3_secret_key": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_SECRET_KEY", ""),
			Description: "AWS Secret Key of an account with the required permissions to post logs. It is **strongly** recommended you create a separate IAM user with permissions to only operate on this Bucket. This secret will be not be encrypted. Not required if `iam_role` is provided. You can provide this secret via an environment variable, `FASTLY_S3_SECRET_KEY`",
			Sensitive:   !DisplaySensitiveFields,
		},
		"server_side_encryption": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Specify what type of server side encryption should be used. Can be either `AES256` or `aws:kms`",
			ValidateDiagFunc: validateLoggingServerSideEncryption(),
		},
		"server_side_encryption_kms_key_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Optional server-side KMS Key Id. Must be set if server_side_encryption is set to `aws:kms`",
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
			Default:     LoggingS3DefaultFormat,
			Description: "Apache-style string or VCL variables to use for log formatting.",
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
func (h *S3LoggingServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	_, err := conn.CreateS3(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	return err
}

// Read refreshes the resource.
func (h *S3LoggingServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing S3 Logging for (%s)", d.Id())
		remoteState, err := conn.ListS3s(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListS3sInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up S3 Logging for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		sl := flattenS3s(remoteState, localState)

		for _, element := range sl {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), sl); err != nil {
			log.Printf("[WARN] Error setting S3 Logging for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *S3LoggingServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateS3Input{
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
	if v, ok := modified["s3_access_key"]; ok {
		opts.AccessKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["s3_secret_key"]; ok {
		opts.SecretKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["s3_iam_role"]; ok {
		opts.IAMRole = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["compression"]; ok {
		opts.CompressionCodec, opts.GzipLevel = CompressionToAPIFields(v.(string))
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
	if v, ok := modified["redundancy"]; ok {
		opts.Redundancy = gofastly.ToPointer(gofastly.S3Redundancy(v.(string)))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["server_side_encryption_kms_key_id"]; ok {
		opts.ServerSideEncryptionKMSKeyID = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["server_side_encryption"]; ok {
		opts.ServerSideEncryption = gofastly.ToPointer(gofastly.S3ServerSideEncryption(v.(string)))
	}
	if v, ok := modified["acl"]; ok {
		opts.ACL = gofastly.ToPointer(gofastly.S3AccessControlList(v.(string)))
	}
	if v, ok := modified["file_max_bytes"]; ok {
		opts.FileMaxBytes = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["processing_region"]; ok {
		opts.ProcessingRegion = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update S3 Opts: %#v", opts)
	_, err := conn.UpdateS3(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *S3LoggingServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly S3 Logging removal opts: %#v", opts)

	err := conn.DeleteS3(gofastly.NewContextForResourceID(ctx, d.Id()), opts)

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
// For S3 logging, period is not VCL-only, so we preserve it.
func (h *S3LoggingServiceAttributeHandler) pruneVCLLoggingAttributes(data map[string]any) {
	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		delete(data, "format")
		delete(data, "format_version")
		delete(data, "placement")
		delete(data, "response_condition")
		// Note: period is not deleted for S3 logging as it's available for both VCL and Compute
	}
}

// flattenS3s models data into format suitable for saving to Terraform state.
func flattenS3s(remoteState []*gofastly.S3, _ []any) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.BucketName != nil {
			data["bucket_name"] = *resource.BucketName
		}
		if resource.AccessKey != nil {
			data["s3_access_key"] = *resource.AccessKey
		}
		if resource.SecretKey != nil {
			data["s3_secret_key"] = *resource.SecretKey
		}
		if resource.IAMRole != nil {
			data["s3_iam_role"] = *resource.IAMRole
		}
		if resource.Path != nil {
			data["path"] = *resource.Path
		}
		if resource.Period != nil {
			data["period"] = *resource.Period
		}
		if resource.Domain != nil {
			data["domain"] = *resource.Domain
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
		if resource.TimestampFormat != nil {
			data["timestamp_format"] = *resource.TimestampFormat
		}
		if resource.Redundancy != nil {
			data["redundancy"] = *resource.Redundancy
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
		if resource.ServerSideEncryption != nil {
			data["server_side_encryption"] = *resource.ServerSideEncryption
		}
		if resource.ServerSideEncryptionKMSKeyID != nil {
			data["server_side_encryption_kms_key_id"] = *resource.ServerSideEncryptionKMSKeyID
		}

		// compression represents the combined value of the compression_codec and gzip_level
		// attributes that we will need to parse to the API accordingly
		if compression := APIFieldsToCompression(resource.CompressionCodec, resource.GzipLevel); compression != "" {
			data["compression"] = compression
		}

		if resource.ACL != nil {
			data["acl"] = *resource.ACL
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

func (h *S3LoggingServiceAttributeHandler) buildCreate(s3Map any, serviceID string, serviceVersion int) *gofastly.CreateS3Input {
	resource := s3Map.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)

	// Convert the compression field to API fields
	var compressionCodec *string
	var gzipLevel *int
	if compression, ok := resource["compression"].(string); ok && compression != "" {
		compressionCodec, gzipLevel = CompressionToAPIFields(compression)
	}

	opts := gofastly.CreateS3Input{
		ACL:                          gofastly.ToPointer(gofastly.S3AccessControlList(resource["acl"].(string))),
		AccessKey:                    gofastly.ToPointer(resource["s3_access_key"].(string)),
		BucketName:                   gofastly.ToPointer(resource["bucket_name"].(string)),
		CompressionCodec:             compressionCodec,
		Domain:                       gofastly.ToPointer(resource["domain"].(string)),
		FileMaxBytes:                 gofastly.ToPointer(resource["file_max_bytes"].(int)),
		Format:                       gofastly.ToPointer(vla.format),
		FormatVersion:                vla.formatVersion,
		GzipLevel:                    gzipLevel,
		IAMRole:                      gofastly.ToPointer(resource["s3_iam_role"].(string)),
		MessageType:                  gofastly.ToPointer(resource["message_type"].(string)),
		Name:                         gofastly.ToPointer(resource["name"].(string)),
		Path:                         gofastly.ToPointer(resource["path"].(string)),
		Period:                       gofastly.ToPointer(resource["period"].(int)),
		PublicKey:                    gofastly.ToPointer(resource["public_key"].(string)),
		Redundancy:                   gofastly.ToPointer(gofastly.S3Redundancy(resource["redundancy"].(string))),
		SecretKey:                    gofastly.ToPointer(resource["s3_secret_key"].(string)),
		ServerSideEncryptionKMSKeyID: gofastly.ToPointer(resource["server_side_encryption_kms_key_id"].(string)),
		ServiceID:                    serviceID,
		ServiceVersion:               serviceVersion,
		TimestampFormat:              gofastly.ToPointer(resource["timestamp_format"].(string)),
		ProcessingRegion:             gofastly.ToPointer(resource["processing_region"].(string)),
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

	encryption := resource["server_side_encryption"].(string)
	switch encryption {
	case string(gofastly.S3ServerSideEncryptionAES):
		opts.ServerSideEncryption = gofastly.ToPointer(gofastly.S3ServerSideEncryptionAES)
	case string(gofastly.S3ServerSideEncryptionKMS):
		opts.ServerSideEncryption = gofastly.ToPointer(gofastly.S3ServerSideEncryptionKMS)
	}

	return &opts
}

func (h *S3LoggingServiceAttributeHandler) buildDelete(s3Map any, serviceID string, serviceVersion int) *gofastly.DeleteS3Input {
	resource := s3Map.(map[string]any)

	return &gofastly.DeleteS3Input{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
