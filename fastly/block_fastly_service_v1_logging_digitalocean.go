package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	oldSet := ol.(*schema.Set)
	newSet := nl.(*schema.Set)

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
		opts := h.buildDelete(resource, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly DigitalOcean Spaces logging endpoint removal opts: %#v", opts)

		if err := deleteDigitalOcean(conn, opts); err != nil {
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
		// during state Gets, specifically `GetChange("logging_digitalocean")` in this case.
		if v, ok := resource["name"]; !ok || v.(string) == "" {
			continue
		}

		opts := h.buildCreate(resource, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly DigitalOcean Spaces logging addition opts: %#v", opts)

		if err := createDigitalOcean(conn, opts); err != nil {
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

		opts := gofastly.UpdateDigitalOceanInput{
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
		if v, ok := modified["bucket_name"]; ok {
			opts.BucketName = gofastly.String(v.(string))
		}
		if v, ok := modified["domain"]; ok {
			opts.Domain = gofastly.String(v.(string))
		}
		if v, ok := modified["access_key"]; ok {
			opts.AccessKey = gofastly.String(v.(string))
		}
		if v, ok := modified["secret_key"]; ok {
			opts.SecretKey = gofastly.String(v.(string))
		}
		if v, ok := modified["path"]; ok {
			opts.Path = gofastly.String(v.(string))
		}
		if v, ok := modified["period"]; ok {
			opts.Period = gofastly.Uint(uint(v.(int)))
		}
		if v, ok := modified["gzip_level"]; ok {
			opts.GzipLevel = gofastly.Uint(uint(v.(int)))
		}
		if v, ok := modified["format"]; ok {
			opts.Format = gofastly.String(v.(string))
		}
		if v, ok := modified["format_version"]; ok {
			opts.FormatVersion = gofastly.Uint(uint(v.(int)))
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
		if v, ok := modified["placement"]; ok {
			opts.Placement = gofastly.String(v.(string))
		}
		if v, ok := modified["public_key"]; ok {
			opts.PublicKey = gofastly.String(v.(string))
		}

		log.Printf("[DEBUG] Update DigitalOcean Opts: %#v", opts)
		_, err := conn.UpdateDigitalOcean(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *DigitalOceanServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh DigitalOcean Spaces.
	log.Printf("[DEBUG] Refreshing DigitalOcean Spaces logging endpoints for (%s)", d.Id())
	digitaloceanList, err := conn.ListDigitalOceans(&gofastly.ListDigitalOceansInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up DigitalOcean Spaces logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenDigitalOcean(digitaloceanList)

	for _, element := range ell {
		element = h.pruneVCLLoggingAttributes(element)
	}

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
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              df["name"].(string),
		BucketName:        df["bucket_name"].(string),
		Domain:            df["domain"].(string),
		AccessKey:         df["access_key"].(string),
		SecretKey:         df["secret_key"].(string),
		PublicKey:         df["public_key"].(string),
		Path:              df["path"].(string),
		Period:            uint(df["period"].(int)),
		GzipLevel:         uint(df["gzip_level"].(int)),
		TimestampFormat:   df["timestamp_format"].(string),
		MessageType:       df["message_type"].(string),
		Format:            vla.format,
		FormatVersion:     uintOrDefault(vla.formatVersion),
		Placement:         vla.placement,
		ResponseCondition: vla.responseCondition,
	}
}

func (h *DigitalOceanServiceAttributeHandler) buildDelete(digitaloceanMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteDigitalOceanInput {
	df := digitaloceanMap.(map[string]interface{})

	return &gofastly.DeleteDigitalOceanInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}

func (h *DigitalOceanServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the DigitalOcean Spaces logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},

		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the DigitalOcean Space",
		},

		"access_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Your DigitalOcean Spaces account access key",
		},

		"secret_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Your DigitalOcean Spaces account secret key",
		},

		// Optional fields
		"domain": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The domain of the DigitalOcean Spaces endpoint (default `nyc3.digitaloceanspaces.com`)",
			Default:     "nyc3.digitaloceanspaces.com",
		},

		"public_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The path to upload logs to",
		},

		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "How frequently log files are finalized so they can be available for reading (in seconds, default `3600`)",
		},

		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "`strftime` specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)",
		},

		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "What level of Gzip encoding to have when dumping logs (default `0`, no compression)",
		},

		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted. One of: `classic` (default), `loggly`, `logplex` or `blank`",
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
