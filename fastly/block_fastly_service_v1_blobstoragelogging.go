package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type BlobStorageLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceBlobStorageLogging(sa ServiceAttributes) ServiceAttributeDefinition {
	return &BlobStorageLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:               "blobstoragelogging",
			serviceAttributes: sa,
		},
	}
}

func (h *BlobStorageLoggingServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	obsl, nbsl := d.GetChange(h.GetKey())
	if obsl == nil {
		obsl = new(schema.Set)
	}
	if nbsl == nil {
		nbsl = new(schema.Set)
	}

	obsls := obsl.(*schema.Set)
	nbsls := nbsl.(*schema.Set)

	remove := obsls.Difference(nbsls).List()
	add := nbsls.Difference(obsls).List()

	// DELETE old Blob Storage logging configurations
	for _, bslRaw := range remove {
		bslf := bslRaw.(map[string]interface{})
		opts := gofastly.DeleteBlobStorageInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    bslf["name"].(string),
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
	}

	// POST new/updated Blob Storage logging configurations
	for _, bslRaw := range add {
		bslf := bslRaw.(map[string]interface{})

		// @HACK for a TF SDK Issue.
		//
		// This ensures that the required, `name`, field is present.
		//
		// If we have made it this far and `name` is not present, it is most-likely due
		// to a defunct diff as noted here - https://github.com/hashicorp/terraform-plugin-sdk/issues/160#issuecomment-522935697.
		//
		// This is caused by using a StateFunc in a nested TypeSet. While the StateFunc
		// properly handles setting state with the StateFunc, it returns extra entries
		// during state Gets, specifically `GetChange("blobstoragelogging")` in this case.
		if v, ok := bslf["name"]; !ok || v.(string) == "" {
			continue
		}

		var vla = NewVCLLoggingAttributes()
		if h.GetServiceAttributes().serviceType == ServiceTypeVCL {
			vla.format = bslf["format"].(string)
			vla.formatVersion = uint(bslf["format_version"].(int))
			vla.placement = bslf["placement"].(string)
			vla.responseCondition = bslf["response_condition"].(string)
		}

		opts := gofastly.CreateBlobStorageInput{
			Service:           d.Id(),
			Version:           latestVersion,
			Name:              bslf["name"].(string),
			Path:              bslf["path"].(string),
			AccountName:       bslf["account_name"].(string),
			Container:         bslf["container"].(string),
			SASToken:          bslf["sas_token"].(string),
			Period:            uint(bslf["period"].(int)),
			TimestampFormat:   bslf["timestamp_format"].(string),
			GzipLevel:         uint(bslf["gzip_level"].(int)),
			PublicKey:         bslf["public_key"].(string),
			MessageType:       bslf["message_type"].(string),
			Format:            vla.format,
			FormatVersion:     vla.formatVersion,
			Placement:         vla.placement,
			ResponseCondition: vla.responseCondition,
		}

		log.Printf("[DEBUG] Blob Storage logging create opts: %#v", opts)
		_, err := conn.CreateBlobStorage(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *BlobStorageLoggingServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Blob Storages for (%s)", d.Id())
	blobStorageList, err := conn.ListBlobStorages(&gofastly.ListBlobStoragesInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Blob Storages for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	bsl := flattenBlobStorages(blobStorageList)

	if err := d.Set(h.GetKey(), bsl); err != nil {
		log.Printf("[WARN] Error setting Blob Storages for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *BlobStorageLoggingServiceAttributeHandler) Register(s *schema.Resource) error {
	var a = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Azure Blob Storage logging endpoint",
		},
		"account_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique Azure Blob Storage namespace in which your data objects are stored",
		},
		"container": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Azure Blob Storage container in which to store logs",
		},
		"sas_token": {
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", ""),
			Description: "The Azure shared access signature providing write access to the blob service objects",
			Sensitive:   true,
		},
		// Optional fields
		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The path to upload logs to. Must end with a trailing slash",
		},
		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently the logs should be transferred, in seconds (default: 3600)",
		},
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: "strftime specified timestamp formatting (default: `%Y-%m-%dT%H:%M:%S.000`)",
		},
		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "The Gzip compression level (default: 0)",
		},
		"public_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},
		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted (default: `classic`)",
			ValidateFunc: validateLoggingMessageType(),
		},
	}

	if h.GetServiceAttributes().serviceType == ServiceTypeVCL {
		a["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%h %l %u %t \"%r\" %>s %b",
			Description: "Apache-style string or VCL variables to use for log formatting (default: `%h %l %u %t \"%r\" %>s %b`)",
		}
		a["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		a["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed",
			ValidateFunc: validateLoggingPlacement(),
		}
		a["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply",
		}
	}

	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: a,
		},
	}

	return nil
}

func flattenBlobStorages(blobStorageList []*gofastly.BlobStorage) []map[string]interface{} {
	var bsl []map[string]interface{}
	for _, bs := range blobStorageList {
		// Convert Blob Storages to a map for saving to state.
		nbs := map[string]interface{}{
			"name":               bs.Name,
			"path":               bs.Path,
			"account_name":       bs.AccountName,
			"container":          bs.Container,
			"sas_token":          bs.SASToken,
			"period":             bs.Period,
			"timestamp_format":   bs.TimestampFormat,
			"gzip_level":         bs.GzipLevel,
			"public_key":         bs.PublicKey,
			"format":             bs.Format,
			"format_version":     bs.FormatVersion,
			"message_type":       bs.MessageType,
			"placement":          bs.Placement,
			"response_condition": bs.ResponseCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nbs {
			if v == "" {
				delete(nbs, k)
			}
		}

		bsl = append(bsl, nbs)
	}

	return bsl
}
