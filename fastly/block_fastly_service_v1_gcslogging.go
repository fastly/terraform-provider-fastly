package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var gcsloggingSchema = &schema.Schema{
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
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_EMAIL", ""),
				Description: "The email address associated with the target GCS bucket on your account.",
			},
			"bucket_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the bucket in which to store the logs.",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("FASTLY_GCS_SECRET_KEY", ""),
				Description: "The secret key associated with the target gcs bucket on your account.",
				Sensitive:   true,
			},
			// Optional fields
			"path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to store the files. Must end with a trailing slash",
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
			"timestamp_format": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "%Y-%m-%dT%H:%M:%S.000",
				Description: "specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)",
			},
			"response_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of a condition to apply this logging.",
			},
			"message_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "classic",
				Description: "The log message type per the fastly docs: https://developer.fastly.com/reference/api/logging/gcs/",
			},
			"placement": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Where in the generated VCL the logging call should be placed.",
				ValidateFunc: validateLoggingPlacement(),
			},
		},
	},
}


func processGCSLogging(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	os, ns := d.GetChange("gcslogging")
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)
	removeGcslogging := oss.Difference(nss).List()
	addGcslogging := nss.Difference(oss).List()

	// DELETE old gcslogging configurations
	for _, pRaw := range removeGcslogging {
		sf := pRaw.(map[string]interface{})
		opts := gofastly.DeleteGCSInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    sf["name"].(string),
		}

		log.Printf("[DEBUG] Fastly gcslogging removal opts: %#v", opts)
		err := conn.DeleteGCS(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated gcslogging
	for _, pRaw := range addGcslogging {
		sf := pRaw.(map[string]interface{})
		opts := gofastly.CreateGCSInput{
			Service:           d.Id(),
			Version:           latestVersion,
			Name:              sf["name"].(string),
			User:              sf["email"].(string),
			Bucket:            sf["bucket_name"].(string),
			SecretKey:         sf["secret_key"].(string),
			Format:            sf["format"].(string),
			Path:              sf["path"].(string),
			Period:            uint(sf["period"].(int)),
			GzipLevel:         uint8(sf["gzip_level"].(int)),
			TimestampFormat:   sf["timestamp_format"].(string),
			MessageType:       sf["message_type"].(string),
			ResponseCondition: sf["response_condition"].(string),
			Placement:         sf["placement"].(string),
		}

		log.Printf("[DEBUG] Create GCS Opts: %#v", opts)
		_, err := conn.CreateGCS(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}


func readGCSLogging(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	log.Printf("[DEBUG] Refreshing GCS for (%s)", d.Id())
	GCSList, err := conn.ListGCSs(&gofastly.ListGCSsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up GCS for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	gcsl := flattenGCS(GCSList)
	if err := d.Set("gcslogging", gcsl); err != nil {
		log.Printf("[WARN] Error setting gcs for (%s): %s", d.Id(), err)
	}

	return nil
}


func flattenGCS(gcsList []*gofastly.GCS) []map[string]interface{} {
	var GCSList []map[string]interface{}
	for _, currentGCS := range gcsList {
		// Convert gcs to a map for saving to state.
		GCSMapString := map[string]interface{}{
			"name":               currentGCS.Name,
			"email":              currentGCS.User,
			"bucket_name":        currentGCS.Bucket,
			"secret_key":         currentGCS.SecretKey,
			"path":               currentGCS.Path,
			"period":             int(currentGCS.Period),
			"gzip_level":         int(currentGCS.GzipLevel),
			"response_condition": currentGCS.ResponseCondition,
			"message_type":       currentGCS.MessageType,
			"format":             currentGCS.Format,
			"timestamp_format":   currentGCS.TimestampFormat,
			"placement":          currentGCS.Placement,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range GCSMapString {
			if v == "" {
				delete(GCSMapString, k)
			}
		}

		GCSList = append(GCSList, GCSMapString)
	}

	return GCSList
}