package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type S3LoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceS3Logging() ServiceAttributeDefinition {
	return &S3LoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key: "s3logging",
		},
	}
}

func (h *S3LoggingServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	os, ns := d.GetChange("s3logging")
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)
	removeS3Logging := oss.Difference(nss).List()
	addS3Logging := nss.Difference(oss).List()

	// DELETE old S3 Log configurations
	for _, sRaw := range removeS3Logging {
		opts := buildDeleteS3(sRaw, d.Id(), latestVersion)
		err := deleteS3(conn, opts)
		if err != nil {
			return err
		}
	}

	// POST new/updated S3 Logging
	for _, sRaw := range addS3Logging {
		opts, _ := buildCreateS3(sRaw, d.Id(), latestVersion)
		err := createS3(conn, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *S3LoggingServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing S3 Logging for (%s)", d.Id())
	s3List, err := conn.ListS3s(&gofastly.ListS3sInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up S3 Logging for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	sl := flattenS3s(s3List)

	if err := d.Set("s3logging", sl); err != nil {
		log.Printf("[WARN] Error setting S3 Logging for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *S3LoggingServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unique name to refer to this logging setup",
				},
				"bucket_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "S3 Bucket name to store logs in",
				},
				"s3_access_key": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_ACCESS_KEY", ""),
					Description: "AWS Access Key",
					Sensitive:   true,
				},
				"s3_secret_key": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("FASTLY_S3_SECRET_KEY", ""),
					Description: "AWS Secret Key",
					Sensitive:   true,
				},
				// Optional fields
				"path": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Path to store the files. Must end with a trailing slash",
				},
				"domain": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Bucket endpoint",
					Default:     "s3.amazonaws.com",
				},
				"gzip_level": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     0,
					Description: "Gzip Compression level",
				},
				"period": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     3600,
					Description: "How frequently the logs should be transferred, in seconds (Default 3600)",
				},
				"format": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "%h %l %u %t %r %>s",
					Description: "Apache-style string or VCL variables to use for log formatting",
				},
				"format_version": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      1,
					Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (Default: 1)",
					ValidateFunc: validateLoggingFormatVersion(),
				},
				"timestamp_format": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "%Y-%m-%dT%H:%M:%S.000",
					Description: "specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)",
				},
				"redundancy": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The S3 redundancy level.",
				},
				"response_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of a condition to apply this logging.",
				},
				"message_type": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "classic",
					Description:  "How the message should be formatted.",
					ValidateFunc: validateLoggingMessageType(),
				},
				"placement": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Where in the generated VCL the logging call should be placed.",
					ValidateFunc: validateLoggingPlacement(),
				},
				"server_side_encryption": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Specify what type of server side encryption should be used. Can be either `AES256` or `aws:kms`.",
					ValidateFunc: validateLoggingServerSideEncryption(),
				},
				"server_side_encryption_kms_key_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Optional server-side KMS Key Id. Must be set if server_side_encryption is set to `aws:kms`",
				},
			},
		},
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

func flattenS3s(s3List []*gofastly.S3) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, s := range s3List {
		// Convert S3s to a map for saving to state.
		ns := map[string]interface{}{
			"name":                              s.Name,
			"bucket_name":                       s.BucketName,
			"s3_access_key":                     s.AccessKey,
			"s3_secret_key":                     s.SecretKey,
			"path":                              s.Path,
			"period":                            s.Period,
			"domain":                            s.Domain,
			"gzip_level":                        s.GzipLevel,
			"format":                            s.Format,
			"format_version":                    s.FormatVersion,
			"timestamp_format":                  s.TimestampFormat,
			"redundancy":                        s.Redundancy,
			"response_condition":                s.ResponseCondition,
			"message_type":                      s.MessageType,
			"public_key":                        s.PublicKey,
			"placement":                         s.Placement,
			"server_side_encryption":            s.ServerSideEncryption,
			"server_side_encryption_kms_key_id": s.ServerSideEncryptionKMSKeyID,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range ns {
			if v == "" {
				delete(ns, k)
			}
		}

		sl = append(sl, ns)
	}

	return sl
}

func buildCreateS3(s3Map interface{}, serviceID string, serviceVersion int) (*gofastly.CreateS3Input, error) {
	df := s3Map.(map[string]interface{})
	// The Fastly API will not error if these are omitted, so we throw an error
	// if any of these are empty.
	for _, sk := range []string{"s3_access_key", "s3_secret_key"} {
		if df[sk].(string) == "" {
			return nil, fmt.Errorf("[ERR] No %s found for S3 Log stream setup for Service (%s)", sk, serviceID)
		}
	}

	opts := gofastly.CreateS3Input{
		Service:                      serviceID,
		Version:                      serviceVersion,
		Name:                         df["name"].(string),
		BucketName:                   df["bucket_name"].(string),
		AccessKey:                    df["s3_access_key"].(string),
		SecretKey:                    df["s3_secret_key"].(string),
		Period:                       uint(df["period"].(int)),
		GzipLevel:                    uint(df["gzip_level"].(int)),
		Domain:                       df["domain"].(string),
		Path:                         df["path"].(string),
		Format:                       df["format"].(string),
		FormatVersion:                uint(df["format_version"].(int)),
		TimestampFormat:              df["timestamp_format"].(string),
		ResponseCondition:            df["response_condition"].(string),
		MessageType:                  df["message_type"].(string),
		PublicKey:                    df["public_key"].(string),
		Placement:                    df["placement"].(string),
		ServerSideEncryptionKMSKeyID: df["server_side_encryption_kms_key_id"].(string),
	}

	redundancy := strings.ToLower(df["redundancy"].(string))
	switch redundancy {
	case "standard":
		opts.Redundancy = gofastly.S3RedundancyStandard
	case "reduced_redundancy":
		opts.Redundancy = gofastly.S3RedundancyReduced
	}

	encryption := df["server_side_encryption"].(string)
	switch encryption {
	case string(gofastly.S3ServerSideEncryptionAES):
		opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionAES
	case string(gofastly.S3ServerSideEncryptionKMS):
		opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionKMS
	}

	return &opts, nil
}

func buildDeleteS3(s3Map interface{}, serviceID string, serviceVersion int) *gofastly.DeleteS3Input {
	df := s3Map.(map[string]interface{})

	return &gofastly.DeleteS3Input{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}
