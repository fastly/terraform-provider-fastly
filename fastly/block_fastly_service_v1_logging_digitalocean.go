package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type DigitalOceanServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingDigitalOcean(sa ServiceMetadata) ServiceAttributeDefinition {
	return &DigitalOceanServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_digitalocean",
			serviceMetadata: sa,
		},
	}
}

func (h *DigitalOceanServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	ol, nl := d.GetChange(h.GetKey())

	if ol == nil {
		ol = new(schema.Set)
	}
	if nl == nil {
		nl = new(schema.Set)
	}

	ols := ol.(*schema.Set)
	nls := nl.(*schema.Set)

	removeDigitalOceanLogging := ols.Difference(nls).List()
	addDigitalOceanLogging := nls.Difference(ols).List()

	// DELETE old DigitalOcean Spaces logging endpoints.
	for _, oRaw := range removeDigitalOceanLogging {
		of := oRaw.(map[string]interface{})
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly DigitalOcean Spaces logging endpoint removal opts: %#v", opts)

		if err := deleteDigitalOcean(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated DigitalOcean Spaces logging endpoints.
	for _, nRaw := range addDigitalOceanLogging {
		lf := nRaw.(map[string]interface{})

		// @HACK for a TF SDK Issue.
		//
		// This ensures that the required, `name`, field is present.
		//
		// If we have made it this far and `name` is not present, it is most-likely due
		// to a defunct diff as noted here - https://github.com/hashicorp/terraform-plugin-sdk/issues/160#issuecomment-522935697.
		//
		// This is caused by using a StateFunc in a nested TypeSet. While the StateFunc
		// properly handles setting state with the StateFunc, it returns extra entries
		// during state Gets, specifically `GetChange("logging_digitalocean")` in this case.
		if v, ok := lf["name"]; !ok || v.(string) == "" {
			continue
		}

		opts := h.buildCreate(lf, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly DigitalOcean Spaces logging addition opts: %#v", opts)

		if err := createDigitalOcean(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *DigitalOceanServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh DigitalOcean Spaces.
	log.Printf("[DEBUG] Refreshing DigitalOcean Spaces logging endpoints for (%s)", d.Id())
	digitaloceanList, err := conn.ListDigitalOceans(&gofastly.ListDigitalOceansInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up DigitalOcean Spaces logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenDigitalOcean(digitaloceanList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting DigitalOcean Spaces logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createDigitalOcean(conn *gofastly.Client, i *gofastly.CreateDigitalOceanInput) error {
	_, err := conn.CreateDigitalOcean(i)
	return err
}

func deleteDigitalOcean(conn *gofastly.Client, i *gofastly.DeleteDigitalOceanInput) error {
	err := conn.DeleteDigitalOcean(i)

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

func flattenDigitalOcean(digitaloceanList []*gofastly.DigitalOcean) []map[string]interface{} {
	var lsl []map[string]interface{}
	for _, ll := range digitaloceanList {
		// Convert DigitalOcean Spaces logging to a map for saving to state.
		nll := map[string]interface{}{
			"name":               ll.Name,
			"bucket_name":        ll.BucketName,
			"domain":             ll.Domain,
			"access_key":         ll.AccessKey,
			"secret_key":         ll.SecretKey,
			"public_key":         ll.PublicKey,
			"path":               ll.Path,
			"period":             ll.Period,
			"timestamp_format":   ll.TimestampFormat,
			"gzip_level":         ll.GzipLevel,
			"format":             ll.Format,
			"format_version":     ll.FormatVersion,
			"message_type":       ll.MessageType,
			"placement":          ll.Placement,
			"response_condition": ll.ResponseCondition,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range nll {
			if v == "" {
				delete(nll, k)
			}
		}

		lsl = append(lsl, nll)
	}

	return lsl
}

func (h *DigitalOceanServiceAttributeHandler) buildCreate(digitaloceanMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateDigitalOceanInput {
	df := digitaloceanMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateDigitalOceanInput{
		Service:           serviceID,
		Version:           serviceVersion,
		Name:              gofastly.NullString(df["name"].(string)),
		BucketName:        gofastly.NullString(df["bucket_name"].(string)),
		Domain:            gofastly.NullString(df["domain"].(string)),
		AccessKey:         gofastly.NullString(df["access_key"].(string)),
		SecretKey:         gofastly.NullString(df["secret_key"].(string)),
		PublicKey:         gofastly.NullString(df["public_key"].(string)),
		Path:              gofastly.NullString(df["path"].(string)),
		Period:            gofastly.Uint(uint(df["period"].(int))),
		GzipLevel:         gofastly.Uint(uint(df["gzip_level"].(int))),
		TimestampFormat:   gofastly.NullString(df["timestamp_format"].(string)),
		MessageType:       gofastly.NullString(df["message_type"].(string)),
		Format:            gofastly.NullString(vla.format),
		FormatVersion:     vla.formatVersion,
		Placement:         gofastly.NullString(vla.placement),
		ResponseCondition: gofastly.NullString(vla.responseCondition),
	}
}

func (h *DigitalOceanServiceAttributeHandler) buildDelete(digitaloceanMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteDigitalOceanInput {
	df := digitaloceanMap.(map[string]interface{})

	return &gofastly.DeleteDigitalOceanInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    df["name"].(string),
	}
}

func (h *DigitalOceanServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the DigitalOcean Spaces logging endpoint.",
		},

		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the DigitalOcean Space.",
		},

		"access_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Your DigitalOcean Spaces account access key.",
		},

		"secret_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Your DigitalOcean Spaces account secret key.",
		},

		// Optional fields
		"domain": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The domain of the DigitalOcean Spaces endpoint (default: nyc3.digitaloceanspaces.com).",
			Default:     "nyc3.digitaloceanspaces.com",
		},

		"public_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A PGP public key that Fastly will use to encrypt your log files before writing them to disk.",
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The path to upload logs to.",
		},

		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "How frequently log files are finalized so they can be available for reading (in seconds, default 3600).",
		},

		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "strftime specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`).",
		},

		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "What level of Gzip encoding to have when dumping logs (default 0, no compression).",
		},

		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted.",
			ValidateFunc: validateLoggingMessageType(),
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			Description:  "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.",
			ValidateFunc: validateLoggingPlacement(),
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
