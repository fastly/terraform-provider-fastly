package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type OpenstackServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingOpenstack(sa ServiceMetadata) ServiceAttributeDefinition {
	return &OpenstackServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_openstack",
			serviceMetadata: sa,
		},
	}
}

func (h *OpenstackServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
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

	removeOpenstackLogging := ols.Difference(nls).List()
	addOpenstackLogging := nls.Difference(ols).List()

	// DELETE old OpenStack logging endpoints.
	for _, oRaw := range removeOpenstackLogging {
		of := oRaw.(map[string]interface{})
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly OpenStack logging endpoint removal opts: %#v", opts)

		if err := deleteOpenstack(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated OpenStack logging endpoints.
	for _, nRaw := range addOpenstackLogging {
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
		// during state Gets, specifically `GetChange("logging_openstack")` in this case.
		if v, ok := lf["name"]; !ok || v.(string) == "" {
			continue
		}

		opts := h.buildCreate(lf, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly OpenStack logging addition opts: %#v", opts)

		if err := createOpenstack(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *OpenstackServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh OpenStack.
	log.Printf("[DEBUG] Refreshing OpenStack logging endpoints for (%s)", d.Id())
	openstackList, err := conn.ListOpenstack(&gofastly.ListOpenstackInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up OpenStack logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenOpenstack(openstackList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting OpenStack logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createOpenstack(conn *gofastly.Client, i *gofastly.CreateOpenstackInput) error {
	_, err := conn.CreateOpenstack(i)
	return err
}

func deleteOpenstack(conn *gofastly.Client, i *gofastly.DeleteOpenstackInput) error {
	err := conn.DeleteOpenstack(i)

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

func flattenOpenstack(openstackList []*gofastly.Openstack) []map[string]interface{} {
	var lsl []map[string]interface{}
	for _, ll := range openstackList {
		// Convert OpenStack logging to a map for saving to state.
		nll := map[string]interface{}{
			"name":               ll.Name,
			"url":                ll.URL,
			"user":               ll.User,
			"bucket_name":        ll.BucketName,
			"access_key":         ll.AccessKey,
			"public_key":         ll.PublicKey,
			"gzip_level":         ll.GzipLevel,
			"message_type":       ll.MessageType,
			"path":               ll.Path,
			"period":             ll.Period,
			"timestamp_format":   ll.TimestampFormat,
			"format":             ll.Format,
			"format_version":     ll.FormatVersion,
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

func (h *OpenstackServiceAttributeHandler) buildCreate(openstackMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateOpenstackInput {
	df := openstackMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateOpenstackInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              df["name"].(string),
		URL:               df["url"].(string),
		User:              df["user"].(string),
		BucketName:        df["bucket_name"].(string),
		AccessKey:         df["access_key"].(string),
		PublicKey:         df["public_key"].(string),
		GzipLevel:         uint(df["gzip_level"].(int)),
		MessageType:       df["message_type"].(string),
		Path:              df["path"].(string),
		Period:            uint(df["period"].(int)),
		TimestampFormat:   df["timestamp_format"].(string),
		Format:            vla.format,
		FormatVersion:     *vla.formatVersion,
		Placement:         vla.placement,
		ResponseCondition: vla.responseCondition,
	}
}

func (h *OpenstackServiceAttributeHandler) buildDelete(openstackMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteOpenstackInput {
	df := openstackMap.(map[string]interface{})

	return &gofastly.DeleteOpenstackInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}

func (h *OpenstackServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the OpenStack logging endpoint.",
		},

		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Your OpenStack auth url.",
		},

		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The username for your OpenStack account.",
		},

		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of your OpenStack container.",
		},

		"access_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Your OpenStack account access key.",
		},

		// Optional fields
		"public_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A PGP public key that Fastly will use to encrypt your log files before writing them to disk.",
			// Related issue for weird behavior - https://github.com/hashicorp/terraform-plugin-sdk/issues/160
			StateFunc: trimSpaceStateFunc,
		},

		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "What level of Gzip encoding to have when dumping logs (default 0, no compression).",
		},

		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently log files are finalized so they can be available for reading (in seconds, default 3600).",
		},

		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The path to upload logs to.",
		},

		"message_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "classic",
			Description:  "How the message should be formatted. One of: classic (default), loggly, logplex or blank.",
			ValidateFunc: validateLoggingMessageType(),
		},

		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: "specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)",
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
		blockAttributes["placement"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.",
			ValidateFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
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
