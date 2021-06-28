package fastly

import (
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		opts := gofastly.DeleteGzipInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
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

	// CREATE new resources
	for _, resource := range diffResult.Added {
		resource := resource.(map[string]interface{})
		opts := gofastly.CreateGzipInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
			CacheCondition: resource["cache_condition"].(string),
		}

		if v, ok := resource["content_types"]; ok {
			opts.ContentTypes = sliceToString(v.([]interface{}))
		}

		if v, ok := resource["extensions"]; ok {
			opts.Extensions = sliceToString(v.([]interface{}))
		}

		log.Printf("[DEBUG] Fastly Gzip Addition opts: %#v", opts)
		_, err := conn.CreateGzip(&opts)
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

		opts := gofastly.UpdateGzipInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
		}

		// NOTE: []interface{} is not comparable in Filter function
		// covert it into string in advance
		resource["content_types"] = sliceToString(resource["content_types"].([]interface{}))
		resource["extensions"] = sliceToString(resource["extensions"].([]interface{}))

		// only attempt to update attributes that have changed
		modified := setDiff.Filter(resource, oldSet)

		if v, ok := modified["content_types"]; ok {
			opts.ContentTypes = gofastly.String(v.(string))
		}
		if v, ok := modified["extensions"]; ok {
			opts.Extensions = gofastly.String(v.(string))
		}
		if v, ok := modified["cache_condition"]; ok {
			opts.CacheCondition = gofastly.String(v.(string))
		}

		log.Printf("[DEBUG] Update Gzip Opts: %#v", opts)
		_, err := conn.UpdateGzip(&opts)
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
					Description: "A name to refer to this gzip condition. It is important to note that changing this attribute will delete and recreate the resource",
				},
				// optional fields
				"content_types": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The content-type for each type of content you wish to have dynamically gzip'ed. Example: `[\"text/html\", \"text/css\"]`",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"extensions": {
					Type:        schema.TypeList,
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
			ng["extensions"] = et
		}

		if g.ContentTypes != "" {
			c := strings.Split(g.ContentTypes, " ")
			var ct []interface{}
			for _, cv := range c {
				ct = append(ct, cv)
			}
			ng["content_types"] = ct
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

func sliceToString(src []interface{}) string {
	var result []string
	for _, el := range src {
		result = append(result, el.(string))
	}
	return strings.Join(result, " ")
}
