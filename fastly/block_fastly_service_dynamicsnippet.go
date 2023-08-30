package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
func (h *DynamicSnippetServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts, err := buildDynamicSnippet(resource)
	if err != nil {
		log.Printf("[DEBUG] Error building VCL Dynamic Snippet: %s", err)
		return err
	}
	opts.ServiceID = d.Id()
	opts.ServiceVersion = serviceVersion

	log.Printf("[DEBUG] Fastly VCL Dynamic Snippet Addition opts: %#v", opts)
	_, err = conn.CreateSnippet(opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *DynamicSnippetServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing VCL Snippets for (%s)", d.Id())
		remoteState, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up VCL Snippets for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		// The ListSnippets endpoint doesn't return the 'content' field for
		// dynamic snippets and so we have to make a separate API call so that we
		// can then store the VCL content into the state file.
		for _, snippet := range remoteState {
			if snippet.Dynamic == 0 {
				continue
			}

			s, err := conn.GetDynamicSnippet(&gofastly.GetDynamicSnippetInput{
				ServiceID: d.Id(),
				ID:        snippet.ID,
			})
			if err != nil {
				return fmt.Errorf("error looking up VCL Dynamic Snippet '%s': %s", snippet.Name, err)
			}
			snippet.Content = s.Content
		}

		dynamicSnippets := flattenDynamicSnippets(remoteState)

		fmt.Printf("remoteState: %#v\n", remoteState[0])
		fmt.Printf("dynamicSnippets: %#v\n", dynamicSnippets)

		if err := d.Set(h.GetKey(), dynamicSnippets); err != nil {
			log.Printf("[WARN] Error setting VCL Dynamic Snippets for (%s): %s", d.Id(), err)
		}

		updatedState := d.Get(h.GetKey()).(*schema.Set)
		fmt.Printf("updatedState: %#v\n", updatedState)
	}

	return nil
}

// Update updates the resource.
func (h *DynamicSnippetServiceAttributeHandler) Update(
	_ context.Context,
	d *schema.ResourceData,
	resource, modified map[string]any,
	_ int, // service version
	conn *gofastly.Client,
) error {
	localState := d.Get(h.GetKey()).(*schema.Set)
	fmt.Printf("localState: %#v\n", localState)
	fmt.Printf("d.Get(dynamicsnippet): %#v\n", d.Get("dynamicsnippet"))
	fmt.Printf("resource: %#v\n", resource)
	fmt.Printf("modified: %#v\n", modified)
	opts := gofastly.UpdateDynamicSnippetInput{
		ServiceID: d.Id(),
		ID:        resource["snippet_id"].(string), // FIXME: "" as it's computed!
	}

	if v, ok := modified["content"]; ok {
		opts.Content = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Dynamic Snippet Opts: %#v", opts)
	_, err := conn.UpdateDynamicSnippet(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *DynamicSnippetServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteSnippetInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly VCL Dynamic Snippet Removal opts: %#v", opts)
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

func buildDynamicSnippet(dynamicSnippetMap any) (*gofastly.CreateSnippetInput, error) {
	resource := dynamicSnippetMap.(map[string]any)
	opts := gofastly.CreateSnippetInput{
		Content:  gofastly.String(resource["content"].(string)),
		Dynamic:  gofastly.Int(1),
		Name:     gofastly.String(resource["name"].(string)),
		Priority: gofastly.Int(resource["priority"].(int)),
	}

	snippetType := strings.ToLower(resource["type"].(string))
	opts.Type = gofastly.SnippetTypePtr(gofastly.SnippetType(snippetType))

	return &opts, nil
}

// flattenDynamicSnippets models data into format suitable for saving to Terraform state.
func flattenDynamicSnippets(remoteState []*gofastly.Snippet) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		// Skip non-dynamic snippets
		if resource.Dynamic == 0 {
			continue
		}

		data := map[string]any{
			"snippet_id": resource.ID,
			"name":       resource.Name,
			"type":       resource.Type,
			"priority":   int(resource.Priority),
			"content":    resource.Content,
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
