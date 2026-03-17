package fastly

import (
	"context"
	"encoding/json"
	"fmt"
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
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: fmt.Sprintf("Page size (maximum number of results per request). Default value `%d`.", apiSecurityDefaultPageLimit),
			},
			"limit_returned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The limit value returned by the API in the response metadata (if present).",
			},
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

	pageLimit := apiSecurityDefaultPageLimit
	if v, ok := d.GetOk("limit"); ok {
		if n := v.(int); n > 0 {
			pageLimit = n
		}
	}

	in := &operations.ListTagsInput{
		ServiceID: gofastly.ToPointer(serviceID),
		Limit:     gofastly.ToPointer(pageLimit),
		Page:      gofastly.ToPointer(0),
	}

	// First request: capture meta.total/meta.limit (for state + docs).
	log.Printf("[DEBUG] Reading API Security operation tags (page 0): %#v", in)
	first, err := operations.ListTags(ctx, conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	// Fetch all pages using go-fastly helper.
	log.Printf("[DEBUG] Reading API Security operation tags (all pages): %#v", in)
	all, err := operations.ListTagsAll(ctx, conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	// Stable ID based on the full result set.
	hashBase, _ := json.Marshal(all)
	d.SetId(strconv.Itoa(hashcode.String(string(hashBase))))

	if err := d.Set("tags", flattenAPISecurityTagsFromSlice(all)); err != nil {
		return diag.FromErr(err)
	}

	total := first.Meta.Total
	if total == 0 {
		// fallback safety: if API doesn't return meta.total for some reason
		total = len(all)
	}

	_ = d.Set("total", total)
	_ = d.Set("limit_returned", first.Meta.Limit)

	return nil
}

func flattenAPISecurityTagsFromSlice(items []operations.OperationTag) []map[string]any {
	if len(items) == 0 {
		return []map[string]any{}
	}

	out := make([]map[string]any, 0, len(items))
	for _, t := range items {
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
