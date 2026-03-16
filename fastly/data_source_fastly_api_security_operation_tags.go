package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"
	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyAPISecurityOperationTags() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyAPISecurityOperationTagsRead,
		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service ID.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Operation tags.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Created timestamp (when present).",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tag description (when present).",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tag ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tag name.",
						},
						"operation_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of operations associated with this tag (when present).",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Updated timestamp (when present).",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of matching results, as returned by the API.",
			},
		},
	}
}

func dataSourceFastlyAPISecurityOperationTagsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	serviceID := d.Get("service_id").(string)

	in := &operations.ListTagsInput{
		ServiceID: gofastly.ToPointer(serviceID),
	}

	log.Printf("[DEBUG] Reading API Security operation tags: %#v", in)
	out, err := operations.ListTags(ctx, conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create a stable ID based on the returned set.
	hashBase, _ := json.Marshal(out)
	d.SetId(strconv.Itoa(hashcode.String(string(hashBase))))

	if err := d.Set("tags", flattenAPISecurityTags(out)); err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("total", out.Meta.Total)

	return nil
}

func flattenAPISecurityTags(remote *operations.OperationTags) []map[string]any {
	if remote == nil || len(remote.Data) == 0 {
		return []map[string]any{}
	}

	out := make([]map[string]any, 0, len(remote.Data))
	for _, t := range remote.Data {
		out = append(out, map[string]any{
			"created_at":      t.CreatedAt,
			"description":     t.Description,
			"id":              t.ID,
			"name":            t.Name,
			"operation_count": t.Count,
			"updated_at":      t.UpdatedAt,
		})
	}
	return out
}
