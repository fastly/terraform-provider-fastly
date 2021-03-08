package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type HeaderServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceHeader(sa ServiceMetadata) ServiceAttributeDefinition {
	return &HeaderServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "header",
			serviceMetadata: sa,
		},
	}
}

func (h *HeaderServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	oh, nh := d.GetChange(h.GetKey())
	if oh == nil {
		oh = new(schema.Set)
	}
	if nh == nil {
		nh = new(schema.Set)
	}

	oldSet := oh.(*schema.Set)
	newSet := nh.(*schema.Set)

	setDiff := NewSetDiff(func(resource interface{}) (interface{}, error) {
		t, ok := resource.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource failed to be type asserted: %+v", resource)
		}
		return t["name"], nil
	})

	diffResult, err := setDiff.Diff(oldSet, newSet)
	if err != nil {
		return err
	}

	// DELETE removed resources
	for _, resource := range diffResult.Deleted {
		resource := resource.(map[string]interface{})
		opts := gofastly.DeleteHeaderInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
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
	}

	// CREATE new resources
	for _, resource := range diffResult.Added {
		opts, err := buildHeader(resource.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building Header: %s", err)
			return err
		}
		opts.ServiceID = d.Id()
		opts.ServiceVersion = latestVersion

		log.Printf("[DEBUG] Fastly Header Addition opts: %#v", opts)
		_, err = conn.CreateHeader(opts)
		if err != nil {
			return err
		}
	}

	// UPDATE modified resources
	//
	// NOTE: although the go-fastly API client enables updating of a resource by
	// its 'name' attribute, this isn't possible within terraform due to
	// constraints in the data model/schema of the resources not having a uid.
	for _, resource := range diffResult.Modified {
		resource := resource.(map[string]interface{})

		opts := gofastly.UpdateHeaderInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
		}

		// only attempt to update attributes that have changed
		modified := setDiff.Filter(resource, oldSet)

		// NOTE: where we transition between interface{} we lose the ability to
		// infer the underlying type being either a uint vs an int. This
		// materializes as a panic (yay) and so it's only at runtime we discover
		// this and so we've updated the below code to convert the type asserted
		// int into a uint before passing the value to gofastly.Uint().
		if v, ok := modified["action"]; ok {
			opts.Action = gofastly.PHeaderAction(gofastly.HeaderAction(v.(string)))
		}
		if v, ok := modified["ignore_if_set"]; ok {
			opts.IgnoreIfSet = gofastly.CBool(v.(bool))
		}
		if v, ok := modified["type"]; ok {
			opts.Type = gofastly.PHeaderType(gofastly.HeaderType(v.(string)))
		}
		if v, ok := modified["dst"]; ok {
			opts.Destination = gofastly.String(v.(string))
		}
		if v, ok := modified["src"]; ok {
			opts.Source = gofastly.String(v.(string))
		}
		if v, ok := modified["regex"]; ok {
			opts.Regex = gofastly.String(v.(string))
		}
		if v, ok := modified["substitution"]; ok {
			opts.Substitution = gofastly.String(v.(string))
		}
		if v, ok := modified["priority"]; ok {
			opts.Priority = gofastly.Uint(uint(v.(int)))
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
	}

	return nil
}

func (h *HeaderServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Headers for (%s)", d.Id())
	headerList, err := conn.ListHeaders(&gofastly.ListHeadersInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Headers for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	hl := flattenHeaders(headerList)

	if err := d.Set(h.GetKey(), hl); err != nil {
		log.Printf("[WARN] Error setting Headers for (%s): %s", d.Id(), err)
	}

	return nil
}

func (h *HeaderServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unique name for this header attribute. It is important to note that changing this attribute will delete and recreate the resource",
				},
				"action": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "The Header manipulation action to take; must be one of `set`, `append`, `delete`, `regex`, or `regex_repeat`",
					ValidateFunc: validateHeaderAction(),
				},
				"type": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "The Request type on which to apply the selected Action; must be one of `request`, `fetch`, `cache` or `response`",
					ValidateFunc: validateHeaderType(),
				},
				"destination": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the header that is going to be affected by the Action",
				},
				// Optional fields, defaults where they exist
				"ignore_if_set": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Don't add the header if it is already. (Only applies to `set` action.). Default `false`",
				},
				"source": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Variable to be used as a source for the header content (Does not apply to `delete` action.)",
				},
				"regex": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Regular expression to use (Only applies to `regex` and `regex_repeat` actions.)",
				},
				"substitution": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Value to substitute in place of regular expression. (Only applies to `regex` and `regex_repeat`.)",
				},
				"priority": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     100,
					Description: "Lower priorities execute first. Default: `100`",
				},
				"request_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to apply. This `condition` must be of type `REQUEST`",
				},
				"cache_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to apply. This `condition` must be of type `CACHE`",
				},
				"response_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals](https://docs.fastly.com/en/guides/using-conditions)",
				},
			},
		},
	}
	return nil
}

func flattenHeaders(headerList []*gofastly.Header) []map[string]interface{} {
	var hl []map[string]interface{}
	for _, h := range headerList {
		// Convert Header to a map for saving to state.
		nh := map[string]interface{}{
			"name":               h.Name,
			"action":             h.Action,
			"ignore_if_set":      h.IgnoreIfSet,
			"type":               h.Type,
			"destination":        h.Destination,
			"source":             h.Source,
			"regex":              h.Regex,
			"substitution":       h.Substitution,
			"priority":           int(h.Priority),
			"request_condition":  h.RequestCondition,
			"cache_condition":    h.CacheCondition,
			"response_condition": h.ResponseCondition,
		}

		for k, v := range nh {
			if v == "" {
				delete(nh, k)
			}
		}

		hl = append(hl, nh)
	}
	return hl
}

func buildHeader(headerMap interface{}) (*gofastly.CreateHeaderInput, error) {
	df := headerMap.(map[string]interface{})
	opts := gofastly.CreateHeaderInput{
		Name:              df["name"].(string),
		IgnoreIfSet:       gofastly.Compatibool(df["ignore_if_set"].(bool)),
		Destination:       df["destination"].(string),
		Priority:          uint(df["priority"].(int)),
		Source:            df["source"].(string),
		Regex:             df["regex"].(string),
		Substitution:      df["substitution"].(string),
		RequestCondition:  df["request_condition"].(string),
		CacheCondition:    df["cache_condition"].(string),
		ResponseCondition: df["response_condition"].(string),
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "set":
		opts.Action = gofastly.HeaderActionSet
	case "append":
		opts.Action = gofastly.HeaderActionAppend
	case "delete":
		opts.Action = gofastly.HeaderActionDelete
	case "regex":
		opts.Action = gofastly.HeaderActionRegex
	case "regex_repeat":
		opts.Action = gofastly.HeaderActionRegexRepeat
	}

	ty := strings.ToLower(df["type"].(string))
	switch ty {
	case "request":
		opts.Type = gofastly.HeaderTypeRequest
	case "fetch":
		opts.Type = gofastly.HeaderTypeFetch
	case "cache":
		opts.Type = gofastly.HeaderTypeCache
	case "response":
		opts.Type = gofastly.HeaderTypeResponse
	}

	return &opts, nil
}
