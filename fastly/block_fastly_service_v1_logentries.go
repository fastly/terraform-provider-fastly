package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var logentriesSchema = &schema.Schema{
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
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Use token based authentication (https://logentries.com/doc/input-token/)",
			},
			// Optional
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     20000,
				Description: "The port number configured in Logentries",
			},
			"use_tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to use TLS for secure logging",
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
			"response_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of a condition to apply this logging.",
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


func processLogEntries(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	os, ns := d.GetChange("logentries")
	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)
	removeLogentries := oss.Difference(nss).List()
	addLogentries := nss.Difference(oss).List()

	// DELETE old logentries configurations
	for _, pRaw := range removeLogentries {
		slf := pRaw.(map[string]interface{})
		opts := gofastly.DeleteLogentriesInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    slf["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Logentries removal opts: %#v", opts)
		err := conn.DeleteLogentries(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated Logentries
	for _, pRaw := range addLogentries {
		slf := pRaw.(map[string]interface{})

		opts := gofastly.CreateLogentriesInput{
			Service:           d.Id(),
			Version:           latestVersion,
			Name:              slf["name"].(string),
			Port:              uint(slf["port"].(int)),
			UseTLS:            gofastly.CBool(slf["use_tls"].(bool)),
			Token:             slf["token"].(string),
			Format:            slf["format"].(string),
			FormatVersion:     uint(slf["format_version"].(int)),
			ResponseCondition: slf["response_condition"].(string),
			Placement:         slf["placement"].(string),
		}

		log.Printf("[DEBUG] Create Logentries Opts: %#v", opts)
		_, err := conn.CreateLogentries(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}
