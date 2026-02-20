package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
)

// ResourceLinkServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type ResourceLinkServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceResourceLink returns a new resource.
func NewServiceResourceLink(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&ResourceLinkServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "resource_link",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *ResourceLinkServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *ResourceLinkServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "A resource link represents a link between a shared resource (such as an KV Store or Config Store) and a service version.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"link_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "An alphanumeric string identifying the resource link.",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the resource link.",
				},
				"resource_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The ID of the underlying linked resource.",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *ResourceLinkServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	input := &gofastly.CreateResourceInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		ResourceID:     gofastly.ToPointer(resource["resource_id"].(string)),
	}

	log.Printf("[DEBUG] CREATE: Resource Links input: %#v", input)

	_, err := conn.CreateResource(gofastly.NewContextForResourceID(ctx, d.Id()), input)
	if err != nil {
		return err
	}

	return nil
}

// Read refreshes the resource.
func (h *ResourceLinkServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		input := &gofastly.ListResourcesInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		}

		log.Printf("[DEBUG] REFRESH: Resource Links input: %#v", input)

		remoteState, err := conn.ListResources(gofastly.NewContextForResourceID(ctx, d.Id()), input)
		if err != nil {
			return err
		}

		data := flattenResourceLinks(remoteState)
		if err := d.Set(h.GetKey(), data); err != nil {
			log.Printf("[WARN] REFRESH: Error setting Resource Links for Service ID '%s': %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *ResourceLinkServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	input := &gofastly.UpdateResourceInput{
		ResourceID:     resource["link_id"].(string),
		Name:           gofastly.ToPointer(resource["name"].(string)),
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
	}

	log.Printf("[DEBUG] UPDATE: Resource Links input: %#v", input)

	_, err := conn.UpdateResource(gofastly.NewContextForResourceID(ctx, d.Id()), input)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the resource.
func (h *ResourceLinkServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	input := &gofastly.DeleteResourceInput{
		ResourceID:     resource["link_id"].(string),
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
	}

	log.Printf("[DEBUG] DELETE: Resource Links input: %#v", input)

	err := conn.DeleteResource(gofastly.NewContextForResourceID(ctx, d.Id()), input)
	if err != nil {
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			// If error is because the resource looks to already be deleted (i.e. 404),
			// then skip returning the error and allow it to fail silently.
			if errRes.StatusCode != 404 {
				return err
			}
		}
	}

	return err
}

// flattenResourceLinks models data into format suitable for saving to Terraform state.
func flattenResourceLinks(remoteState []*gofastly.Resource) []map[string]any {
	result := make([]map[string]any, 0, len(remoteState))

	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.LinkID != nil {
			data["link_id"] = *resource.LinkID
		}
		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.ResourceID != nil {
			data["resource_id"] = *resource.ResourceID
		}

		result = append(result, data)
	}

	return result
}
