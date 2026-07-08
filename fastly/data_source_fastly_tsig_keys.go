package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v16/fastly/dns/v1/tsigkeys"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyTSIGKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyTSIGKeysRead,
		Schema: map[string]*schema.Schema{
			"keys": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A list of TSIG keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"algorithm": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The algorithm of the TSIG key.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A freeform descriptive note.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "TSIG Key Identifier (UUID).",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the TSIG key.",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "The total number of TSIG keys returned.",
			},
		},
	}
}

func dataSourceFastlyTSIGKeysRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading all TSIG keys")

	keys, err := tsigkeys.List(ctx, conn, &tsigkeys.ListInput{})
	if err != nil {
		return diag.Errorf("error fetching TSIG keys: %s", err)
	}

	hashBase, _ := json.Marshal(keys)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("keys", flattenTSIGKeys(keys)); err != nil {
		return diag.Errorf("error setting keys: %s", err)
	}

	if err := d.Set("total", len(keys)); err != nil {
		return diag.Errorf("error setting total: %s", err)
	}

	return nil
}

func flattenTSIGKeys(keys []tsigkeys.TSIGKey) []map[string]any {
	if len(keys) == 0 {
		return []map[string]any{}
	}

	result := make([]map[string]any, len(keys))
	for i, key := range keys {
		data := map[string]any{
			"algorithm":   "",
			"description": "",
			"id":          "",
			"name":        "",
		}
		if key.Algorithm != nil {
			data["algorithm"] = *key.Algorithm
		}
		if key.Description != nil {
			data["description"] = *key.Description
		}
		if key.ID != nil {
			data["id"] = *key.ID
		}
		if key.Name != nil {
			data["name"] = *key.Name
		}
		result[i] = data
	}

	return result
}
