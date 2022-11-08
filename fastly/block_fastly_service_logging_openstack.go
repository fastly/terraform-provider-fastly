package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// OpenstackServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type OpenstackServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingOpenstack returns a new resource.
func NewServiceLoggingOpenstack(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&OpenstackServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_openstack",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *OpenstackServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *OpenstackServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"access_key": {
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			Description: "Your OpenStack account access key",
		},
		"bucket_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of your OpenStack container",
		},
		"compression_codec": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`,
			ValidateDiagFunc: validateLoggingCompressionCodec(),
		},
		"gzip_level": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: GzipLevelDescription,
		},
		"message_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "classic",
			Description:      MessageTypeDescription,
			ValidateDiagFunc: validateLoggingMessageType(),
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the OpenStack logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Path to store the files. Must end with a trailing slash. If this field is left empty, the files will be saved in the bucket's root path",
		},
		"period": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     3600,
			Description: "How frequently the logs should be transferred, in seconds. Default `3600`",
		},
		"public_key": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "A PGP public key that Fastly will use to encrypt your log files before writing them to disk",
			ValidateDiagFunc: validateStringTrimmed,
		},
		"timestamp_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "%Y-%m-%dT%H:%M:%S.000",
			Description: TimestampFormatDescription,
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Your OpenStack auth url",
		},
		"user": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The username for your OpenStack account",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["format"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Apache style log formatting.",
		}
		blockAttributes["format_version"] = &schema.Schema{
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          2,
			Description:      "The version of the custom logging format used for the configured endpoint. Can be either `1` or `2`. (default: `2`).",
			ValidateDiagFunc: validateLoggingFormatVersion(),
		}
		blockAttributes["placement"] = &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Where in the generated VCL the logging call should be placed. Can be `none` or `waf_debug`.",
			ValidateDiagFunc: validateLoggingPlacement(),
		}
		blockAttributes["response_condition"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
}

// Create creates the resource.
func (h *OpenstackServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly OpenStack logging addition opts: %#v", opts)

	return createOpenstack(conn, opts)
}

// Read refreshes the resource.
func (h *OpenstackServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing OpenStack logging endpoints for (%s)", d.Id())
		openstackList, err := conn.ListOpenstack(&gofastly.ListOpenstackInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up OpenStack logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenOpenstack(openstackList)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting OpenStack logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *OpenstackServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateOpenstackInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["access_key"]; ok {
		opts.AccessKey = gofastly.String(v.(string))
	}
	if v, ok := modified["bucket_name"]; ok {
		opts.BucketName = gofastly.String(v.(string))
	}
	if v, ok := modified["url"]; ok {
		opts.URL = gofastly.String(v.(string))
	}
	if v, ok := modified["user"]; ok {
		opts.User = gofastly.String(v.(string))
	}
	if v, ok := modified["path"]; ok {
		opts.Path = gofastly.String(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}
	if v, ok := modified["period"]; ok {
		opts.Period = gofastly.Int(v.(int))
	}
	if v, ok := modified["compression_codec"]; ok {
		opts.CompressionCodec = gofastly.String(v.(string))
	}
	if v, ok := modified["gzip_level"]; ok {
		opts.GzipLevel = gofastly.Int(v.(int))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.String(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.Int(v.(int))
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
	if v, ok := modified["public_key"]; ok {
		opts.PublicKey = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update OpenStack Opts: %#v", opts)
	_, err := conn.UpdateOpenstack(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *OpenstackServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly OpenStack logging endpoint removal opts: %#v", opts)

	return deleteOpenstack(conn, opts)
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

func flattenOpenstack(openstackList []*gofastly.Openstack) []map[string]any {
	var lsl []map[string]any
	for _, ll := range openstackList {
		// Convert OpenStack logging to a map for saving to state.
		nll := map[string]any{
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
			"compression_codec":  ll.CompressionCodec,
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

func (h *OpenstackServiceAttributeHandler) buildCreate(openstackMap any, serviceID string, serviceVersion int) *gofastly.CreateOpenstackInput {
	df := openstackMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(df)
	return &gofastly.CreateOpenstackInput{
		ServiceID:         serviceID,
		ServiceVersion:    serviceVersion,
		Name:              gofastly.String(df["name"].(string)),
		URL:               gofastly.String(df["url"].(string)),
		User:              gofastly.String(df["user"].(string)),
		BucketName:        gofastly.String(df["bucket_name"].(string)),
		AccessKey:         gofastly.String(df["access_key"].(string)),
		PublicKey:         gofastly.String(df["public_key"].(string)),
		GzipLevel:         gofastly.Int(df["gzip_level"].(int)),
		MessageType:       gofastly.String(df["message_type"].(string)),
		Path:              gofastly.String(df["path"].(string)),
		Period:            gofastly.Int(df["period"].(int)),
		TimestampFormat:   gofastly.String(df["timestamp_format"].(string)),
		CompressionCodec:  gofastly.String(df["compression_codec"].(string)),
		Format:            gofastly.String(vla.format),
		FormatVersion:     gofastly.Int(intOrDefault(vla.formatVersion)),
		Placement:         gofastly.String(vla.placement),
		ResponseCondition: gofastly.String(vla.responseCondition),
	}
}

func (h *OpenstackServiceAttributeHandler) buildDelete(openstackMap any, serviceID string, serviceVersion int) *gofastly.DeleteOpenstackInput {
	df := openstackMap.(map[string]any)

	return &gofastly.DeleteOpenstackInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           df["name"].(string),
	}
}
