package fastly

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	eventmappings "github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"
)

// auditLogEventMappingScopeTypes is the single source of truth for valid
// `scope_type` values, so the schema's Description and ValidateDiagFunc
// can't drift out of sync with one another.
var auditLogEventMappingScopeTypes = []string{
	eventmappings.ScopeTypeAccount,
	eventmappings.ScopeTypeVCL,
	eventmappings.ScopeTypeWasm,
	eventmappings.ScopeTypeNGWAF,
}

func resourceFastlyAuditLogEventMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyAuditLogEventMappingCreate,
		ReadContext:   resourceFastlyAuditLogEventMappingRead,
		UpdateContext: resourceFastlyAuditLogEventMappingUpdate,
		DeleteContext: resourceFastlyAuditLogEventMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (RFC3339) when the mapping was created.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the mapping.",
			},
			"event_types": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The audit event types that trigger a notification. Each event type must be valid for the given `scope_type`.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
			},
			"integration_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The IDs of the integrations that should receive notifications. Must reference integrations belonging to the account linked to the configured token.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
			},
			"mapping_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Whether the mapping is permitted to send notifications. One of `active` or `inactive`. A mapping is `active` as long as it has at least one integration ID.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Descriptive name for the mapping.",
			},
			"scope_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The specific service or workspace IDs to scope the mapping to. Omit to apply the mapping to all resources of the given `scope_type`. Must be empty when `scope_type` is `account`.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"scope_type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The category of Fastly resource the mapping applies to. One of `account`, `vcl`, `wasm`, or `ngwaf`.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(auditLogEventMappingScopeTypes, false)),
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (RFC3339) when the mapping was last updated.",
			},
		},
	}
}

func resourceFastlyAuditLogEventMappingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := eventmappings.CreateInput{
		Name:           gofastly.ToPointer(d.Get("name").(string)),
		ScopeType:      gofastly.ToPointer(d.Get("scope_type").(string)),
		EventTypes:     expandStringSet(d.Get("event_types").(*schema.Set)),
		IntegrationIDs: expandStringSet(d.Get("integration_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("description"); ok {
		i.Description = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("scope_ids"); ok {
		i.ScopeIDs = expandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] CREATE: Audit log event mapping input: %#v", i)

	mapping, err := eventmappings.Create(ctx, conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(mapping.ID)

	return resourceFastlyAuditLogEventMappingRead(ctx, d, meta)
}

func resourceFastlyAuditLogEventMappingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := eventmappings.GetInput{
		MappingID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] REFRESH: Audit log event mapping input: %#v", i)

	mapping, err := eventmappings.Get(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Audit log event mapping not found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

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

func resourceFastlyAuditLogEventMappingUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := eventmappings.UpdateInput{
		MappingID:      gofastly.ToPointer(d.Id()),
		Name:           gofastly.ToPointer(d.Get("name").(string)),
		ScopeType:      gofastly.ToPointer(d.Get("scope_type").(string)),
		EventTypes:     expandStringSet(d.Get("event_types").(*schema.Set)),
		IntegrationIDs: expandStringSet(d.Get("integration_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("description"); ok {
		i.Description = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("scope_ids"); ok {
		i.ScopeIDs = expandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] UPDATE: Audit log event mapping input: %#v", i)

	_, err := eventmappings.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyAuditLogEventMappingRead(ctx, d, meta)
}

func resourceFastlyAuditLogEventMappingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := eventmappings.DeleteInput{
		MappingID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] DELETE: Audit log event mapping input: %#v", i)

	if err := eventmappings.Delete(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
