package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// HeaderServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type HeaderServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceHeader returns a new resource.
func NewServiceHeader(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&HeaderServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "header",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *HeaderServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *HeaderServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"action": {
					Type:             schema.TypeString,
					Required:         true,
					Description:      "The Header manipulation action to take; must be one of `set`, `append`, `delete`, `regex`, or `regex_repeat`",
					ValidateDiagFunc: validateHeaderAction(),
				},
				"cache_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to apply. This `condition` must be of type `CACHE`",
				},
				"destination": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the header that is going to be affected by the Action",
				},
				"ignore_if_set": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Don't add the header if it is already. (Only applies to `set` action.). Default `false`",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unique name for this header attribute. It is important to note that changing this attribute will delete and recreate the resource",
				},
				"priority": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     100,
					Description: "Lower priorities execute first. Default: `100`",
				},
				"regex": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Regular expression to use (Only applies to `regex` and `regex_repeat` actions.)",
				},
				"request_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to apply. This `condition` must be of type `REQUEST`",
				},
				"response_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals](https://docs.fastly.com/en/guides/using-conditions)",
				},
				"source": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Variable to be used as a source for the header content (Does not apply to `delete` action.)",
				},
				"substitution": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Value to substitute in place of regular expression. (Only applies to `regex` and `regex_repeat`.)",
				},
				"type": {
					Type:             schema.TypeString,
					Required:         true,
					Description:      "The Request type on which to apply the selected Action; must be one of `request`, `fetch`, `cache` or `response`",
					ValidateDiagFunc: validateHeaderType(),
				},
			},
		},
	}
}

// Create creates the resource.
func (h *HeaderServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts, err := buildHeader(resource)
	if err != nil {
		log.Printf("[DEBUG] Error building Header: %s", err)
		return err
	}
	opts.ServiceID = d.Id()
	opts.ServiceVersion = serviceVersion

	log.Printf("[DEBUG] Fastly Header Addition opts: %#v", opts)
	_, err = conn.CreateHeader(opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *HeaderServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Headers for (%s)", d.Id())
		remoteState, err := conn.ListHeaders(&gofastly.ListHeadersInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Headers for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		hl := flattenHeaders(remoteState)

		if err := d.Set(h.GetKey(), hl); err != nil {
			log.Printf("[WARN] Error setting Headers for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *HeaderServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateHeaderInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["action"]; ok {
		opts.Action = gofastly.HeaderActionPtr(gofastly.HeaderAction(v.(string)))
	}
	if v, ok := modified["ignore_if_set"]; ok {
		opts.IgnoreIfSet = gofastly.CBool(v.(bool))
	}
	if v, ok := modified["type"]; ok {
		opts.Type = gofastly.HeaderTypePtr(gofastly.HeaderType(v.(string)))
	}
	if v, ok := modified["destination"]; ok {
		opts.Destination = gofastly.String(v.(string))
	}
	if v, ok := modified["source"]; ok {
		opts.Source = gofastly.String(v.(string))
	}
	if v, ok := modified["regex"]; ok {
		opts.Regex = gofastly.String(v.(string))
	}
	if v, ok := modified["substitution"]; ok {
		opts.Substitution = gofastly.String(v.(string))
	}
	if v, ok := modified["priority"]; ok {
		opts.Priority = gofastly.Int(v.(int))
	}
	if v, ok := modified["request_condition"]; ok {
		opts.RequestCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["cache_condition"]; ok {
		opts.CacheCondition = gofastly.String(v.(string))
	}
	if v, ok := modified["response_condition"]; ok {
		opts.ResponseCondition = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Header Opts: %#v", opts)
	_, err := conn.UpdateHeader(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *HeaderServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteHeaderInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Header removal opts: %#v", opts)
	err := conn.DeleteHeader(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenHeaders models data into format suitable for saving to Terraform state.
func flattenHeaders(remoteState []*gofastly.Header) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{
			"name":               resource.Name,
			"action":             resource.Action,
			"ignore_if_set":      resource.IgnoreIfSet,
			"type":               resource.Type,
			"destination":        resource.Destination,
			"source":             resource.Source,
			"regex":              resource.Regex,
			"substitution":       resource.Substitution,
			"priority":           int(resource.Priority),
			"request_condition":  resource.RequestCondition,
			"cache_condition":    resource.CacheCondition,
			"response_condition": resource.ResponseCondition,
		}

		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}
	return result
}

func buildHeader(headerMap any) (*gofastly.CreateHeaderInput, error) {
	resource := headerMap.(map[string]any)
	opts := gofastly.CreateHeaderInput{
		Name:              gofastly.String(resource["name"].(string)),
		IgnoreIfSet:       gofastly.CBool(resource["ignore_if_set"].(bool)),
		Destination:       gofastly.String(resource["destination"].(string)),
		Priority:          gofastly.Int(resource["priority"].(int)),
		Source:            gofastly.String(resource["source"].(string)),
		Regex:             gofastly.String(resource["regex"].(string)),
		Substitution:      gofastly.String(resource["substitution"].(string)),
		RequestCondition:  gofastly.String(resource["request_condition"].(string)),
		CacheCondition:    gofastly.String(resource["cache_condition"].(string)),
		ResponseCondition: gofastly.String(resource["response_condition"].(string)),
	}

	act := strings.ToLower(resource["action"].(string))
	switch act {
	case "set":
		opts.Action = gofastly.HeaderActionPtr(gofastly.HeaderActionSet)
	case "append":
		opts.Action = gofastly.HeaderActionPtr(gofastly.HeaderActionAppend)
	case "delete":
		opts.Action = gofastly.HeaderActionPtr(gofastly.HeaderActionDelete)
	case "regex":
		opts.Action = gofastly.HeaderActionPtr(gofastly.HeaderActionRegex)
	case "regex_repeat":
		opts.Action = gofastly.HeaderActionPtr(gofastly.HeaderActionRegexRepeat)
	}

	ty := strings.ToLower(resource["type"].(string))
	switch ty {
	case "request":
		opts.Type = gofastly.HeaderTypePtr(gofastly.HeaderTypeRequest)
	case "fetch":
		opts.Type = gofastly.HeaderTypePtr(gofastly.HeaderTypeFetch)
	case "cache":
		opts.Type = gofastly.HeaderTypePtr(gofastly.HeaderTypeCache)
	case "response":
		opts.Type = gofastly.HeaderTypePtr(gofastly.HeaderTypeResponse)
	}

	return &opts, nil
}
