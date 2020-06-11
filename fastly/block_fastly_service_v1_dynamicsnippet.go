package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type DynamicSnippetServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceDynamicSnippet() ServiceAttributeDefinition {
	return &DynamicSnippetServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			schema: dynamicsnippetSchema,
			key:    "dynamicsnippet",
		},
	}
}

var dynamicsnippetSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name to refer to this VCL snippet",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of init, recv, hit, miss, pass, fetch, error, deliver, log, none",
				ValidateFunc: validateSnippetType(),
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Determines ordering for multiple snippets. Lower priorities execute first. (Default: 100)",
			},
			"snippet_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Generated VCL snippet Id",
			},
		},
	},
}

func (h *DynamicSnippetServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	// Note: as above with Gzip and S3 logging, we don't utilize the PUT
	// endpoint to update a VCL dynamic snippet, we simply destroy it and create a new one.
	oldDynamicSnippetVal, newDynamicSnippetVal := d.GetChange("dynamicsnippet")
	if oldDynamicSnippetVal == nil {
		oldDynamicSnippetVal = new(schema.Set)
	}
	if newDynamicSnippetVal == nil {
		newDynamicSnippetVal = new(schema.Set)
	}

	oldDynamicSnippetSet := oldDynamicSnippetVal.(*schema.Set)
	newDynamicSnippetSet := newDynamicSnippetVal.(*schema.Set)

	remove := oldDynamicSnippetSet.Difference(newDynamicSnippetSet).List()
	add := newDynamicSnippetSet.Difference(oldDynamicSnippetSet).List()

	// Delete removed VCL Snippet configurations
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteSnippetInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly VCL Dynamic Snippet Removal opts: %#v", opts)
		err := conn.DeleteSnippet(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new VCL Snippet configurations
	for _, dRaw := range add {
		opts, err := buildDynamicSnippet(dRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building VCL Dynamic Snippet: %s", err)
			return err
		}
		opts.Service = d.Id()
		opts.Version = latestVersion

		log.Printf("[DEBUG] Fastly VCL Dynamic Snippet Addition opts: %#v", opts)
		_, err = conn.CreateSnippet(opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *DynamicSnippetServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing VCL Snippets for (%s)", d.Id())
	snippetList, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up VCL Snippets for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	dynamicSnippets := flattenDynamicSnippets(snippetList)
	if err := d.Set("dynamicsnippet", dynamicSnippets); err != nil {
		log.Printf("[WARN] Error setting VCL Dynamic Snippets for (%s): %s", d.Id(), err)
	}

	return nil
}

func buildDynamicSnippet(dynamicSnippetMap interface{}) (*gofastly.CreateSnippetInput, error) {
	df := dynamicSnippetMap.(map[string]interface{})
	opts := gofastly.CreateSnippetInput{
		Name:     df["name"].(string),
		Priority: df["priority"].(int),
		Dynamic:  1,
	}

	snippetType := strings.ToLower(df["type"].(string))
	switch snippetType {
	case "init":
		opts.Type = gofastly.SnippetTypeInit
	case "recv":
		opts.Type = gofastly.SnippetTypeRecv
	case "hash":
		opts.Type = gofastly.SnippetTypeHash
	case "hit":
		opts.Type = gofastly.SnippetTypeHit
	case "miss":
		opts.Type = gofastly.SnippetTypeMiss
	case "pass":
		opts.Type = gofastly.SnippetTypePass
	case "fetch":
		opts.Type = gofastly.SnippetTypeFetch
	case "error":
		opts.Type = gofastly.SnippetTypeError
	case "deliver":
		opts.Type = gofastly.SnippetTypeDeliver
	case "log":
		opts.Type = gofastly.SnippetTypeLog
	case "none":
		opts.Type = gofastly.SnippetTypeNone
	}

	return &opts, nil
}

func flattenDynamicSnippets(dynamicSnippetList []*gofastly.Snippet) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, dynamicSnippet := range dynamicSnippetList {
		// Skip non-dynamic snippets
		if dynamicSnippet.Dynamic == 0 {
			continue
		}

		// Convert VCLs to a map for saving to state.
		dynamicSnippetMap := map[string]interface{}{
			"snippet_id": dynamicSnippet.ID,
			"name":       dynamicSnippet.Name,
			"type":       dynamicSnippet.Type,
			"priority":   int(dynamicSnippet.Priority),
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range dynamicSnippetMap {
			if v == "" {
				delete(dynamicSnippetMap, k)
			}
		}

		sl = append(sl, dynamicSnippetMap)
	}

	return sl
}
