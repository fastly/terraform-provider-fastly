package fastly

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
)

const (
	badAlertSourceServiceIDConfig = "empty `service_id` is only supported for `stats` as a source"
	ManagedByTerraform            = "Managed by Terraform"
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
		CustomizeDiff: validateDimensionsForSource,

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
				Description: "Additional filters and aggregation criteria depending on the source type. For source type `stats`, use empty `{}` to monitor a single service or all services (aggregated), or include the `services` array to monitor all or a set of services (not aggregated). For source types `domains` or `origins`, the `domains` or `origins` array may be included to monitor all or a set (not aggregated).",
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
						"services": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "List of service IDs to monitor individually, or empty array `[]` to monitor all services.",
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
						"ignore_below": {
							Type:        schema.TypeFloat,
							Optional:    true,
							Description: "Threshold for the denominator value used in evaluations that calculate a rate or ratio. Usually used to filter out noise.",
						},
						"period": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The length of time to evaluate whether the conditions have been met. The data is polled every minute. One of: `2m`, `3m`, `5m`, `15m`, `30m`.",
						},
						"threshold": {
							Type:        schema.TypeFloat,
							Required:    true,
							Description: "Threshold used to alert.",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Type of strategy to use to evaluate. One of: `above_threshold`, `all_above_threshold`, `below_threshold`, `percent_absolute`, `percent_decrease`, `percent_increase`.",
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
				Optional:    true,
				ForceNew:    true,
				Description: "The service to monitor. Can be a specific service ID to monitor a single service, or omit to monitor all services (aggregated) or all or a set of services (not aggregated) (see `dimensions` for details).",
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

func resourceFastlyAlertCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := validateSourceWithServiceID(d.Get("source").(string), d.Get("service_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	input := gofastly.CreateAlertDefinitionInput{
		Metric:    gofastly.ToPointer(d.Get("metric").(string)),
		Name:      gofastly.ToPointer(d.Get("name").(string)),
		ServiceID: gofastly.ToPointer(d.Get("service_id").(string)),
		Source:    gofastly.ToPointer(d.Get("source").(string)),
	}

	description := ManagedByTerraform
	if v, ok := d.GetOk("description"); ok {
		description = v.(string) + " " + ManagedByTerraform
	}
	input.Description = gofastly.ToPointer(description)

	input.Dimensions = buildDimensions(d)

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

	ad, err := conn.CreateAlertDefinition(gofastly.NewContextForResourceID(ctx, d.Get("service_id").(string)), &input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ad.ID)

	return nil
}

func resourceFastlyAlertRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Alert Configuration for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	ad, err := conn.GetAlertDefinition(gofastly.NewContextForResourceID(ctx, d.Get("service_id").(string)), &gofastly.GetAlertDefinitionInput{
		ID: gofastly.ToPointer(d.Id()),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	description := strings.TrimSpace(strings.TrimSuffix(ad.Description, ManagedByTerraform))
	if len(description) > 0 {
		err = d.Set("description", description)
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

	err := validateSourceWithServiceID(d.Get("source").(string), d.Get("service_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	input := gofastly.UpdateAlertDefinitionInput{
		ID:     gofastly.ToPointer(d.Id()),
		Metric: gofastly.ToPointer(d.Get("metric").(string)),
		Name:   gofastly.ToPointer(d.Get("name").(string)),
	}

	description := ManagedByTerraform
	if v, ok := d.GetOk("description"); ok {
		description = v.(string) + " " + ManagedByTerraform
	}
	input.Description = gofastly.ToPointer(description)

	input.Dimensions = buildDimensions(d)

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

	_, err = conn.UpdateAlertDefinition(gofastly.NewContextForResourceID(ctx, d.Get("service_id").(string)), &input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyAlertRead(ctx, d, meta)
}

func resourceFastlyAlertDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteAlertDefinition(gofastly.NewContextForResourceID(ctx, d.Get("service_id").(string)), &gofastly.DeleteAlertDefinitionInput{
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

func getDimensionsMapFromRawConfig(c cty.Value) (map[string]cty.Value, bool) {
	if c.IsNull() {
		return nil, false
	}

	m := c.AsValueMap()
	dimensionsVal, ok := m["dimensions"]
	if !ok || dimensionsVal.IsNull() || dimensionsVal.LengthInt() == 0 {
		return nil, false
	}

	dimsList := dimensionsVal.AsValueSlice()
	if len(dimsList) == 0 {
		return nil, false
	}

	return dimsList[0].AsValueMap(), true
}

func buildDimensions(d *schema.ResourceData) map[string][]string {
	data := map[string][]string{}

	dimsMap, ok := getDimensionsMapFromRawConfig(d.GetRawConfig())
	if !ok {
		return data
	}

	var stateMap map[string]any
	if v := d.Get("dimensions").([]any); len(v) > 0 && v[0] != nil {
		stateMap, _ = v[0].(map[string]any)
	}

	for key, val := range dimsMap {
		if !val.IsNull() {
			if stateMap != nil {
				if set, ok := stateMap[key].(*schema.Set); ok {
					data[key] = buildStringSlice(set)
					continue
				}
			}
			data[key] = []string{}
		}
	}

	return data
}

func buildEvaluationStrategy(v map[string]any) map[string]any {
	// Required attributes
	m := map[string]any{
		"type":      v["type"].(string),
		"period":    v["period"].(string),
		"threshold": v["threshold"].(float64),
	}

	// Optional attributes
	if value, ok := v["ignore_below"]; ok {
		if v := value.(float64); v > 0 {
			m["ignore_below"] = v
		}
	}

	return m
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

func validateSourceWithServiceID(source string, serviceID string) error {
	if source != "stats" && serviceID == "" {
		return errors.New(badAlertSourceServiceIDConfig)
	}

	return nil
}

func validateDimensionsForSource(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	source := d.Get("source").(string)

	dimsMap, ok := getDimensionsMapFromRawConfig(d.GetRawConfig())
	if !ok {
		return nil
	}

	switch source {
	case "origins":
		originsVal, hasOrigins := dimsMap["origins"]
		if !hasOrigins || originsVal.IsNull() {
			return errors.New("dimensions.origins must be explicitly set when using source 'origins'")
		}
		for key := range dimsMap {
			if key != "origins" && !dimsMap[key].IsNull() {
				return errors.New("only dimensions.origins is valid for source 'origins'")
			}
		}
	case "domains":
		domainsVal, hasDomains := dimsMap["domains"]
		if !hasDomains || domainsVal.IsNull() {
			return errors.New("dimensions.domains must be explicitly set when using source 'domains'")
		}
		for key := range dimsMap {
			if key != "domains" && !dimsMap[key].IsNull() {
				return errors.New("only dimensions.domains is valid for source 'domains'")
			}
		}
	case "stats":
		for key := range dimsMap {
			if key != "services" && !dimsMap[key].IsNull() {
				return errors.New("only dimensions.services is valid for source 'stats'")
			}
		}
	}

	return nil
}
