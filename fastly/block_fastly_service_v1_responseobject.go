package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ResponseObjectServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceResponseObject() ServiceAttributeDefinition {
	return &ResponseObjectServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			schema: responseobjectSchema,
			key:    "response_object",
		},
	}
}

var responseobjectSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name to refer to this request object",
			},
			// Optional fields
			"status": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     200,
				Description: "The HTTP Status Code of the object",
			},
			"response": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "OK",
				Description: "The HTTP Response of the object",
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
				Description: "Name of the condition to be checked during the request phase to see if the object should be delivered",
			},
			"cache_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of the condition checked after we have retrieved an object. If the condition passes then deliver this Request Object instead.",
			},
		},
	},
}

func (h *ResponseObjectServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	or, nr := d.GetChange("response_object")
	if or == nil {
		or = new(schema.Set)
	}
	if nr == nil {
		nr = new(schema.Set)
	}

	ors := or.(*schema.Set)
	nrs := nr.(*schema.Set)
	removeResponseObject := ors.Difference(nrs).List()
	addResponseObject := nrs.Difference(ors).List()

	// DELETE old response object configurations
	for _, rRaw := range removeResponseObject {
		rf := rRaw.(map[string]interface{})
		opts := gofastly.DeleteResponseObjectInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    rf["name"].(string),
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
	}

	// POST new/updated Response Object
	for _, rRaw := range addResponseObject {
		rf := rRaw.(map[string]interface{})

		opts := gofastly.CreateResponseObjectInput{
			Service:          d.Id(),
			Version:          latestVersion,
			Name:             rf["name"].(string),
			Status:           uint(rf["status"].(int)),
			Response:         rf["response"].(string),
			Content:          rf["content"].(string),
			ContentType:      rf["content_type"].(string),
			RequestCondition: rf["request_condition"].(string),
			CacheCondition:   rf["cache_condition"].(string),
		}

		log.Printf("[DEBUG] Create Response Object Opts: %#v", opts)
		_, err := conn.CreateResponseObject(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *ResponseObjectServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Response Object for (%s)", d.Id())
	responseObjectList, err := conn.ListResponseObjects(&gofastly.ListResponseObjectsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Response Object for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	rol := flattenResponseObjects(responseObjectList)

	if err := d.Set("response_object", rol); err != nil {
		log.Printf("[WARN] Error setting Response Object for (%s): %s", d.Id(), err)
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
