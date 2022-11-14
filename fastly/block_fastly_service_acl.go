package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
func (h *ACLServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, latestVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateACLInput{
		ServiceID:      d.Id(),
		ServiceVersion: latestVersion,
		Name:           gofastly.String(resource["name"].(string)),
	}

	log.Printf("[DEBUG] Fastly ACL creation opts: %#v", opts)
	_, err := conn.CreateACL(&opts)
	if err != nil {
		return err
	}

	return nil
}

// Read refreshes the resource.
func (h *ACLServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, latestVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.Key()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		log.Printf("[DEBUG] Refreshing ACLs for (%s)", d.Id())
		aclList, err := conn.ListACLs(&gofastly.ListACLsInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up ACLs for (%s), version (%v): %s", d.Id(), latestVersion, err)
		}

		al := flattenACLs(aclList)

		// Match up force_destroy on each ACL from schema.ResourceData to avoid d.Set overwriting it with null
		stateACLs := d.Get(h.Key()).(*schema.Set).List()
		for _, acl := range al {
			for _, sa := range stateACLs {
				stateACL := sa.(map[string]any)
				if acl["name"] == stateACL["name"] {
					acl["force_destroy"] = stateACL["force_destroy"]
					break
				}
			}
		}

		if err := d.Set(h.Key(), al); err != nil {
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
func (h *ACLServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, latestVersion int, conn *gofastly.Client) error {
	if !resource["force_destroy"].(bool) {
		mayDelete, err := isACLEmpty(d.Id(), resource["acl_id"].(string), conn)
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
	err := conn.DeleteACL(&opts)

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
func flattenACLs(aclList []*gofastly.ACL) []map[string]any {
	var result []map[string]any
	for _, resource := range aclList {
		data := map[string]any{
			"acl_id": resource.ID,
			"name":   resource.Name,
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

func isACLEmpty(serviceID, aclID string, conn *gofastly.Client) (bool, error) {
	entries, err := conn.ListACLEntries(&gofastly.ListACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}
