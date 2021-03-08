package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ACLServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceACL(sa ServiceMetadata) ServiceAttributeDefinition {
	return &ACLServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "acl",
			serviceMetadata: sa,
		},
	}
}

func (h *ACLServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	oldACLVal, newACLVal := d.GetChange(h.GetKey())
	if oldACLVal == nil {
		oldACLVal = new(schema.Set)
	}
	if newACLVal == nil {
		newACLVal = new(schema.Set)
	}

	oldSet := oldACLVal.(*schema.Set)
	newSet := newACLVal.(*schema.Set)

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

		if !resource["force_destroy"].(bool) {
			mayDelete, err := isACLEmpty(d.Id(), resource["acl_id"].(string), conn)
			if err != nil {
				return err
			}

			if !mayDelete {
				return fmt.Errorf("Cannot delete ACL (%s), list is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change.", resource["acl_id"].(string))
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
	}

	// CREATE new resources
	for _, resource := range diffResult.Added {
		resource := resource.(map[string]interface{})
		opts := gofastly.CreateACLInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			Name:           resource["name"].(string),
		}

		log.Printf("[DEBUG] Fastly ACL creation opts: %#v", opts)
		_, err := conn.CreateACL(&opts)
		if err != nil {
			return err
		}
	}

	// UPDATE modified resources (NOT IMPLEMENTED)
	//
	// Although the go-fastly API client enables updating of a resource by
	// its 'name' attribute, this isn't possible within terraform due to
	// constraints in the data model/schema of the resources not having a uid.
	//
	// Because of this we do not implement any logic for updating the ACL
	// resource, only CREATE and DELETE functionality.

	return nil
}

func (h *ACLServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {

	log.Printf("[DEBUG] Refreshing ACLs for (%s)", d.Id())
	aclList, err := conn.ListACLs(&gofastly.ListACLsInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up ACLs for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	al := flattenACLs(aclList)

	// Match up force_destroy on each ACL from schema.ResourceData to avoid d.Set overwriting it with null
	stateACLs := d.Get(h.GetKey()).(*schema.Set).List()
	for _, acl := range al {
		for _, sa := range stateACLs {
			stateACL := sa.(map[string]interface{})
			if acl["name"] == stateACL["name"] {
				acl["force_destroy"] = stateACL["force_destroy"]
				break
			}
		}
	}

	if err := d.Set(h.GetKey(), al); err != nil {
		log.Printf("[WARN] Error setting ACLs for (%s): %s", d.Id(), err)
	}

	return nil
}

func (h *ACLServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A unique name to identify this ACL. It is important to note that changing this attribute will delete and recreate the ACL, and discard the current items in the ACL",
				},
				// Optional fields
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
			},
		},
	}
	return nil
}

func flattenACLs(aclList []*gofastly.ACL) []map[string]interface{} {
	var al []map[string]interface{}
	for _, acl := range aclList {
		// Convert ACLs to a map for saving to state.
		aclMap := map[string]interface{}{
			"acl_id": acl.ID,
			"name":   acl.Name,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range aclMap {
			if v == "" {
				delete(aclMap, k)
			}
		}

		al = append(al, aclMap)
	}

	return al
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
