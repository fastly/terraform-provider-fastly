package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type GzipServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceGzip(sa ServiceMetadata) ServiceAttributeDefinition {
	return &GzipServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "gzip",
			serviceMetadata: sa,
		},
	}
}

func (h *GzipServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	og, ng := d.GetChange(h.GetKey())
	if og == nil {
		og = new(schema.Set)
	}
	if ng == nil {
		ng = new(schema.Set)
	}

	oldSet := og.(*schema.Set)
	newSet := ng.(*schema.Set)

	setDiff := NewSetDiff(func(resource interface{}) (interface{}, error) {
		// Use the resource name as the key
		return resource.(map[string]interface{})["name"], nil
	})

	diffResult, err := setDiff.Diff(oldSet, newSet)
	if err != nil {
		return err
	}

	// Delete removed gzip rules
	for _, dRaw := range diffResult.Deleted {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteGzipInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Gzip removal opts: %#v", opts)
		err := conn.DeleteGzip(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new Gzips
	for _, dRaw := range diffResult.Added {
		df := dRaw.(map[string]interface{})
		opts := gofastly.CreateGzipInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           df["name"].(string),
			CacheCondition: df["cache_condition"].(string),
		}

		if v, ok := df["content_types"]; ok {
			if len(v.(*schema.Set).List()) > 0 {
				var cl []string
				for _, c := range v.(*schema.Set).List() {
					cl = append(cl, c.(string))
				}
				opts.ContentTypes = strings.Join(cl, " ")
			}
		}

		if v, ok := df["extensions"]; ok {
			if len(v.(*schema.Set).List()) > 0 {
				var el []string
				for _, e := range v.(*schema.Set).List() {
					el = append(el, e.(string))
				}
				opts.Extensions = strings.Join(el, " ")
			}
		}

		log.Printf("[DEBUG] Fastly Gzip Addition opts: %#v", opts)
		_, err := conn.CreateGzip(&opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *GzipServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Gzips for (%s)", d.Id())
	gzipsList, err := conn.ListGzips(&gofastly.ListGzipsInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Gzips for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	gl := flattenGzips(gzipsList)

	if err := d.Set(h.GetKey(), gl); err != nil {
		log.Printf("[WARN] Error setting Gzips for (%s): %s", d.Id(), err)
	}

	return nil
}

func (h *GzipServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A name to refer to this gzip condition",
				},
				// optional fields
				"content_types": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "The content-type for each type of content you wish to have dynamically gzip'ed. Example: `[\"text/html\", \"text/css\"]`",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"extensions": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "File extensions for each file type to dynamically gzip. Example: `[\"css\", \"js\"]`",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"cache_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` controlling when this gzip configuration applies. This `condition` must be of type `CACHE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals](https://docs.fastly.com/en/guides/using-conditions)",
				},
			},
		},
	}
	return nil
}

func flattenGzips(gzipsList []*gofastly.Gzip) []map[string]interface{} {
	var gl []map[string]interface{}
	for _, g := range gzipsList {
		// Convert Gzip to a map for saving to state.
		ng := map[string]interface{}{
			"name":            g.Name,
			"cache_condition": g.CacheCondition,
		}

		if g.Extensions != "" {
			e := strings.Split(g.Extensions, " ")
			var et []interface{}
			for _, ev := range e {
				et = append(et, ev)
			}
			ng["extensions"] = schema.NewSet(schema.HashString, et)
		}

		if g.ContentTypes != "" {
			c := strings.Split(g.ContentTypes, " ")
			var ct []interface{}
			for _, cv := range c {
				ct = append(ct, cv)
			}
			ng["content_types"] = schema.NewSet(schema.HashString, ct)
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range ng {
			if v == "" {
				delete(ng, k)
			}
		}

		gl = append(gl, ng)
	}

	return gl
}
