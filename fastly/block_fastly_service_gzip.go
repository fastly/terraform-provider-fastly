package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GzipServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type GzipServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceGzip returns a new resource.
func NewServiceGzip(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&GzipServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "gzip",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *GzipServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *GzipServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cache_condition": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Name of already defined `condition` controlling when this gzip configuration applies. This `condition` must be of type `CACHE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals](https://docs.fastly.com/en/guides/using-conditions)",
				},
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
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A name to refer to this gzip condition. It is important to note that changing this attribute will delete and recreate the resource",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *GzipServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateGzipInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
		CacheCondition: resource["cache_condition"].(string),
	}

	if v, ok := resource["content_types"]; ok {
		opts.ContentTypes = sliceToString(v.([]any))
	}

	if v, ok := resource["extensions"]; ok {
		opts.Extensions = sliceToString(v.([]any))
	}

	log.Printf("[DEBUG] Fastly Gzip Addition opts: %#v", opts)
	_, err := conn.CreateGzip(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *GzipServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing Gzips for (%s)", d.Id())
		gzipsList, err := conn.ListGzips(&gofastly.ListGzipsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Gzips for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		gl := flattenGzips(gzipsList)

		// NOTE: Although "content_types" and "extensions" fields are optional in spec,
		// Fastly API will actually set the default value silently when these fields are not sent
		// or an empty field value is sent. This will cause unexpected diff.
		// We need to ignore these fields in the API response unless field values are explicitly set.
		{
			type IgnoreFields struct {
				Name string
			}
			ignoreList := map[string][]IgnoreFields{}

			for _, elem := range d.Get("gzip").(*schema.Set).List() {
				m := elem.(map[string]any)
				name := m["name"].(string)
				if len(m["content_types"].([]any)) == 0 {
					ignoreList[name] = append(ignoreList[name], IgnoreFields{Name: "content_types"})
				}
				if len(m["extensions"].([]any)) == 0 {
					ignoreList[name] = append(ignoreList[name], IgnoreFields{Name: "extensions"})
				}
			}

			for i, g := range gl {
				if v, ok := ignoreList[g["name"].(string)]; ok && len(v) > 0 {
					for _, sl := range v {
						gl[i][sl.Name] = nil
					}
				}
			}
		}

		// lintignore:R001
		if err := d.Set(h.GetKey(), gl); err != nil {
			log.Printf("[WARN] Error setting Gzips for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *GzipServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateGzipInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: where we transition between any we lose the ability to
	// infer the underlying type being either a uint vs an int. This
	// materializes as a panic (yay) and so it's only at runtime we discover
	// this and so we've updated the below code to convert the type asserted
	// int into a uint before passing the value to gofastly.Uint().
	if v, ok := modified["content_types"]; ok {
		// NOTE: this particular line was added to address a change in the backend API
		// where it used to accept an empty value but now will use a default value if no value provided.
		// To allow "resetting" the value on modify (user removed the attribute or set empty value)
		// we always default to sending an empty string
		opts.ContentTypes = gofastly.String("")

		list := v.([]any)
		if len(list) > 0 {
			opts.ContentTypes = gofastly.String(sliceToString(list))
		}
	}
	if v, ok := modified["extensions"]; ok {
		opts.Extensions = gofastly.String("")
		list := v.([]any)
		if len(list) > 0 {
			opts.Extensions = gofastly.String(sliceToString(list))
		}
	}
	if v, ok := modified["cache_condition"]; ok {
		opts.CacheCondition = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Gzip Opts: %#v", opts)
	_, err := conn.UpdateGzip(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *GzipServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteGzipInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
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
	return nil
}

func flattenGzips(gzipsList []*gofastly.Gzip) []map[string]any {
	var gl []map[string]any
	for _, g := range gzipsList {
		// Convert Gzip to a map for saving to state.
		ng := map[string]any{
			"name":            g.Name,
			"cache_condition": g.CacheCondition,
		}

		if g.Extensions != "" {
			e := strings.Split(g.Extensions, " ")
			var et []any
			for _, ev := range e {
				et = append(et, ev)
			}
			ng["extensions"] = et
		}

		if g.ContentTypes != "" {
			c := strings.Split(g.ContentTypes, " ")
			var ct []any
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

func sliceToString(src []any) string {
	var result []string
	for _, el := range src {
		result = append(result, el.(string))
	}
	return strings.Join(result, " ")
}
