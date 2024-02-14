package fastly

import (
	"context"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyAlert() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyAlertCreate,
		ReadContext:   resourceFastlyAlertRead,
		UpdateContext: resourceFastlyAlertUpdate,
		DeleteContext: resourceFastlyAlertDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional text that is included in the alert notification.",
			},

			"dimensions": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "More filters depending on the source type.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domains": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Names of a subset of domains that the alert monitors.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"origins": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Addresses of a subset of backends that the alert monitors.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"evaluation_strategy": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Criteria on how to alert.",
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"period": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The length of time to evaluate whether the conditions have been met. The data is polled every minute. One of: `5m`, `15m`, `30m`.",
						},
						"threshold": {
							Type:        schema.TypeFloat,
							Required:    true,
							Description: "Threshold used to alert.",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Type of strategy to use to evaluate. One of: `above_threshold`, `below_threshold`.",
						},
					},
				},
			},

			"integration_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of integrations used to notify when alert fires.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"metric": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The metric name to alert on for a specific source: [domains](https://developer.fastly.com/reference/api/metrics-stats/domain-inspector/historical), [origins](https://developer.fastly.com/reference/api/metrics-stats/origin-inspector/historical), or [stats](https://developer.fastly.com/reference/api/metrics-stats/historical-stats).",
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the alert.",
			},

			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The service which the alert monitors.",
			},

			"source": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The source where the metric comes from. One of: `domains`, `origins`, `stats`.",
			},
		},
	}
}

func resourceFastlyAlertCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := gofastly.CreateAlertDefinitionInput{
		Metric:    gofastly.ToPointer(d.Get("metric").(string)),
		Name:      gofastly.ToPointer(d.Get("name").(string)),
		ServiceID: gofastly.ToPointer(d.Get("service_id").(string)),
		Source:    gofastly.ToPointer(d.Get("source").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}

	if v, ok := d.GetOk("dimensions"); ok {
		for _, r := range v.([]any) {
			if m, ok := r.(map[string]any); ok {
				input.Dimensions = buildDimensions(m)
			}
		}
	}

	if v, ok := d.GetOk("evaluation_strategy"); ok {
		for _, r := range v.([]any) {
			if m, ok := r.(map[string]any); ok {
				input.EvaluationStrategy = buildEvaluationStrategy(m)
			}
		}
	}

	if v, ok := d.GetOk("integration_ids"); ok {
		input.IntegrationIDs = buildStringSlice(v.(*schema.Set))
	} else {
		input.IntegrationIDs = []string{}
	}

	ad, err := conn.CreateAlertDefinition(&input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ad.ID)

	return nil
}

func resourceFastlyAlertRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Alert Configuration for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	ad, err := conn.GetAlertDefinition(&gofastly.GetAlertDefinitionInput{
		ID: gofastly.ToPointer(d.Id()),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if len(ad.Description) > 0 {
		err = d.Set("description", ad.Description)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if len(ad.Dimensions) > 0 {
		err = d.Set("dimensions", flattenDimensions(ad.Dimensions))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if len(ad.EvaluationStrategy) > 0 {
		err = d.Set("evaluation_strategy", []map[string]any{ad.EvaluationStrategy})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if len(ad.IntegrationIDs) > 0 {
		err = d.Set("integration_ids", ad.IntegrationIDs)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = d.Set("metric", ad.Metric)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("name", ad.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("service_id", ad.ServiceID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("source", ad.Source)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyAlertUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := gofastly.UpdateAlertDefinitionInput{
		ID:     gofastly.ToPointer(d.Id()),
		Metric: gofastly.ToPointer(d.Get("metric").(string)),
		Name:   gofastly.ToPointer(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}

	if v, ok := d.GetOk("dimensions"); ok {
		for _, r := range v.([]any) {
			if m, ok := r.(map[string]any); ok {
				input.Dimensions = buildDimensions(m)
			}
		}
	}

	if v, ok := d.GetOk("evaluation_strategy"); ok {
		for _, r := range v.([]any) {
			if m, ok := r.(map[string]any); ok {
				input.EvaluationStrategy = buildEvaluationStrategy(m)
			}
		}
	}

	if v, ok := d.GetOk("integration_ids"); ok {
		input.IntegrationIDs = buildStringSlice(v.(*schema.Set))
	} else {
		input.IntegrationIDs = []string{}
	}

	_, err := conn.UpdateAlertDefinition(&input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyAlertRead(ctx, d, meta)
}

func resourceFastlyAlertDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteAlertDefinition(&gofastly.DeleteAlertDefinitionInput{
		ID: gofastly.ToPointer(d.Id()),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenDimensions(remoteState map[string][]string) []map[string]any {
	data := map[string]any{}
	for k, v := range remoteState {
		s := &schema.Set{F: schema.HashString}
		for _, i := range v {
			s.Add(i)
		}
		data[k] = s
	}
	return []map[string]any{data}
}

func buildDimensions(v map[string]any) map[string][]string {
	dimensions := map[string][]string{}
	for dimension, values := range v {
		dimensions[dimension] = buildStringSlice(values.(*schema.Set))
	}
	return dimensions
}

func buildEvaluationStrategy(v map[string]any) map[string]any {
	return map[string]any{
		"type":      v["type"].(string),
		"period":    v["period"].(string),
		"threshold": v["threshold"].(float64),
	}
}

func buildStringSlice(s *schema.Set) []string {
	l := s.List()
	sl := make([]string, 0, len(l))
	for _, i := range l {
		if v, ok := i.(string); ok {
			sl = append(sl, v)
		}
	}
	return sl
}
