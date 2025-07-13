package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
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
func (h *GzipServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateGzipInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		CacheCondition: gofastly.ToPointer(resource["cache_condition"].(string)),
	}

	if v, ok := resource["content_types"]; ok {
		opts.ContentTypes = gofastly.ToPointer(sliceToString(v.([]any)))
	}

	if v, ok := resource["extensions"]; ok {
		opts.Extensions = gofastly.ToPointer(sliceToString(v.([]any)))
	}

	log.Printf("[DEBUG] Fastly Gzip Addition opts: %#v", opts)
	_, err := conn.CreateGzip(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *GzipServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Gzips for (%s)", d.Id())
		remoteState, err := conn.ListGzips(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListGzipsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Gzips for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		gl := flattenGzips(remoteState)

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

		if err := d.Set(h.GetKey(), gl); err != nil {
			log.Printf("[WARN] Error setting Gzips for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *GzipServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateGzipInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["content_types"]; ok {
		// NOTE: this particular line was added to address a change in the backend API
		// where it used to accept an empty value but now will use a default value if no value provided.
		// To allow "resetting" the value on modify (user removed the attribute or set empty value)
		// we always default to sending an empty string
		opts.ContentTypes = gofastly.ToPointer("")

		list := v.([]any)
		if len(list) > 0 {
			opts.ContentTypes = gofastly.ToPointer(sliceToString(list))
		}
	}
	if v, ok := modified["extensions"]; ok {
		opts.Extensions = gofastly.ToPointer("")
		list := v.([]any)
		if len(list) > 0 {
			opts.Extensions = gofastly.ToPointer(sliceToString(list))
		}
	}
	if v, ok := modified["cache_condition"]; ok {
		opts.CacheCondition = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Update Gzip Opts: %#v", opts)
	_, err := conn.UpdateGzip(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *GzipServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteGzipInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Gzip removal opts: %#v", opts)
	err := conn.DeleteGzip(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenGzips models data into format suitable for saving to Terraform state.
func flattenGzips(remoteState []*gofastly.Gzip) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.CacheCondition != nil {
			data["cache_condition"] = *resource.CacheCondition
		}
		if resource.Extensions != nil {
			e := strings.Split(*resource.Extensions, " ")
			var et []any
			for _, ev := range e {
				et = append(et, ev)
			}
			data["extensions"] = et
		}
		if resource.ContentTypes != nil {
			c := strings.Split(*resource.ContentTypes, " ")
			var ct []any
			for _, cv := range c {
				ct = append(ct, cv)
			}
			data["content_types"] = ct
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

func sliceToString(src []any) string {
	var result []string
	for _, el := range src {
		result = append(result, el.(string))
	}
	return strings.Join(result, " ")
}
