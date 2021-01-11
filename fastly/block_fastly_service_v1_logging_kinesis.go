package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type KinesisServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceLoggingKinesis(sa ServiceMetadata) ServiceAttributeDefinition {
	return &KinesisServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_kinesis",
			serviceMetadata: sa,
		},
	}
}

func (h *KinesisServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
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

	removeKinesisLogging := ols.Difference(nls).List()
	addKinesisLogging := nls.Difference(ols).List()

	// DELETE old Kinesis logging endpoints.
	for _, oRaw := range removeKinesisLogging {
		of := oRaw.(map[string]interface{})
		opts := h.buildDelete(of, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Kinesis logging endpoint removal opts: %#v", opts)

		if err := deleteKinesis(conn, opts); err != nil {
			return err
		}
	}

	// POST new/updated Kinesis logging endpoints.
	for _, nRaw := range addKinesisLogging {
		lf := nRaw.(map[string]interface{})
		opts := h.buildCreate(lf, serviceID, latestVersion)

		log.Printf("[DEBUG] Fastly Kinesis logging addition opts: %#v", opts)

		if err := createKinesis(conn, opts); err != nil {
			return err
		}
	}

	return nil
}

func (h *KinesisServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	// Refresh Kinesis.
	log.Printf("[DEBUG] Refreshing Kinesis logging endpoints for (%s)", d.Id())
	kinesisList, err := conn.ListKineses(&gofastly.ListKinesesInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Kinesis logging endpoints for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	ell := flattenKinesis(kinesisList)

	if err := d.Set(h.GetKey(), ell); err != nil {
		log.Printf("[WARN] Error setting Kinesis logging endpoints for (%s): %s", d.Id(), err)
	}

	return nil
}

func createKinesis(conn *gofastly.Client, i *gofastly.CreateKinesisInput) error {
	_, err := conn.CreateKinesis(i)
	return err
}

func deleteKinesis(conn *gofastly.Client, i *gofastly.DeleteKinesisInput) error {
	err := conn.DeleteKinesis(i)

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

func flattenKinesis(kinesisList []*gofastly.Kinesis) []map[string]interface{} {
	var lsl []map[string]interface{}
	for _, ll := range kinesisList {
		// Convert Kinesis logging to a map for saving to state.
		nll := map[string]interface{}{
			"name":               ll.Name,
			"topic":              ll.StreamName,
			"region":             ll.Region,
			"access_key":         ll.AccessKey,
			"secret_key":         ll.SecretKey,
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

func (h *KinesisServiceAttributeHandler) buildCreate(kinesisMap interface{}, serviceID string, serviceVersion int) *gofastly.CreateKinesisInput {
	df := kinesisMap.(map[string]interface{})

	var vla = h.getVCLLoggingAttributes(df)
	return &gofastly.CreateKinesisInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              df["name"].(string),
		StreamName:        df["topic"].(string),
		Region:            df["region"].(string),
		AccessKey:         df["access_key"].(string),
		SecretKey:         df["secret_key"].(string),
		Format:            vla.format,
		FormatVersion:     uintOrDefault(vla.formatVersion),
		Placement:         vla.placement,
		ResponseCondition: vla.responseCondition,
	}
}

func (h *KinesisServiceAttributeHandler) buildDelete(kinesisMap interface{}, serviceID string, serviceVersion int) *gofastly.DeleteKinesisInput {
	df := kinesisMap.(map[string]interface{})

	return &gofastly.DeleteKinesisInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}

func (h *KinesisServiceAttributeHandler) Register(s *schema.Resource) error {
	var blockAttributes = map[string]*schema.Schema{
		// Required fields
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Kinesis logging endpoint.",
		},

		"topic": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Kinesis stream name.",
		},

		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "us-east-1",
			Description: "The AWS region the stream resides in.",
		},

		"access_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The AWS access key to be used to write to the stream.",
		},

		"secret_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "The AWS secret access key to authenticate with.",
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
