package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/usagemetrics"
)

func dataSourceFastlyAIRuntimeControlUsageMetrics() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyAIRuntimeControlUsageMetricsRead,
		Schema: map[string]*schema.Schema{
			"from": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The inclusive start of the time range to retrieve usage metrics for, in RFC 3339 format.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
			},
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter the results by virtual key ID.",
			},
			"model": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter the results by AI model identifier.",
			},
			// The Terraform SDK reserves the field name "provider" on every
			// data source, so the API's `provider` filter is exposed as
			// `provider_name`.
			"provider_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter the results by AI model provider name.",
			},
			"to": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The inclusive end of the time range to retrieve usage metrics for, in RFC 3339 format. Defaults to now.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The total number of usage metric records returned.",
			},
			"usage_metrics": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The usage metric records matching the given filters.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date of the usage record.",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The AI model identifier.",
						},
						"provider": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The AI provider name.",
						},
						"quantity": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The quantity of the usage type.",
						},
						"usage_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of usage being measured. One of `requests`, `sessions`, `input_tokens`, `output_tokens`.",
						},
						"virtual_key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the virtual key.",
						},
						"virtual_key_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The human-readable name of the virtual key.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyAIRuntimeControlUsageMetricsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading AI Runtime Control usage metrics")

	i := usagemetrics.ListInput{}
	if v, ok := d.GetOk("key"); ok {
		i.Key = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("provider_name"); ok {
		i.Provider = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("model"); ok {
		i.Model = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("from"); ok {
		from, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		i.From = &from
	}
	if v, ok := d.GetOk("to"); ok {
		to, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		i.To = &to
	}

	remoteState, err := usagemetrics.List(ctx, conn, &i)
	if err != nil {
		return diag.Errorf("error fetching AI Runtime Control usage metrics: %s", err)
	}

	all := remoteState.Data

	hashBase, _ := json.Marshal(all)
	d.SetId(strconv.Itoa(hashcode.String(string(hashBase))))

	if err := d.Set("usage_metrics", flattenAIRuntimeControlUsageMetrics(all)); err != nil {
		return diag.Errorf("error setting usage_metrics: %s", err)
	}
	if err := d.Set("total", len(all)); err != nil {
		return diag.Errorf("error setting total: %s", err)
	}

	return nil
}

func flattenAIRuntimeControlUsageMetrics(remoteState []usagemetrics.UsageMetric) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, m := range remoteState {
		result[i] = map[string]any{
			"date":             m.Date,
			"usage_type":       m.UsageType,
			"quantity":         m.Quantity,
			"virtual_key_id":   m.VirtualKeyID,
			"virtual_key_name": m.VirtualKeyName,
			"provider":         m.Provider,
			"model":            m.Model,
		}
	}

	return result
}
