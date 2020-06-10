package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var headerSchema = &schema.Schema{
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


func processHeader(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	oh, nh := d.GetChange("header")
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
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
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
		opts.Service = d.Id()
		opts.Version = latestVersion

		log.Printf("[DEBUG] Fastly Header Addition opts: %#v", opts)
		_, err = conn.CreateHeader(opts)
		if err != nil {
			return err
		}
	}

	return nil
}


func readHeader(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	log.Printf("[DEBUG] Refreshing Headers for (%s)", d.Id())
	headerList, err := conn.ListHeaders(&gofastly.ListHeadersInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Headers for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	hl := flattenHeaders(headerList)

	if err := d.Set("header", hl); err != nil {
		log.Printf("[WARN] Error setting Headers for (%s): %s", d.Id(), err)
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