package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
)

// ResponseObjectServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type ResponseObjectServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceResponseObject returns a new resource.
func NewServiceResponseObject(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&ResponseObjectServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "response_object",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *ResponseObjectServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *ResponseObjectServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cache_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to check after we have retrieved an object. If the condition passes then deliver this Request Object instead. This `condition` must be of type `CACHE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals](https://docs.fastly.com/en/guides/using-conditions)",
				},
				"content": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "The content to deliver for the response object",
				},
				"content_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "The MIME type of the content",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A unique name to identify this Response Object. It is important to note that changing this attribute will delete and recreate the resource",
				},
				"request_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to be checked during the request phase. If the condition passes then this object will be delivered. This `condition` must be of type `REQUEST`",
				},
				"response": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "OK",
					Description: "The HTTP Response. Default `OK`",
				},
				"status": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     200,
					Description: "The HTTP Status Code. Default `200`",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *ResponseObjectServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateResponseObjectInput{
		ServiceID:        d.Id(),
		ServiceVersion:   serviceVersion,
		Name:             gofastly.ToPointer(resource["name"].(string)),
		Status:           gofastly.ToPointer(resource["status"].(int)),
		Response:         gofastly.ToPointer(resource["response"].(string)),
		Content:          gofastly.ToPointer(resource["content"].(string)),
		ContentType:      gofastly.ToPointer(resource["content_type"].(string)),
		RequestCondition: gofastly.ToPointer(resource["request_condition"].(string)),
		CacheCondition:   gofastly.ToPointer(resource["cache_condition"].(string)),
	}

	log.Printf("[DEBUG] Create Response Object Opts: %#v", opts)
	_, err := conn.CreateResponseObject(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *ResponseObjectServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Response Object for (%s)", d.Id())
		remoteState, err := conn.ListResponseObjects(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListResponseObjectsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Response Object for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		rol := flattenResponseObjects(remoteState)

		if err := d.Set(h.GetKey(), rol); err != nil {
			log.Printf("[WARN] Error setting Response Object for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *ResponseObjectServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateResponseObjectInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["status"]; ok {
		opts.Status = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["response"]; ok {
		opts.Response = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["content"]; ok {
		opts.Content = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["content_type"]; ok {
		opts.ContentType = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["request_condition"]; ok {
		opts.RequestCondition = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["cache_condition"]; ok {
		opts.CacheCondition = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Response Object Opts: %#v", opts)
	_, err := conn.UpdateResponseObject(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *ResponseObjectServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteResponseObjectInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Response Object removal opts: %#v", opts)
	err := conn.DeleteResponseObject(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenResponseObjects models data into format suitable for saving to Terraform state.
func flattenResponseObjects(remoteState []*gofastly.ResponseObject) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Status != nil {
			data["status"] = *resource.Status
		}
		if resource.Response != nil {
			data["response"] = *resource.Response
		}
		if resource.Content != nil {
			data["content"] = *resource.Content
		}
		if resource.ContentType != nil {
			data["content_type"] = *resource.ContentType
		}
		if resource.RequestCondition != nil {
			data["request_condition"] = *resource.RequestCondition
		}
		if resource.CacheCondition != nil {
			data["cache_condition"] = *resource.CacheCondition
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
