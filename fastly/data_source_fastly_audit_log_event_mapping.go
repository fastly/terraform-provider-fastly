package fastly

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	eventmappings "github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"
)

func dataSourceFastlyAuditLogEventMapping() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyAuditLogEventMappingRead,

		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (RFC3339) when the mapping was created.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the mapping.",
			},
			"event_types": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The audit event types that trigger a notification.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Unique ID of the mapping. Conflicts with all the other filters.",
				ConflictsWith: []string{"integration_id", "mapping_status", "name", "scope_id", "scope_type"},
			},
			"integration_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Filters results to mappings that reference the given integration ID.",
				ConflictsWith: []string{"id"},
			},
			"integration_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The IDs of the integrations that receive notifications.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"mapping_status": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Filters results by mapping status: `active` or `inactive`.",
				ConflictsWith: []string{"id"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Filters results to mappings whose name contains the given string (case-insensitive).",
				ConflictsWith: []string{"id"},
			},
			"scope_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Filters results to mappings that apply to the given service or workspace ID.",
				ConflictsWith: []string{"id"},
			},
			"scope_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The specific service or workspace IDs the mapping is scoped to.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"scope_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Filters results to the given scope type: `account`, `vcl`, `wasm`, or `ngwaf`.",
				ConflictsWith: []string{"id"},
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (RFC3339) when the mapping was last updated.",
			},
		},
	}
}

func dataSourceFastlyAuditLogEventMappingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	var mapping *eventmappings.EventMapping

	if v, ok := d.GetOk("id"); ok {
		m, err := eventmappings.Get(ctx, conn, &eventmappings.GetInput{
			MappingID: gofastly.ToPointer(v.(string)),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		mapping = m
	} else {
		i := eventmappings.ListInput{}
		if v, ok := d.GetOk("name"); ok {
			i.Name = gofastly.ToPointer(v.(string))
		}
		if v, ok := d.GetOk("scope_type"); ok {
			i.ScopeType = gofastly.ToPointer(v.(string))
		}
		if v, ok := d.GetOk("scope_id"); ok {
			i.ScopeID = gofastly.ToPointer(v.(string))
		}
		if v, ok := d.GetOk("integration_id"); ok {
			i.IntegrationID = gofastly.ToPointer(v.(string))
		}
		if v, ok := d.GetOk("mapping_status"); ok {
			i.MappingStatus = gofastly.ToPointer(v.(string))
		}

		mappings, err := eventmappings.List(ctx, conn, &i)
		if err != nil {
			return diag.FromErr(err)
		}

		if len(mappings) == 0 {
			return diag.Errorf("your query returned no results. Please change your search criteria and try again.")
		}
		if len(mappings) > 1 {
			return diag.Errorf("your query returned more than one result. Please try a more specific search criteria and try again.")
		}

		mapping = &mappings[0]
	}

	d.SetId(mapping.ID)
	if err := d.Set("name", mapping.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", mapping.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("scope_type", mapping.ScopeType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("scope_ids", mapping.ScopeIDs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("event_types", mapping.EventTypes); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("integration_ids", mapping.IntegrationIDs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mapping_status", mapping.MappingStatus); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", mapping.CreatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", mapping.UpdatedAt.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
