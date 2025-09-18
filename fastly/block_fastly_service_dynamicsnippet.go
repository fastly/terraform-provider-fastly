package fastly

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

// DynamicSnippetServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type DynamicSnippetServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceDynamicSnippet returns a new resource.
func NewServiceDynamicSnippet(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&DynamicSnippetServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "dynamicsnippet",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *DynamicSnippetServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *DynamicSnippetServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"content": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
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
				"snippet_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the dynamic snippet",
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
func (h *DynamicSnippetServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := buildDynamicSnippet(resource)

	opts.ServiceID = d.Id()
	opts.ServiceVersion = serviceVersion

	log.Printf("[DEBUG] Fastly VCL Dynamic Snippet Addition opts: %#v", opts)
	_, err := conn.CreateSnippet(gofastly.NewContextForResourceID(ctx, d.Id()), opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *DynamicSnippetServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing VCL Snippets for (%s)", d.Id())
		remoteState, err := conn.ListSnippets(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListSnippetsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up VCL Snippets for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		dynamicSnippets := flattenDynamicSnippets(remoteState)
		if err := d.Set(h.GetKey(), dynamicSnippets); err != nil {
			log.Printf("[WARN] Error setting VCL Dynamic Snippets for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *DynamicSnippetServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateSnippetInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["priority"]; ok {
		opts.Priority = gofastly.ToPointer(strconv.Itoa(v.(int)))
	}
	if v, ok := modified["content"]; ok {
		opts.Content = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["type"]; ok {
		opts.Type = gofastly.ToPointer(gofastly.SnippetType(v.(string)))
	}

	log.Printf("[DEBUG] Update Dynamic Snippet Opts: %#v", opts)
	_, err := conn.UpdateSnippet(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *DynamicSnippetServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteSnippetInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly VCL Dynamic Snippet Removal opts: %#v", opts)
	err := conn.DeleteSnippet(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func buildDynamicSnippet(dynamicSnippetMap any) *gofastly.CreateSnippetInput {
	resource := dynamicSnippetMap.(map[string]any)
	opts := gofastly.CreateSnippetInput{
		Content:  gofastly.ToPointer(resource["content"].(string)),
		Dynamic:  gofastly.ToPointer(1),
		Name:     gofastly.ToPointer(resource["name"].(string)),
		Priority: gofastly.ToPointer(strconv.Itoa(resource["priority"].(int))),
	}

	snippetType := strings.ToLower(resource["type"].(string))
	opts.Type = gofastly.ToPointer(gofastly.SnippetType(snippetType))

	return &opts
}

// flattenDynamicSnippets models data into format suitable for saving to Terraform state.
func flattenDynamicSnippets(remoteState []*gofastly.Snippet) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		// Skip non-dynamic snippets
		if resource.Dynamic != nil && *resource.Dynamic == 0 {
			continue
		}

		data := map[string]any{}

		if resource.SnippetID != nil {
			data["snippet_id"] = *resource.SnippetID
		}
		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Type != nil {
			data["type"] = *resource.Type
		}
		if resource.Priority != nil {
			p, _ := strconv.Atoi(*resource.Priority)
			data["priority"] = p
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
