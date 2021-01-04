package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

	ohs := oh.(*schema.Set)
	nhs := nh.(*schema.Set)

	remove := ohs.Difference(nhs).List()
	add := nhs.Difference(ohs).List()

	// Delete removed headers
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteHeaderInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           df["name"].(string),
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

	// POST new Headers
	for _, dRaw := range add {
		opts, err := buildHeader(dRaw.(map[string]interface{}))
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
					Description: "A name to refer to this Header object",
				},
				"action": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "One of set, append, delete, regex, or regex_repeat",
					ValidateFunc: validateHeaderAction(),
				},
				"type": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "Type to manipulate: request, fetch, cache, response",
					ValidateFunc: validateHeaderType(),
				},
				"destination": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Header this affects",
				},
				// Optional fields, defaults where they exist
				"ignore_if_set": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Don't add the header if it is already. (Only applies to 'set' action.). Default `false`",
				},
				"source": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Variable to be used as a source for the header content (Does not apply to 'delete' action.)",
				},
				"regex": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Regular expression to use (Only applies to 'regex' and 'regex_repeat' actions.)",
				},
				"substitution": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Value to substitute in place of regular expression. (Only applies to 'regex' and 'regex_repeat'.)",
				},
				"priority": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     100,
					Description: "Lower priorities execute first. (Default: 100.)",
				},
				"request_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Optional name of a request condition to apply.",
				},
				"cache_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Optional name of a cache condition to apply.",
				},
				"response_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Optional name of a response condition to apply.",
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
