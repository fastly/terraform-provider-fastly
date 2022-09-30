package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				// Required
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A unique name to identify this Response Object. It is important to note that changing this attribute will delete and recreate the resource",
				},
				// Optional fields
				"status": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     200,
					Description: "The HTTP Status Code. Default `200`",
				},
				"response": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "OK",
					Description: "The HTTP Response. Default `OK`",
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
				"request_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to be checked during the request phase. If the condition passes then this object will be delivered. This `condition` must be of type `REQUEST`",
				},
				"cache_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to check after we have retrieved an object. If the condition passes then deliver this Request Object instead. This `condition` must be of type `CACHE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals](https://docs.fastly.com/en/guides/using-conditions)",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *ResponseObjectServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateResponseObjectInput{
		ServiceID:        d.Id(),
		ServiceVersion:   serviceVersion,
		Name:             resource["name"].(string),
		Status:           gofastly.Uint(uint(resource["status"].(int))),
		Response:         resource["response"].(string),
		Content:          resource["content"].(string),
		ContentType:      resource["content_type"].(string),
		RequestCondition: resource["request_condition"].(string),
		CacheCondition:   resource["cache_condition"].(string),
	}

	log.Printf("[DEBUG] Create Response Object Opts: %#v", opts)
	_, err := conn.CreateResponseObject(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *ResponseObjectServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Response Object for (%s)", d.Id())
		responseObjectList, err := conn.ListResponseObjects(&gofastly.ListResponseObjectsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Response Object for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		rol := flattenResponseObjects(responseObjectList)

		if err := d.Set(h.GetKey(), rol); err != nil {
			log.Printf("[WARN] Error setting Response Object for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *ResponseObjectServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateResponseObjectInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: where we transition between interface{} we lose the ability to
	// infer the underlying type being either a uint vs an int. This
	// materializes as a panic (yay) and so it's only at runtime we discover
	// this and so we've updated the below code to convert the type asserted
	// int into a uint before passing the value to gofastly.Uint().
	if v, ok := modified["status"]; ok {
		opts.Status = gofastly.Uint(uint(v.(int)))
	}
	if v, ok := modified["response"]; ok {
		opts.Response = gofastly.String(v.(string))
	}
	if v, ok := modified["content"]; ok {
		opts.Content = gofastly.String(v.(string))
	}
	if v, ok := modified["content_type"]; ok {
		opts.ContentType = gofastly.String(v.(string))
	}
	if v, ok := modified["request_condition"]; ok {
		opts.RequestCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["cache_condition"]; ok {
		opts.CacheCondition = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Response Object Opts: %#v", opts)
	_, err := conn.UpdateResponseObject(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *ResponseObjectServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]interface{}, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteResponseObjectInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Response Object removal opts: %#v", opts)
	err := conn.DeleteResponseObject(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func flattenResponseObjects(responseObjectList []*gofastly.ResponseObject) []map[string]interface{} {
	var rol []map[string]interface{}
	for _, ro := range responseObjectList {
		// Convert ResponseObjects to a map for saving to state.
		nro := map[string]interface{}{
			"name":              ro.Name,
			"status":            ro.Status,
			"response":          ro.Response,
			"content":           ro.Content,
			"content_type":      ro.ContentType,
			"request_condition": ro.RequestCondition,
			"cache_condition":   ro.CacheCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nro {
			if v == "" {
				delete(nro, k)
			}
		}

		rol = append(rol, nro)
	}

	return rol
}
