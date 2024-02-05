package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ConditionServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type ConditionServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceCondition returns a new resource.
func NewServiceCondition(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&ConditionServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "condition",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *ConditionServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *ConditionServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The unique name for the condition. It is important to note that changing this attribute will delete and recreate the resource",
				},
				"priority": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     10,
					Description: "A number used to determine the order in which multiple conditions execute. Lower numbers execute first. Default `10`",
				},
				"statement": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The statement used to determine if the condition is met",
				},
				"type": {
					Type:             schema.TypeString,
					Required:         true,
					Description:      "Type of condition, either `REQUEST` (req), `RESPONSE` (req, resp), or `CACHE` (req, beresp)",
					ValidateDiagFunc: validateConditionType(),
				},
			},
		},
	}
}

// Create creates the resource.
func (h *ConditionServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateConditionInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		Type:           gofastly.ToPointer(resource["type"].(string)),
		// need to trim leading/tailing spaces, incase the config has HEREDOC
		// formatting and contains a trailing new line
		Statement: gofastly.ToPointer(strings.TrimSpace(resource["statement"].(string))),
		Priority:  gofastly.ToPointer(resource["priority"].(int)),
	}

	log.Printf("[DEBUG] Create Conditions Opts: %#v", opts)
	_, err := conn.CreateCondition(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *ConditionServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Conditions for (%s)", d.Id())
		remoteState, err := conn.ListConditions(&gofastly.ListConditionsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Conditions for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		cl := flattenConditions(remoteState)

		if err := d.Set(h.GetKey(), cl); err != nil {
			log.Printf("[WARN] Error setting Conditions for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *ConditionServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	optsCreate := gofastly.CreateConditionInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		Type:           gofastly.ToPointer(resource["type"].(string)),
		Statement:      gofastly.ToPointer(strings.TrimSpace(resource["statement"].(string))),
		Priority:       gofastly.ToPointer(resource["priority"].(int)),
	}

	optsUpdate := gofastly.UpdateConditionInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["comment"]; ok {
		optsUpdate.Comment = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["statement"]; ok {
		optsCreate.Statement = gofastly.ToPointer(v.(string))
		optsUpdate.Statement = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["priority"]; ok {
		optsCreate.Priority = gofastly.ToPointer(v.(int))
		optsUpdate.Priority = gofastly.ToPointer(v.(int))
	}
	// NOTE: Fastly API doesn't support updating the condition "type".
	// Therefore, we need to DELETE and CREATE if "type" attribute is changed.
	if v, ok := modified["type"]; ok {
		optsCreate.Type = gofastly.ToPointer(v.(string))
		log.Printf("[DEBUG] Delete Condition: %s (type changed)", resource["name"].(string))
		err := conn.DeleteCondition(&gofastly.DeleteConditionInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
			Name:           resource["name"].(string),
		})
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Create Condition Opts: %#v", optsCreate)
		_, err = conn.CreateCondition(&optsCreate)
		if err != nil {
			return err
		}
		return nil
	}

	log.Printf("[DEBUG] Update Condition Opts: %#v", optsUpdate)
	_, err := conn.UpdateCondition(&optsUpdate)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *ConditionServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteConditionInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Conditions Removal opts: %#v", opts)
	err := conn.DeleteCondition(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenConditions models data into format suitable for saving to Terraform state.
func flattenConditions(remoteState []*gofastly.Condition) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Statement != nil {
			data["statement"] = *resource.Statement
		}
		if resource.Type != nil {
			data["type"] = *resource.Type
		}
		if resource.Priority != nil {
			data["priority"] = *resource.Priority
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
