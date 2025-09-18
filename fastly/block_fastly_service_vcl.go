package fastly

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

// VCLServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type VCLServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceVCL returns a new resource.
func NewServiceVCL(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&VCLServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "vcl",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *VCLServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *VCLServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"content": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The custom VCL code to upload",
				},
				"main": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "If `true`, use this block as the main configuration. If `false`, use this block as an includable library. Only a single VCL block can be marked as the main block. Default is `false`",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A unique name for this configuration block. It is important to note that changing this attribute will delete and recreate the resource",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *VCLServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateVCLInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		Content:        gofastly.ToPointer(resource["content"].(string)),
		Main:           gofastly.ToPointer(resource["main"].(bool)),
	}

	log.Printf("[DEBUG] Fastly VCL Addition opts: %#v", opts)
	_, err := conn.CreateVCL(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *VCLServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.Key()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing VCLs for (%s)", d.Id())
		remoteState, err := conn.ListVCLs(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListVCLsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up VCLs for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		vl := flattenVCLs(remoteState)

		if err := d.Set(h.GetKey(), vl); err != nil {
			log.Printf("[WARN] Error setting VCLs for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *VCLServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateVCLInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	if v, ok := modified["content"]; ok {
		opts.Content = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update VCL Opts: %#v", opts)
	_, err := conn.UpdateVCL(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *VCLServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteVCLInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly VCL Removal opts: %#v", opts)
	err := conn.DeleteVCL(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenVCLs models data into format suitable for saving to Terraform state.
func flattenVCLs(remoteState []*gofastly.VCL) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Content != nil {
			data["content"] = *resource.Content
		}
		if resource.Main != nil {
			data["main"] = *resource.Main
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
}

func validateVCLs(d *schema.ResourceData) error {
	// TODO: this would be nice to move into a resource/collection validation function, once that is available
	// (see https://github.com/hashicorp/terraform/pull/4348 and https://github.com/hashicorp/terraform/pull/6508)
	vcls, exists := d.GetOk("vcl")
	if !exists {
		return nil
	}

	numberOfMainVCLs, numberOfIncludeVCLs := 0, 0
	for _, vclElem := range vcls.(*schema.Set).List() {
		vcl := vclElem.(map[string]any)
		if mainVal, hasMain := vcl["main"]; hasMain && mainVal.(bool) {
			numberOfMainVCLs++
		} else {
			numberOfIncludeVCLs++
		}
	}
	if numberOfMainVCLs == 0 && numberOfIncludeVCLs > 0 {
		return errors.New("if you include VCL configurations, one of them should have main = true")
	}
	if numberOfMainVCLs > 1 {
		return errors.New("you cannot have more than one VCL configuration with main = true")
	}
	return nil
}
