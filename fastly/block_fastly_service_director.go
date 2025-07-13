package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

// DirectorServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type DirectorServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceDirector constructs a service attribute.
func NewServiceDirector(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&DirectorServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "director",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *DirectorServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *DirectorServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"backends": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "Names of defined backends to map the director to. Example: `[ \"origin1\", \"origin2\" ]`",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"comment": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "An optional comment about the Director",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unique name for this Director. It is important to note that changing this attribute will delete and recreate the resource",
				},
				"quorum": {
					Type:             schema.TypeInt,
					Optional:         true,
					Default:          75,
					Description:      "Percentage of capacity that needs to be up for the director itself to be considered up. Default `75`",
					ValidateDiagFunc: validateDirectorQuorum(),
				},
				"retries": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     5,
					Description: "How many backends to search if it fails. Default `5`",
				},
				"shield": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
					Description: "Selected POP to serve as a \"shield\" for backends. Valid values for `shield` are included in the [`GET /datacenters`](https://developer.fastly.com/reference/api/utils/datacenter/) API response",
				},
				"type": {
					Type:             schema.TypeInt,
					Optional:         true,
					Default:          1,
					Description:      "Type of load balance group to use. Integer, 1 to 4. Values: `1` (random), `3` (hash), `4` (client). Default `1`",
					ValidateDiagFunc: validateDirectorType(),
				},
			},
		},
	}
}

// Create creates a new resource instance.
func (h *DirectorServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateDirectorInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
		Comment:        gofastly.ToPointer(resource["comment"].(string)),
		Shield:         gofastly.ToPointer(resource["shield"].(string)),
		Quorum:         gofastly.ToPointer(resource["quorum"].(int)),
		Retries:        gofastly.ToPointer(resource["retries"].(int)),
	}

	switch resource["type"].(int) {
	case 1:
		opts.Type = gofastly.ToPointer(gofastly.DirectorTypeRandom)
	case 2:
		opts.Type = gofastly.ToPointer(gofastly.DirectorTypeRoundRobin)
	case 3:
		opts.Type = gofastly.ToPointer(gofastly.DirectorTypeHash)
	case 4:
		opts.Type = gofastly.ToPointer(gofastly.DirectorTypeClient)
	}

	log.Printf("[DEBUG] Director Create opts: %#v", opts)
	_, err := conn.CreateDirector(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}

	if v, ok := resource["backends"]; ok {
		backends := v.(*schema.Set).List()
		if len(backends) > 0 {
			for _, backend := range backends {
				opts := gofastly.CreateDirectorBackendInput{
					ServiceID:      d.Id(),
					ServiceVersion: serviceVersion,
					Director:       resource["name"].(string),
					Backend:        backend.(string),
				}

				log.Printf("[DEBUG] Director Backend Create opts: %#v", opts)
				_, err := conn.CreateDirectorBackend(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Read refreshes the resource state.
func (h *DirectorServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.GetKey()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Directors for (%s)", d.Id())
		remoteState, err := conn.ListDirectors(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListDirectorsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Directors for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		dirl := flattenDirectors(remoteState)

		if err := d.Set(h.GetKey(), dirl); err != nil {
			log.Printf("[WARN] Error setting Directors for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource instance.
func (h *DirectorServiceAttributeHandler) Update(ctx context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateDirectorInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	// NOTE: When converting from an interface{} we lose the underlying type.
	// Converting to the wrong type will result in a runtime panic.
	if v, ok := modified["comment"]; ok {
		opts.Comment = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["shield"]; ok {
		opts.Shield = gofastly.ToPointer(v.(string))
	}
	if v, ok := modified["quorum"]; ok {
		opts.Quorum = gofastly.ToPointer(v.(int))
	}
	if v, ok := modified["type"]; ok {
		switch v.(int) {
		case 1:
			opts.Type = gofastly.ToPointer(gofastly.DirectorTypeRandom)
		case 2:
			opts.Type = gofastly.ToPointer(gofastly.DirectorTypeRoundRobin)
		case 3:
			opts.Type = gofastly.ToPointer(gofastly.DirectorTypeHash)
		case 4:
			opts.Type = gofastly.ToPointer(gofastly.DirectorTypeClient)
		}
	}
	if v, ok := modified["retries"]; ok {
		opts.Retries = gofastly.ToPointer(v.(int))
	}

	log.Printf("[DEBUG] Update Director Opts: %#v", opts)
	_, err := conn.UpdateDirector(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}

	if _, ok := modified["backends"]; ok {
		odb, ndb := getDirectorBackendChange(d, resource)

		remove := odb.Difference(ndb).List()
		for _, b := range remove {
			opts := gofastly.DeleteDirectorBackendInput{
				ServiceID:      d.Id(),
				ServiceVersion: serviceVersion,
				Director:       resource["name"].(string),
				Backend:        b.(string),
			}
			log.Printf("[DEBUG] Director Backend Update opts: %#v", opts)
			err := conn.DeleteDirectorBackend(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
			if err != nil {
				// If we end up trying to remove a backend that no longer exists, then the
				// API will return a '404 Not Found'. We don't want to return those errors
				// as they ultimately don't mean anything useful to the user.
				if !strings.Contains(err.Error(), "404 - Not Found") {
					return err
				}
			}
		}

		add := ndb.Difference(odb).List()
		for _, b := range add {
			opts := gofastly.CreateDirectorBackendInput{
				ServiceID:      d.Id(),
				ServiceVersion: serviceVersion,
				Director:       resource["name"].(string),
				Backend:        b.(string),
			}
			log.Printf("[DEBUG] Director Backend Update opts: %#v", opts)
			_, err := conn.CreateDirectorBackend(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete deletes the resource instance.
func (h *DirectorServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteDirectorInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Director Removal opts: %#v", opts)
	err := conn.DeleteDirector(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// flattenDirectors models data into format suitable for saving to Terraform state.
func flattenDirectors(remoteState []*gofastly.Director) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Comment != nil {
			data["comment"] = *resource.Comment
		}
		if resource.Shield != nil {
			data["shield"] = *resource.Shield
		}
		if resource.Type != nil {
			data["type"] = *resource.Type
		}
		if resource.Quorum != nil {
			data["quorum"] = *resource.Quorum
		}
		if resource.Retries != nil {
			data["retries"] = *resource.Retries
		}

		// NOTE: schema.NewSet expects slice of empty interface so we have to build
		// this from the Dictionary's Backend field.
		var b []any
		for _, v := range resource.Backends {
			b = append(b, v)
		}
		if len(b) > 0 {
			data["backends"] = schema.NewSet(schema.HashString, b)
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

func getDirectorBackendChange(d *schema.ResourceData, resource map[string]any) (odb *schema.Set, ndb *schema.Set) {
	od, nd := d.GetChange("director")

	if od == nil {
		od = new(schema.Set)
	}
	if nd == nil {
		nd = new(schema.Set)
	}

	get := func(targetDirectorName string, directorsSet *schema.Set) *schema.Set {
		for _, director := range directorsSet.List() {
			director := director.(map[string]any)

			if director["name"] == targetDirectorName {
				return director["backends"].(*schema.Set)
			}
		}
		return new(schema.Set)
	}

	name := resource["name"]
	odb = get(name.(string), od.(*schema.Set))
	ndb = get(name.(string), nd.(*schema.Set))

	return odb, ndb
}
