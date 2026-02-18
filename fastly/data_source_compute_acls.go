package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v13/fastly/computeacls"
	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyComputeACLs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyComputeACLsRead,
		Schema: map[string]*schema.Schema{
			"acls": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all Compute ACLs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identifier of the Compute ACL.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the Compute ACL.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyComputeACLsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading Compute ACLs")

	acls, err := computeacls.ListACLs(ctx, conn)
	if err != nil {
		return diag.Errorf("error fetching Compute ACLs: %s", err)
	}

	hashBase, _ := json.Marshal(acls)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("acls", flattenDataSourceComputeACLs(acls)); err != nil {
		return diag.Errorf("error setting acls: %s", err)
	}

	return nil
}

func flattenDataSourceComputeACLs(acls *computeacls.ComputeACLs) []map[string]any {
	result := make([]map[string]any, len(acls.Data))
	for i, acl := range acls.Data {
		result[i] = map[string]any{
			"id":   acl.ComputeACLID,
			"name": acl.Name,
		}
	}
	return result
}
