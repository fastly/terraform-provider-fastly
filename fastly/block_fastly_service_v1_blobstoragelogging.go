package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type BlobStorageLoggingServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceBlobStorageLogging(sa ServiceMetadata) ServiceAttributeDefinition {
	return &BlobStorageLoggingServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "blobstoragelogging",
			serviceMetadata: sa,
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

	oldSet := obsl.(*schema.Set)
	newSet := nbsl.(*schema.Set)

	setDiff := NewSetDiff(func(resource interface{}) (interface{}, error) {
		t, ok := resource.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource failed to be type asserted: %+v", resource)
		}
		return t["name"], nil
	})

	diffResult, err := setDiff.Diff(oldSet, newSet)
	if err != nil {
		return err
	}

	// DELETE removed resources
	for _, resource := range diffResult.Deleted {
		resource := resource.(map[string]interface{})
		opts := gofastly.DeleteBlobStorageInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
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
	}

	// CREATE new resources
	for _, resource := range diffResult.Added {
		resource := resource.(map[string]interface{})

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
		if v, ok := resource["name"]; !ok || v.(string) == "" {
			continue
		}

		var vla = h.getVCLLoggingAttributes(resource)
		opts := gofastly.CreateBlobStorageInput{
			ServiceID:         d.Id(),
			ServiceVersion:    latestVersion,
			Name:              resource["name"].(string),
			Path:              resource["path"].(string),
			AccountName:       resource["account_name"].(string),
			Container:         resource["container"].(string),
			SASToken:          resource["sas_token"].(string),
			Period:            uint(resource["period"].(int)),
			TimestampFormat:   resource["timestamp_format"].(string),
			GzipLevel:         uint(resource["gzip_level"].(int)),
			PublicKey:         resource["public_key"].(string),
			MessageType:       resource["message_type"].(string),
			Format:            vla.format,
			FormatVersion:     uintOrDefault(vla.formatVersion),
			Placement:         vla.placement,
			ResponseCondition: vla.responseCondition,
		}

		log.Printf("[DEBUG] Blob Storage logging create opts: %#v", opts)
		_, err := conn.CreateBlobStorage(&opts)
		if err != nil {
			return err
		}
	}

	// UPDATE modified resources
	//
	// NOTE: although the go-fastly API client enables updating of a resource by
	// its 'name' attribute, this isn't possible within terraform due to
	// constraints in the data model/schema of the resources not having a uid.
	for _, resource := range diffResult.Modified {
		resource := resource.(map[string]interface{})

		opts := gofastly.UpdateBlobStorageInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
		}

		// only attempt to update attributes that have changed
		modified := setDiff.Filter(resource, oldSet)

		// NOTE: where we transition between interface{} we lose the ability to
		// infer the underlying type being either a uint vs an int. This
		// materializes as a panic (yay) and so it's only at runtime we discover
		// this and so we've updated the below code to convert the type asserted
		// int into a uint before passing the value to gofastly.Uint().
		if v, ok := modified["path"]; ok {
			opts.Path = gofastly.String(v.(string))
		}
		if v, ok := modified["account_name"]; ok {
			opts.AccountName = gofastly.String(v.(string))
		}
		if v, ok := modified["container"]; ok {
			opts.Container = gofastly.String(v.(string))
		}
		if v, ok := modified["sas_token"]; ok {
			opts.SASToken = gofastly.String(v.(string))
		}
		if v, ok := modified["period"]; ok {
			opts.Period = gofastly.Uint(uint(v.(int)))
		}
		if v, ok := modified["timestamp_format"]; ok {
			opts.TimestampFormat = gofastly.String(v.(string))
		}
		if v, ok := modified["compression_codec"]; ok {
			opts.CompressionCodec = gofastly.String(v.(string))
		}
		if v, ok := modified["gzip_level"]; ok {
			opts.GzipLevel = gofastly.Uint(uint(v.(int)))
		}
		if v, ok := modified["public_key"]; ok {
			opts.PublicKey = gofastly.String(v.(string))
		}
		if v, ok := modified["format"]; ok {
			opts.Format = gofastly.String(v.(string))
		}
		if v, ok := modified["format_version"]; ok {
			opts.FormatVersion = gofastly.Uint(uint(v.(int)))
		}
		if v, ok := modified["message_type"]; ok {
			opts.MessageType = gofastly.String(v.(string))
		}
		if v, ok := modified["placement"]; ok {
			opts.Placement = gofastly.String(v.(string))
		}
		if v, ok := modified["response_condition"]; ok {
			opts.ResponseCondition = gofastly.String(v.(string))
		}

		log.Printf("[DEBUG] Update Blob Storage Opts: %#v", opts)
		_, err := conn.UpdateBlobStorage(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *BlobStorageLoggingServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Blob Storages for (%s)", d.Id())
	blobStorageList, err := conn.ListBlobStorages(&gofastly.ListBlobStoragesInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
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
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A unique name to identify the Azure Blob Storage endpoint. It is important to note that changing this attribute will delete and recreate the resource",
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
			Description: "The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work",
			Sensitive:   true,
		},
		// Optional fields
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
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: "`strftime` specified timestamp formatting. Default `%Y-%m-%dT%H:%M:%S.000`",
		},
		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "Level of Gzip compression from `0-9`. `0` means no compression. `1` is the fastest and the least compressed version, `9` is the slowest and the most compressed version. Default `0`",
		},
		"public_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},
		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted. Can be either `classic`, `loggly`, `logplex` or `blank`. Default `classic`",
			ValidateFunc: validateLoggingMessageType(),
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
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either 1 or 2. (default: 2)",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed",
			ValidateFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the condition to apply",
		}
	}

	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
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
