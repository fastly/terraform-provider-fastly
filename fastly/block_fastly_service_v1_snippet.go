package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type SnippetServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceSnippet(sa ServiceMetadata) ServiceAttributeDefinition {
	return &SnippetServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "snippet",
			serviceMetadata: sa,
		},
	}
}

func (h *SnippetServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	// Note: as above with Gzip and S3 logging, we don't utilize the PUT
	// endpoint to update a VCL snippet, we simply destroy it and create a new one.
	oldSnippetVal, newSnippetVal := d.GetChange(h.GetKey())
	if oldSnippetVal == nil {
		oldSnippetVal = new(schema.Set)
	}
	if newSnippetVal == nil {
		newSnippetVal = new(schema.Set)
	}

	oldSnippetSet := oldSnippetVal.(*schema.Set)
	newSnippetSet := newSnippetVal.(*schema.Set)

	remove := oldSnippetSet.Difference(newSnippetSet).List()
	add := newSnippetSet.Difference(oldSnippetSet).List()

	// Delete removed VCL Snippet configurations
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteSnippetInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly VCL Snippet Removal opts: %#v", opts)
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
		opts, err := buildSnippet(dRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building VCL Snippet: %s", err)
			return err
		}
		opts.Service = d.Id()
		opts.Version = latestVersion

		log.Printf("[DEBUG] Fastly VCL Snippet Addition opts: %#v", opts)
		_, err = conn.CreateSnippet(opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *SnippetServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing VCL Snippets for (%s)", d.Id())
	snippetList, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up VCL Snippets for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	vsl := flattenSnippets(snippetList)

	if err := d.Set(h.GetKey(), vsl); err != nil {
		log.Printf("[WARN] Error setting VCL Snippets for (%s): %s", d.Id(), err)
	}
	return nil
}

func (h *SnippetServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
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
				"content": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The contents of the VCL snippet",
				},
				"priority": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     100,
					Description: "Determines ordering for multiple snippets. Lower priorities execute first. (Default: 100)",
				},
			},
		},
	}
	return nil
}

func buildSnippet(snippetMap interface{}) (*gofastly.CreateSnippetInput, error) {
	df := snippetMap.(map[string]interface{})
	opts := gofastly.CreateSnippetInput{
		Name:     df["name"].(string),
		Content:  df["content"].(string),
		Priority: df["priority"].(int),
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

func flattenSnippets(snippetList []*gofastly.Snippet) []map[string]interface{} {
	var sl []map[string]interface{}
	for _, snippet := range snippetList {
		// Skip dynamic snippets
		if snippet.Dynamic == 1 {
			continue
		}

		// Convert VCLs to a map for saving to state.
		snippetMap := map[string]interface{}{
			"name":     snippet.Name,
			"type":     snippet.Type,
			"priority": int(snippet.Priority),
			"content":  snippet.Content,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range snippetMap {
			if v == "" {
				delete(snippetMap, k)
			}
		}

		sl = append(sl, snippetMap)
	}

	return sl
}
