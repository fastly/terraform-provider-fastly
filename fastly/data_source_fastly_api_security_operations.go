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

func dataSourceFastlyAPISecurityOperations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyAPISecurityOperationsRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Filter by one or more domains (exact match).",
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "A domain value used for filtering.",
				},
			},
			"method": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Filter by one or more HTTP methods.",
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "An HTTP method used for filtering.",
				},
			},
			"operations": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Matching API Security operations.",
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
							Description: "Operation description (when present).",
						},
						"domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Operation domain.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Operation ID.",
						},
						"last_seen_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last seen timestamp (when present).",
						},
						"method": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Operation HTTP method.",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Operation path.",
						},
						"rps": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Observed requests per second (when present).",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Discovery status (when present). One of `DISCOVERED`, `SAVED`, or `IGNORED`.",
						},
						"tag_ids": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Associated operation tag IDs.",
							Elem: &schema.Schema{
								Type:        schema.TypeString,
								Description: "A tag ID.",
							},
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Updated timestamp (when present).",
						},
					},
				},
			},
			"path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by path (exact match).",
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service ID.",
			},
			"tag_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by tag ID.",
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of matching results, as returned by the API.",
			},
		},
	}
}

func dataSourceFastlyAPISecurityOperationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	serviceID := d.Get("service_id").(string)

	in := &operations.ListOperationsInput{
		ServiceID: gofastly.ToPointer(serviceID),
		// Keep pagination internal; go-fastly will paginate across all pages.
		Page:  gofastly.ToPointer(0),
		Limit: gofastly.ToPointer(apiSecurityDefaultPageLimit),
	}

	if v, ok := d.GetOk("tag_id"); ok {
		s := v.(string)
		if s != "" {
			in.TagID = gofastly.ToPointer(s)
		}
	}
	if v, ok := d.GetOk("method"); ok {
		in.Method = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("domain"); ok {
		in.Domain = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("path"); ok {
		s := v.(string)
		if s != "" {
			in.Path = gofastly.ToPointer(s)
		}
	}

	// Capture meta.total from the first page.
	log.Printf("[DEBUG] Reading API Security operations (page 0): %#v", in)
	first, err := operations.ListOperations(ctx, conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	// Fetch all pages using go-fastly helper.
	log.Printf("[DEBUG] Reading API Security operations (all pages): %#v", in)
	all, err := operations.ListOperationsAll(ctx, conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	// Stable ID based on the full result set.
	hashBase, _ := json.Marshal(all)
	d.SetId(strconv.Itoa(hashcode.String(string(hashBase))))

	if err := d.Set("operations", flattenAPISecurityOperationsFromSlice(all)); err != nil {
		return diag.FromErr(err)
	}

	total := first.Meta.Total
	if total == 0 {
		total = len(all)
	}
	_ = d.Set("total", total)

	return nil
}

func flattenAPISecurityOperationsFromSlice(items []operations.Operation) []map[string]any {
	if len(items) == 0 {
		return []map[string]any{}
	}

	out := make([]map[string]any, 0, len(items))
	for _, op := range items {
		m := map[string]any{
			"created_at":   op.CreatedAt,
			"description":  op.Description,
			"domain":       op.Domain,
			"id":           op.ID,
			"last_seen_at": op.LastSeenAt,
			"method":       op.Method,
			"path":         op.Path,
			"rps":          op.RPS,
			"status":       op.Status,
			"updated_at":   op.UpdatedAt,
		}
		m["tag_ids"] = flattenStringSliceToSet(op.TagIDs)
		out = append(out, m)
	}
	return out
}
