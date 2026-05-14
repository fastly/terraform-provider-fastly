package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/dns/v1/dnszones"
)

func resourceFastlyDNSZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyDNSZoneCreate,
		ReadContext:   resourceFastlyDNSZoneRead,
		UpdateContext: resourceFastlyDNSZoneUpdate,
		DeleteContext: resourceFastlyDNSZoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description for your dns zone.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The domain name for your zone in FQDN format (e.g. `example.com.`). Must include a trailing dot.",
				// The API requires zone names to be in FQDN format with a trailing dot, so we are adding validation accordingly.
				ValidateFunc: func(v any, k string) (warnings []string, errors []error) {
					if !strings.HasSuffix(v.(string), ".") {
						errors = append(errors, fmt.Errorf("%q must be in FQDN format, ending with a trailing dot (e.g. `example.com.`)", k))
					}
					return
				},
			},
			"xfr_config_inbound": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "All attributes associated with inbound zone transfers.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inbound_tsig_key_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the TSIG key used to secure inbound zone transfers.",
						},
						"primaries": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Primary DNS Servers",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"address": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "An IPv4 address for the Primary DNS Server.",
									},
									"description": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "A description of the Primary DNS server.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceFastlyDNSZoneCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	var input dnszones.CreateInput
	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		input.Name = gofastly.ToPointer(v.(string))
	}
	// Type can only have the value of [secondary], so we'll set that here as it's a required field.
	input.Type = gofastly.ToPointer("secondary")

	if v, ok := d.GetOk("xfr_config_inbound"); ok {
		for _, r := range v.([]any) {
			if m, ok := r.(map[string]any); ok {
				xfrInput := &dnszones.XfrConfigInboundInput{}

				if v, ok := m["inbound_tsig_key_id"].(string); ok && v != "" {
					// NewNullable allows the value to be explicitly set to null
					//  in JSON (e.g. to clear the field on update).
					xfrInput.InboundTSIGKeyID = gofastly.NewNullable(v)
				}

				if primaries, ok := m["primaries"].([]any); ok {
					for _, p := range primaries {
						if pm, ok := p.(map[string]any); ok {
							primary := dnszones.Primary{}
							if v, ok := pm["address"].(string); ok {
								primary.Address = gofastly.ToPointer(v)
							}
							if v, ok := pm["description"].(string); ok {
								primary.Description = gofastly.ToPointer(v)
							}
							xfrInput.Primaries = append(xfrInput.Primaries, primary)
						}
					}
				}
				input.XfrConfigInbound = xfrInput
			}
		}
	}
	data, err := dnszones.Create(ctx, conn, &input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*data.ID)
	return resourceFastlyDNSZoneRead(ctx, d, meta)
}

func resourceFastlyDNSZoneRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing DNS Zone for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	input := &dnszones.GetInput{
		ZoneID: gofastly.ToPointer(d.Id()),
	}

	data, err := dnszones.Get(ctx, conn, input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("description", data.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", data.Name); err != nil {
		return diag.FromErr(err)
	}
	var xfrConfigInbound []map[string]any
	if xfr := data.XfrConfigInbound; xfr != nil && (xfr.InboundTSIGKeyID != nil || len(xfr.Primaries) > 0) {
		xfrConfigInbound = flattenXfrConfigInbound(xfr)
	}
	if err := d.Set("xfr_config_inbound", xfrConfigInbound); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyDNSZoneUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &dnszones.UpdateInput{
		ZoneID: gofastly.ToPointer(d.Id()),
	}
	if d.HasChange("description") {
		input.Description = gofastly.NewNullable(d.Get("description").(string))
	}
	if v, ok := d.GetOk("xfr_config_inbound"); ok {
		for _, r := range v.([]any) {
			if m, ok := r.(map[string]any); ok {
				xfrInput := &dnszones.XfrConfigInboundInput{}

				if v, ok := m["inbound_tsig_key_id"].(string); ok && v != "" {
					xfrInput.InboundTSIGKeyID = gofastly.NewNullable(v)
				} else {
					// NullValue serializes as JSON null, explicitly clearing the field on update.
					xfrInput.InboundTSIGKeyID = gofastly.NullValue[string]()
				}

				if primaries, ok := m["primaries"].([]any); ok {
					for _, p := range primaries {
						if pm, ok := p.(map[string]any); ok {
							primary := dnszones.Primary{}
							if v, ok := pm["address"].(string); ok {
								primary.Address = gofastly.ToPointer(v)
							}
							if v, ok := pm["description"].(string); ok {
								primary.Description = gofastly.ToPointer(v)
							}
							xfrInput.Primaries = append(xfrInput.Primaries, primary)
						}
					}
				}
				input.XfrConfigInbound = xfrInput
			}
		}
	}

	_, err := dnszones.Update(ctx, conn, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyDNSZoneRead(ctx, d, meta)
}

func resourceFastlyDNSZoneDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	input := &dnszones.DeleteInput{
		ZoneID: gofastly.ToPointer(d.Id()),
	}
	err := dnszones.Delete(ctx, conn, input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			return nil
		}
		return diag.FromErr(err)
	}
	return nil
}

func flattenXfrConfigInbound(xfr *dnszones.XfrConfigInbound) []map[string]any {
	m := map[string]any{}
	if xfr.InboundTSIGKeyID != nil {
		m["inbound_tsig_key_id"] = *xfr.InboundTSIGKeyID
	}
	if len(xfr.Primaries) > 0 {
		primaries := make([]map[string]any, len(xfr.Primaries))
		for i, p := range xfr.Primaries {
			pm := map[string]any{}
			if p.Address != nil {
				pm["address"] = *p.Address
			}
			if p.Description != nil {
				pm["description"] = *p.Description
			}
			primaries[i] = pm
		}
		m["primaries"] = primaries
	}
	return []map[string]any{m}
}
