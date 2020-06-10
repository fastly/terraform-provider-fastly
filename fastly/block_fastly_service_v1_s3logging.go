package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var s3loggingSchema = &schema.Schema{
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

func processS3Logging(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
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
		sf := sRaw.(map[string]interface{})
		opts := gofastly.DeleteS3Input{
			Service: d.Id(),
			Version: latestVersion,
			Name:    sf["name"].(string),
		}

		log.Printf("[DEBUG] Fastly S3 Logging removal opts: %#v", opts)
		err := conn.DeleteS3(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated S3 Logging
	for _, sRaw := range addS3Logging {
		sf := sRaw.(map[string]interface{})

		// Fastly API will not error if these are omitted, so we throw an error
		// if any of these are empty
		for _, sk := range []string{"s3_access_key", "s3_secret_key"} {
			if sf[sk].(string) == "" {
				return fmt.Errorf("[ERR] No %s found for S3 Log stream setup for Service (%s)", sk, d.Id())
			}
		}

		opts := gofastly.CreateS3Input{
			Service:                      d.Id(),
			Version:                      latestVersion,
			Name:                         sf["name"].(string),
			BucketName:                   sf["bucket_name"].(string),
			AccessKey:                    sf["s3_access_key"].(string),
			SecretKey:                    sf["s3_secret_key"].(string),
			Period:                       uint(sf["period"].(int)),
			GzipLevel:                    uint(sf["gzip_level"].(int)),
			Domain:                       sf["domain"].(string),
			Path:                         sf["path"].(string),
			Format:                       sf["format"].(string),
			FormatVersion:                uint(sf["format_version"].(int)),
			TimestampFormat:              sf["timestamp_format"].(string),
			ResponseCondition:            sf["response_condition"].(string),
			MessageType:                  sf["message_type"].(string),
			Placement:                    sf["placement"].(string),
			ServerSideEncryptionKMSKeyID: sf["server_side_encryption_kms_key_id"].(string),
		}

		redundancy := strings.ToLower(sf["redundancy"].(string))
		switch redundancy {
		case "standard":
			opts.Redundancy = gofastly.S3RedundancyStandard
		case "reduced_redundancy":
			opts.Redundancy = gofastly.S3RedundancyReduced
		}

		encryption := sf["server_side_encryption"].(string)
		switch encryption {
		case string(gofastly.S3ServerSideEncryptionAES):
			opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionAES
		case string(gofastly.S3ServerSideEncryptionKMS):
			opts.ServerSideEncryption = gofastly.S3ServerSideEncryptionKMS
		}

		log.Printf("[DEBUG] Create S3 Logging Opts: %#v", opts)
		_, err := conn.CreateS3(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}