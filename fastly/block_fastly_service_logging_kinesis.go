package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// KinesisServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type KinesisServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceLoggingKinesis returns a new resource.
func NewServiceLoggingKinesis(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&KinesisServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "logging_kinesis",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *KinesisServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *KinesisServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"access_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "The AWS access key to be used to write to the stream",
		},
		"iam_role": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Amazon Resource Name (ARN) for the IAM role granting Fastly access to Kinesis. Not required if `access_key` and `secret_key` are provided.",
			Sensitive:   false,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique name of the Kinesis logging endpoint. It is important to note that changing this attribute will delete and recreate the resource",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "us-east-1",
			Description: "The AWS region the stream resides in. (Default: `us-east-1`)",
		},
		"secret_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "The AWS secret access key to authenticate with",
		},
		"topic": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Kinesis stream name",
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
func (h *KinesisServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildCreate(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Kinesis logging addition opts: %#v", opts)

	return createKinesis(conn, opts)
}

// Read refreshes the resource.
func (h *KinesisServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Kinesis logging endpoints for (%s)", d.Id())
		kinesisList, err := conn.ListKinesis(&gofastly.ListKinesisInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Kinesis logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenKinesis(kinesisList)

		for _, element := range ell {
			h.pruneVCLLoggingAttributes(element)
		}

		if err := d.Set(h.GetKey(), ell); err != nil {
			log.Printf("[WARN] Error setting Kinesis logging endpoints for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *KinesisServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateKinesisInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["topic"]; ok {
		opts.StreamName = gofastly.String(v.(string))
	}
	if v, ok := modified["region"]; ok {
		opts.Region = gofastly.String(v.(string))
	}
	if v, ok := modified["access_key"]; ok {
		opts.AccessKey = gofastly.String(v.(string))
	}
	if v, ok := modified["secret_key"]; ok {
		opts.SecretKey = gofastly.String(v.(string))
	}
	if v, ok := modified["iam_role"]; ok {
		opts.IAMRole = gofastly.String(v.(string))
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
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Kinesis Opts: %#v", opts)
	_, err := conn.UpdateKinesis(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *KinesisServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := h.buildDelete(resource, d.Id(), serviceVersion)

	log.Printf("[DEBUG] Fastly Kinesis logging endpoint removal opts: %#v", opts)

	return deleteKinesis(conn, opts)
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

// flattenKinesis models data into format suitable for saving to Terraform state.
func flattenKinesis(kinesisList []*gofastly.Kinesis) []map[string]any {
	var result []map[string]any
	for _, resource := range kinesisList {
		data := map[string]any{
			"name":               resource.Name,
			"topic":              resource.StreamName,
			"region":             resource.Region,
			"access_key":         resource.AccessKey,
			"secret_key":         resource.SecretKey,
			"iam_role":           resource.IAMRole,
			"format":             resource.Format,
			"format_version":     resource.FormatVersion,
			"placement":          resource.Placement,
			"response_condition": resource.ResponseCondition,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
}

func (h *KinesisServiceAttributeHandler) buildCreate(kinesisMap any, serviceID string, serviceVersion int) *gofastly.CreateKinesisInput {
	resource := kinesisMap.(map[string]any)

	vla := h.getVCLLoggingAttributes(resource)
	opts := &gofastly.CreateKinesisInput{
		AccessKey:      gofastly.String(resource["access_key"].(string)),
		Format:         gofastly.String(vla.format),
		FormatVersion:  vla.formatVersion,
		IAMRole:        gofastly.String(resource["iam_role"].(string)),
		Name:           gofastly.String(resource["name"].(string)),
		Region:         gofastly.String(resource["region"].(string)),
		SecretKey:      gofastly.String(resource["secret_key"].(string)),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		StreamName:     gofastly.String(resource["topic"].(string)),
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.placement != "" {
		opts.Placement = gofastly.String(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.String(vla.responseCondition)
	}

	return opts
}

func (h *KinesisServiceAttributeHandler) buildDelete(kinesisMap any, serviceID string, serviceVersion int) *gofastly.DeleteKinesisInput {
	resource := kinesisMap.(map[string]any)

	return &gofastly.DeleteKinesisInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}
}
