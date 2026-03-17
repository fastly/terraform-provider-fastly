package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"
	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyAPISecurityDiscoveredOperations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyAPISecurityDiscoveredOperationsRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Filter by one or more fully-qualified domains (exact match).",
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
				Description: "Discovered operations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Discovered operation domain.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Discovered operation ID.",
						},
						"last_seen_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last seen timestamp (when present).",
						},
						"method": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Discovered operation HTTP method.",
						},
						"path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Discovered operation path.",
						},
						"rps": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Observed requests per second (when present).",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Discovered operation status (when present).",
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
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Filter discovered operations by status. Accepted values are `DISCOVERED`, `SAVED`, and `IGNORED`.",
				ValidateFunc: validation.StringInSlice([]string{"DISCOVERED", "SAVED", "IGNORED"}, false),
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of matching results, as returned by the API.",
			},
		},
	}
}

func dataSourceFastlyAPISecurityDiscoveredOperationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	serviceID := d.Get("service_id").(string)

	in := &operations.ListDiscoveredInput{
		ServiceID: gofastly.ToPointer(serviceID),
		// Keep pagination internal; go-fastly will paginate across all pages.
		Page:  gofastly.ToPointer(0),
		Limit: gofastly.ToPointer(apiSecurityDefaultPageLimit),
	}

	if v, ok := d.GetOk("status"); ok {
		s := v.(string)
		if s != "" {
			in.Status = gofastly.ToPointer(s)
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
	log.Printf("[DEBUG] Reading API Security discovered operations (page 0): %#v", in)
	first, err := operations.ListDiscovered(ctx, conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	// Fetch all pages using go-fastly helper.
	log.Printf("[DEBUG] Reading API Security discovered operations (all pages): %#v", in)
	all, err := operations.ListDiscoveredAll(ctx, conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	hashBase, _ := json.Marshal(all)
	d.SetId(strconv.Itoa(hashcode.String(string(hashBase))))

	if err := d.Set("operations", flattenDiscoveredOperationsFromSlice(all)); err != nil {
		return diag.FromErr(err)
	}

	total := first.Meta.Total
	if total == 0 {
		total = len(all)
	}
	_ = d.Set("total", total)

	return nil
}

func flattenDiscoveredOperationsFromSlice(items []operations.DiscoveredOperation) []map[string]any {
	if len(items) == 0 {
		return []map[string]any{}
	}

	out := make([]map[string]any, 0, len(items))
	for _, op := range items {
		out = append(out, map[string]any{
			"domain":       op.Domain,
			"id":           op.ID,
			"last_seen_at": op.LastSeenAt,
			"method":       op.Method,
			"path":         op.Path,
			"rps":          op.RPS,
			"status":       op.Status,
			"updated_at":   op.UpdatedAt,
		})
	}
	return out
}
