package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ConditionServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceCondition(sa ServiceMetadata) ServiceAttributeDefinition {
	return &ConditionServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "condition",
			serviceMetadata: sa,
		},
	}
}

func (h *ConditionServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	// Note: we don't utilize the PUT endpoint to update these objects, we simply
	// destroy any that have changed, and create new ones with the updated
	// values. This is how Terraform works with nested sub resources, we only
	// get the full diff not a partial set item diff. Because this is done
	// on a new version of the Fastly Service configuration, this is considered safe

	oc, nc := d.GetChange(h.GetKey())
	if oc == nil {
		oc = new(schema.Set)
	}
	if nc == nil {
		nc = new(schema.Set)
	}

	ocs := oc.(*schema.Set)
	ncs := nc.(*schema.Set)
	removeConditions := ocs.Difference(ncs).List()
	addConditions := ncs.Difference(ocs).List()

	// DELETE old Conditions
	for _, cRaw := range removeConditions {
		cf := cRaw.(map[string]interface{})
		opts := gofastly.DeleteConditionInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    cf["name"].(string),
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
	}

	// POST new Conditions
	for _, cRaw := range addConditions {
		cf := cRaw.(map[string]interface{})
		opts := gofastly.CreateConditionInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    cf["name"].(string),
			Type:    cf["type"].(string),
			// need to trim leading/tailing spaces, incase the config has HEREDOC
			// formatting and contains a trailing new line
			Statement: strings.TrimSpace(cf["statement"].(string)),
			Priority:  cf["priority"].(int),
		}

		log.Printf("[DEBUG] Create Conditions Opts: %#v", opts)
		_, err := conn.CreateCondition(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *ConditionServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Conditions for (%s)", d.Id())
	conditionList, err := conn.ListConditions(&gofastly.ListConditionsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Conditions for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	cl := flattenConditions(conditionList)

	if err := d.Set(h.GetKey(), cl); err != nil {
		log.Printf("[WARN] Error setting Conditions for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *ConditionServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"statement": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The statement used to determine if the condition is met",
				},
				"priority": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     10,
					Description: "A number used to determine the order in which multiple conditions execute. Lower numbers execute first",
				},
				"type": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "Type of the condition, either `REQUEST`, `RESPONSE`, or `CACHE`",
					ValidateFunc: validateConditionType(),
				},
			},
		},
	}
	return nil
}

func flattenConditions(conditionList []*gofastly.Condition) []map[string]interface{} {
	var cl []map[string]interface{}
	for _, c := range conditionList {
		// Convert Conditions to a map for saving to state.
		nc := map[string]interface{}{
			"name":      c.Name,
			"statement": c.Statement,
			"type":      c.Type,
			"priority":  c.Priority,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nc {
			if v == "" {
				delete(nc, k)
			}
		}

		cl = append(cl, nc)
	}

	return cl
}
