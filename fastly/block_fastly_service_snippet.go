package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SnippetServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type SnippetServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceSnippet returns a new resource.
func NewServiceSnippet(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&SnippetServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "snippet",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *SnippetServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *SnippetServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"content": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The VCL code that specifies exactly what the snippet does",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: `A name that is unique across "regular" and "dynamic" VCL Snippet configuration blocks. It is important to note that changing this attribute will delete and recreate the resource`,
				},
				"priority": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     100,
					Description: "Priority determines the ordering for multiple snippets. Lower numbers execute first. Defaults to `100`",
				},
				"type": {
					Type:             schema.TypeString,
					Required:         true,
					Description:      SnippetTypeDescription,
					ValidateDiagFunc: validateSnippetType(),
				},
			},
		},
	}
}

// Create creates the resource.
func (h *SnippetServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts, err := buildSnippet(resource)
	if err != nil {
		log.Printf("[DEBUG] Error building VCL Snippet: %s", err)
		return err
	}
	opts.ServiceID = d.Id()
	opts.ServiceVersion = serviceVersion

	log.Printf("[DEBUG] Fastly VCL Snippet Addition opts: %#v", opts)
	_, err = conn.CreateSnippet(opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *SnippetServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.Key()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing VCL Snippets for (%s)", d.Id())
		remoteState, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up VCL Snippets for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		vsl := flattenSnippets(remoteState)

		if err := d.Set(h.GetKey(), vsl); err != nil {
			log.Printf("[WARN] Error setting VCL Snippets for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *SnippetServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	// Safety check in case keys aren't actually set in the HCL.
	name, _ := resource["name"].(string)
	priority, _ := resource["priority"].(int)
	content, _ := resource["content"].(string)
	stype, _ := resource["type"].(string)

	opts := gofastly.UpdateSnippetInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           name,
		NewName:        gofastly.ToPointer(name),
		Priority:       gofastly.ToPointer(priority),
		Content:        gofastly.ToPointer(content),
		Type:           gofastly.ToPointer(gofastly.SnippetType(stype)),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["priority"]; ok {
		opts.Priority = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["content"]; ok {
		opts.Content = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["type"]; ok {
		snippetType := strings.ToLower(v.(string))
		opts.Type = gofastly.ToPointer(gofastly.SnippetType(snippetType))
	}

	log.Printf("[DEBUG] Update VCL Snippet Opts: %#v", opts)
	_, err := conn.UpdateSnippet(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *SnippetServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteSnippetInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly VCL Snippet Removal opts: %#v", opts)
	err := conn.DeleteSnippet(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func buildSnippet(snippetMap any) (*gofastly.CreateSnippetInput, error) {
	resource := snippetMap.(map[string]any)
	opts := gofastly.CreateSnippetInput{
		Name:     gofastly.ToPointer(resource["name"].(string)),
		Content:  gofastly.ToPointer(resource["content"].(string)),
		Priority: gofastly.ToPointer(resource["priority"].(int)),
		Dynamic:  gofastly.ToPointer(0),
	}

	snippetType := strings.ToLower(resource["type"].(string))
	opts.Type = gofastly.ToPointer(gofastly.SnippetType(snippetType))

	return &opts, nil
}

// flattenSnippets models data into format suitable for saving to Terraform state.
func flattenSnippets(remoteState []*gofastly.Snippet) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		// Skip dynamic snippets
		if resource.Dynamic != nil && *resource.Dynamic == 1 {
			continue
		}

		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Type != nil {
			data["type"] = *resource.Type
		}
		if resource.Priority != nil {
			data["priority"] = *resource.Priority
		}
		if resource.Content != nil {
			data["content"] = *resource.Content
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
