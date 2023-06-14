package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
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
			Sensitive:   true,
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
			Sensitive:   true,
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
			Default:     `%h %l %u %t "%r" %>s %b`,
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
func (h *S3LoggingServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts, err := h.buildCreate(resource, d.Id(), serviceVersion)
	if err != nil {
		return err
	}

	err = createS3(conn, opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *S3LoggingServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing S3 Logging for (%s)", d.Id())
		remoteState, err := conn.ListS3s(&gofastly.ListS3sInput{
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
func (h *S3LoggingServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateS3Input{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["bucket_name"]; ok {
		opts.BucketName = gofastly.String(v.(string))
	}
	if v, ok := modified["domain"]; ok {
		opts.Domain = gofastly.String(v.(string))
	}
	if v, ok := modified["s3_access_key"]; ok {
		opts.AccessKey = gofastly.String(v.(string))
	}
	if v, ok := modified["s3_secret_key"]; ok {
		opts.SecretKey = gofastly.String(v.(string))
	}
	if v, ok := modified["s3_iam_role"]; ok {
		opts.IAMRole = gofastly.String(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.String(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.Int(v.(int))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.String(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		opts.GzipLevel = gofastly.Int(v.(int))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["message_type"]; ok {
		opts.MessageType = gofastly.String(v.(string))
	}
	if v, ok := modified["timestamp_format"]; ok {
		opts.TimestampFormat = gofastly.String(v.(string))
	}
	if v, ok := modified["redundancy"]; ok {
		opts.Redundancy = gofastly.S3RedundancyPtr(gofastly.S3Redundancy(v.(string)))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = gofastly.String(v.(string))
	}
	if v, ok := modified["server_side_encryption_kms_key_id"]; ok {
		opts.ServerSideEncryptionKMSKeyID = gofastly.String(v.(string))
	}
	if v, ok := modified["server_side_encryption"]; ok {
		opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionPtr(gofastly.S3ServerSideEncryption(v.(string)))
	}
	if v, ok := modified["acl"]; ok {
		opts.ACL = gofastly.S3AccessControlListPtr(gofastly.S3AccessControlList(v.(string)))
	}
	if v, ok := modified["file_max_bytes"]; ok {
		opts.FileMaxBytes = gofastly.Int(v.(int))
	}

	log.Printf("[DEBUG] Update S3 Opts: %#v", opts)
	_, err := conn.UpdateS3(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *S3LoggingServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)
	err := deleteS3(conn, opts)
	if err != nil {
		return err
	}
	return nil
}

func createS3(conn *gofastly.Client, i *gofastly.CreateS3Input) error {
	_, err := conn.CreateS3(i)
	return err
}

func deleteS3(conn *gofastly.Client, i *gofastly.DeleteS3Input) error {
	log.Printf("[DEBUG] Fastly S3 Logging removal opts: %#v", i)

	err := conn.DeleteS3(i)

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

// flattenS3s models data into format suitable for saving to Terraform state.
func flattenS3s(remoteState []*gofastly.S3, localState []any) []map[string]any {
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
			"name":                              resource.Name,
			"bucket_name":                       resource.BucketName,
			"s3_access_key":                     resource.AccessKey,
			"s3_secret_key":                     resource.SecretKey,
			"s3_iam_role":                       resource.IAMRole,
			"path":                              resource.Path,
			"period":                            resource.Period,
			"domain":                            resource.Domain,
			"file_max_bytes":                    resource.FileMaxBytes,
			"gzip_level":                        resource.GzipLevel,
			"format":                            resource.Format,
			"format_version":                    resource.FormatVersion,
			"timestamp_format":                  resource.TimestampFormat,
			"redundancy":                        resource.Redundancy,
			"response_condition":                resource.ResponseCondition,
			"message_type":                      resource.MessageType,
			"public_key":                        resource.PublicKey,
			"placement":                         resource.Placement,
			"server_side_encryption":            resource.ServerSideEncryption,
			"server_side_encryption_kms_key_id": resource.ServerSideEncryptionKMSKeyID,
			"compression_codec":                 resource.CompressionCodec,
			"acl":                               resource.ACL,
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

func (h *S3LoggingServiceAttributeHandler) buildCreate(s3Map any, serviceID string, serviceVersion int) (*gofastly.CreateS3Input, error) {
	resource := s3Map.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := gofastly.CreateS3Input{
		ACL:                          gofastly.S3AccessControlListPtr(gofastly.S3AccessControlList(resource["acl"].(string))),
		AccessKey:                    gofastly.String(resource["s3_access_key"].(string)),
		BucketName:                   gofastly.String(resource["bucket_name"].(string)),
		CompressionCodec:             gofastly.String(resource["compression_codec"].(string)),
		Domain:                       gofastly.String(resource["domain"].(string)),
		FileMaxBytes:                 gofastly.Int(resource["file_max_bytes"].(int)),
		Format:                       gofastly.String(vla.format),
		FormatVersion:                vla.formatVersion,
		IAMRole:                      gofastly.String(resource["s3_iam_role"].(string)),
		MessageType:                  gofastly.String(resource["message_type"].(string)),
		Name:                         gofastly.String(resource["name"].(string)),
		Path:                         gofastly.String(resource["path"].(string)),
		Period:                       gofastly.Int(resource["period"].(int)),
		PublicKey:                    gofastly.String(resource["public_key"].(string)),
		Redundancy:                   gofastly.S3RedundancyPtr(gofastly.S3Redundancy(resource["redundancy"].(string))),
		SecretKey:                    gofastly.String(resource["s3_secret_key"].(string)),
		ServerSideEncryptionKMSKeyID: gofastly.String(resource["server_side_encryption_kms_key_id"].(string)),
		ServiceID:                    serviceID,
		ServiceVersion:               serviceVersion,
		TimestampFormat:              gofastly.String(resource["timestamp_format"].(string)),
	}

	// NOTE: go-fastly v7+ expects a pointer, so TF can't set the zero type value.
	// If we set a default value for an attribute, then it will be sent to the API.
	// In some scenarios this can cause the API to reject the request.
	// For example, configuring compression_codec + gzip_level is invalid.
	if gl, ok := resource["gzip_level"].(int); ok && gl != -1 {
		opts.GzipLevel = gofastly.Int(gl)
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.placement != "" {
		opts.Placement = gofastly.String(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.String(vla.responseCondition)
	}

	encryption := resource["server_side_encryption"].(string)
	switch encryption {
	case string(gofastly.S3ServerSideEncryptionAES):
		opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionPtr(gofastly.S3ServerSideEncryptionAES)
	case string(gofastly.S3ServerSideEncryptionKMS):
		opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionPtr(gofastly.S3ServerSideEncryptionKMS)
	}

	return &opts, nil
}

func (h *S3LoggingServiceAttributeHandler) buildDelete(s3Map any, serviceID string, serviceVersion int) *gofastly.DeleteS3Input {
	resource := s3Map.(map[string]any)

	return &gofastly.DeleteS3Input{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
