package fastly

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"strconv"

	"github.com/fastly/go-fastly/v10/fastly/computeacls"
	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyComputeACLs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyComputeACLsRead,
		Schema: map[string]*schema.Schema{
			"acls": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of ACLs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"acl_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying the ACL.",
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name for the ACL.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyComputeACLsRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading ACLs")

	acls, err := computeacls.ListACLs(conn)
	if err != nil {
		return diag.Errorf("error fetching ACLs: %s", err)
	}

	var filteredACLs []computeacls.ComputeACL = []computeacls.ComputeACL{}
	if input := d.Get("acls").(*schema.Set); input.Len() != 0 {
		name := input.List()[0].(map[string]any)["name"].(string)
		regex := regexp.MustCompile(name)
		for _, acl := range acls.Data {
			if regex.MatchString(acl.Name) {
				filteredACLs = append(filteredACLs, acl)
			}
		}
	} else {
		filteredACLs = acls.Data
	}

	aclsToSet := &computeacls.ComputeACLs{
		Data: filteredACLs,
		Meta: computeacls.MetaACLs{
			Total: len(filteredACLs),
		},
	}
	if err := d.Set("acls", flattenDataSourceACLs(aclsToSet)); err != nil {
		return diag.Errorf("error setting ACLs: %s", err)
	}

	hashBase, _ := json.Marshal(aclsToSet)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	return nil
}

// flattenDataSourceACLs models data into format suitable for saving to
// Terraform state.
func flattenDataSourceACLs(acls *computeacls.ComputeACLs) []map[string]any {
	result := make([]map[string]any, len(acls.Data))
	if acls.Meta.Total == 0 {
		return result
	}

	for i, resource := range acls.Data {
		result[i] = map[string]any{
			"acl_id": resource.ComputeACLID,
			"name":   resource.Name,
		}
	}

	return result
}
