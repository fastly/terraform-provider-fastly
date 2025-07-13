package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

// ACLServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type ACLServiceAttributeHandler struct {
	key string
}

// NewServiceACL returns a new resource.
func NewServiceACL() ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&ACLServiceAttributeHandler{
		key: "acl",
	})
}

// Key returns the resource key.
func (h *ACLServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *ACLServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"acl_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the ACL",
				},
				"force_destroy": {
					Type:        schema.TypeBool,
					Default:     false,
					Optional:    true,
					Description: "Allow the ACL to be deleted, even if it contains entries. Defaults to false.",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A unique name to identify this ACL. It is important to note that changing this attribute will delete and recreate the ACL, and discard the current items in the ACL",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *ACLServiceAttributeHandler) Create(ctx context.Context, d *schema.ResourceData, resource map[string]any, latestVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateACLInput{
		ServiceID:      d.Id(),
		ServiceVersion: latestVersion,
		Name:           gofastly.ToPointer(resource["name"].(string)),
	}

	log.Printf("[DEBUG] Fastly ACL creation opts: %#v", opts)
	_, err := conn.CreateACL(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)
	if err != nil {
		return err
	}

	return nil
}

// Read refreshes the resource.
func (h *ACLServiceAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, _ map[string]any, latestVersion int, conn *gofastly.Client) error {
	v, ok := d.Get(h.Key()).(*schema.Set)
	if !ok {
		return fmt.Errorf("failed to convert ACL state structure to *schema.Set")
	}
	localState := v.List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing ACLs for (%s)", d.Id())
		remoteState, err := conn.ListACLs(gofastly.NewContextForResourceID(ctx, d.Id()), &gofastly.ListACLsInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up ACLs for (%s), version (%v): %s", d.Id(), latestVersion, err)
		}

		ms := flattenACLs(remoteState)

		// Match up force_destroy on each ACL from schema.ResourceData to avoid d.Set overwriting it with null
		for _, m := range ms {
			for _, i := range localState {
				lm, ok := i.(map[string]any)
				if !ok {
					return fmt.Errorf("failed to convert ACL state structure to map[string]any")
				}
				if m["name"] == lm["name"] {
					m["force_destroy"] = lm["force_destroy"]
					break
				}
			}
		}

		if err := d.Set(h.Key(), ms); err != nil {
			log.Printf("[WARN] Error setting ACLs for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *ACLServiceAttributeHandler) Update(context.Context, *schema.ResourceData, map[string]any, map[string]any, int, *gofastly.Client) error {
	return nil
}

// Delete deletes the resource.
func (h *ACLServiceAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, resource map[string]any, latestVersion int, conn *gofastly.Client) error {
	if !resource["force_destroy"].(bool) {
		mayDelete, err := isACLEmpty(ctx, d.Id(), resource["acl_id"].(string), conn)
		if err != nil {
			return err
		}

		if !mayDelete {
			return fmt.Errorf("cannot delete ACL (%s), list is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", resource["acl_id"].(string))
		}
	}

	opts := gofastly.DeleteACLInput{
		ServiceID:      d.Id(),
		ServiceVersion: latestVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly ACL removal opts: %#v", opts)
	err := conn.DeleteACL(gofastly.NewContextForResourceID(ctx, d.Id()), &opts)

	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// flattenACLs models data into format suitable for saving to Terraform state.
func flattenACLs(remoteState []*gofastly.ACL) []map[string]any {
	var result []map[string]any
	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.ACLID != nil {
			data["acl_id"] = *resource.ACLID
		}
		if resource.Name != nil {
			data["name"] = *resource.Name
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

func isACLEmpty(ctx context.Context, serviceID, aclID string, conn *gofastly.Client) (bool, error) {
	entries, err := conn.ListACLEntries(gofastly.NewContextForResourceID(ctx, serviceID), &gofastly.ListACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}
