package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v16/fastly/dns/v1/dnszones"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyDNSZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyDNSZonesRead,
		Schema: map[string]*schema.Schema{
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "The total number of DNS zones returned.",
			},
			"zones": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A list of DNS zones.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A freeform descriptive note.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Zone Identifier (UUID).",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The domain name for the zone.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyDNSZonesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading all DNS zones")

	zones, err := dnszones.List(ctx, conn, &dnszones.ListInput{})
	if err != nil {
		return diag.Errorf("error fetching DNS zones: %s", err)
	}

	hashBase, _ := json.Marshal(zones)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("zones", flattenDNSZones(zones)); err != nil {
		return diag.Errorf("error setting zones: %s", err)
	}

	if err := d.Set("total", len(zones)); err != nil {
		return diag.Errorf("error setting total: %s", err)
	}

	return nil
}

func flattenDNSZones(zones []dnszones.Zone) []map[string]any {
	if len(zones) == 0 {
		return []map[string]any{}
	}

	result := make([]map[string]any, len(zones))
	for i, zone := range zones {
		data := map[string]any{
			"id":          "",
			"name":        "",
			"description": "",
		}
		if zone.ID != nil {
			data["id"] = *zone.ID
		}
		if zone.Name != nil {
			data["name"] = *zone.Name
		}
		if zone.Description != nil {
			data["description"] = *zone.Description
		}
		result[i] = data
	}

	return result
}
