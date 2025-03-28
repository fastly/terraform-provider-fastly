package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
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
			Description:      "Where in the generated VCL the logging call should be placed. Can be `none` or `none`.",
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
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Kinesis logging endpoints for (%s)", d.Id())
		remoteState, err := conn.ListKinesis(&gofastly.ListKinesisInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Kinesis logging endpoints for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		ell := flattenKinesis(remoteState)

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
		opts.StreamName = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["region"]; ok {
		opts.Region = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["access_key"]; ok {
		opts.AccessKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["secret_key"]; ok {
		opts.SecretKey = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["iam_role"]; ok {
		opts.IAMRole = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format"]; ok {
		opts.Format = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["format_version"]; ok {
		opts.FormatVersion = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["placement"]; ok {
		opts.Placement = gofastly.ToPointer(v.(string))
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
func flattenKinesis(remoteState []*gofastly.Kinesis) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.StreamName != nil {
			data["topic"] = *resource.StreamName
		}
		if resource.Region != nil {
			data["region"] = *resource.Region
		}
		if resource.AccessKey != nil {
			data["access_key"] = *resource.AccessKey
		}
		if resource.SecretKey != nil {
			data["secret_key"] = *resource.SecretKey
		}
		if resource.IAMRole != nil {
			data["iam_role"] = *resource.IAMRole
		}
		if resource.Format != nil {
			data["format"] = *resource.Format
		}
		if resource.FormatVersion != nil {
			data["format_version"] = *resource.FormatVersion
		}
		if resource.Placement != nil {
			data["placement"] = *resource.Placement
		}
		if resource.ResponseCondition != nil {
			data["response_condition"] = *resource.ResponseCondition
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
		AccessKey:      gofastly.ToPointer(resource["access_key"].(string)),
		Format:         gofastly.ToPointer(vla.format),
		FormatVersion:  vla.formatVersion,
		IAMRole:        gofastly.ToPointer(resource["iam_role"].(string)),
		Name:           gofastly.ToPointer(resource["name"].(string)),
		Region:         gofastly.ToPointer(resource["region"].(string)),
		SecretKey:      gofastly.ToPointer(resource["secret_key"].(string)),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		StreamName:     gofastly.ToPointer(resource["topic"].(string)),
	}

	// WARNING: The following fields shouldn't have an empty string passed.
	// As it will cause the Fastly API to return an error.
	// This is because go-fastly v7+ will not 'omitempty' due to pointer type.
	if vla.placement != "" {
		opts.Placement = gofastly.ToPointer(vla.placement)
	}
	if vla.responseCondition != "" {
		opts.ResponseCondition = gofastly.ToPointer(vla.responseCondition)
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
