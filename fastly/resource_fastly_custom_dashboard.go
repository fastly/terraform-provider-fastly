package fastly

import (
	"context"
	"errors"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	schemaDataSource = schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "An object which describes the data to display.",
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"config": &schemaDataSourceConfig,
				"type": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The source of the data to display. One of: `stats.edge`, `stats.domain`, `stats.origin`.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
						[]string{"stats.edge", "stats.domain", "stats.origin"},
						false,
					)),
				},
			},
		},
	}

	schemaDataSourceConfig = schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "Configuration options for the selected data source.",
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"metrics": {
					Type:        schema.TypeList,
					Required:    true,
					Description: "The metrics to visualize. Valid options are defined by the selected data source: [stats.edge](https://www.fastly.com/documentation/reference/api/observability/custom-dashboards/metrics/edge/), [stats.domain](https://www.fastly.com/documentation/reference/api/observability/custom-dashboards/metrics/domain/), [stats.origin](https://www.fastly.com/documentation/reference/api/observability/custom-dashboards/metrics/origin/).",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}

	schemaVisualization = schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "An object which describes the data visualization to display.",
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"config": &schemaVisualizationConfig,
				"type": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The type of visualization to display. One of: `chart`.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
						[]string{"chart"},
						false,
					)),
				},
			},
		},
	}

	schemaVisualizationConfig = schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "Configuration options for the selected data source.",
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"calculation_method": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The aggregation function to apply to the dataset. One of: `avg`, `sum`, `min`, `max`, `latest`, `p95`.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
						[]string{"avg", "sum", "min", "max", "latest", "p95"},
						false,
					)),
				},
				"format": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The units to use to format the data. One of: `number`, `bytes`, `percent`, `requests`, `responses`, `seconds`, `milliseconds`, `ratio`, `bitrate`.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
						[]string{"number", "bytes", "percent", "requests", "responses"},
						false,
					)),
				},
				"plot_type": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The type of chart to display. One of: `line`, `bar`, `single-metric`, `donut`.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
						[]string{"line", "bar", "single-metric", "donut"},
						false,
					)),
				},
			},
		},
	}
)

func resourceFastlyCustomDashboard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyCustomDashboardCreate,
		ReadContext:   resourceFastlyCustomDashboardRead,
		UpdateContext: resourceFastlyCustomDashboardUpdate,
		DeleteContext: resourceFastlyCustomDashboardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"dashboard_item": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of dashboard items.",
				MinItems:    0,
				MaxItems:    100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":          {Type: schema.TypeString, Computed: true},
						"data_source": &schemaDataSource,
						"span": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     4,
							Description: "The number of columns for the dashboard item to span. Dashboards are rendered on a 12-column grid on \"desktop\" screen sizes.",
						},
						"subtitle": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A human-readable subtitle for the dashboard item. Often a description of the visualization.",
						},
						"title": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A human-readable title for the dashboard item.",
						},
						"visualization": &schemaVisualization,
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A short description of the dashboard.",
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A human-readable name.",
			},
		},
	}
}

func resourceFastlyCustomDashboardCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := gofastly.CreateObservabilityCustomDashboardInput{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}

	if v, ok := d.GetOk("dashboard_item"); ok {
		var errs []error
		for _, r := range v.([]interface{}) {
			if m, ok := r.(map[string]any); ok {
				di, err := mapToDashboardItem(m)
				if err != nil {
					errs = append(errs, err)
				} else {
					input.Items = append(input.Items, *di)
				}
			}
		}
		if errs != nil {
			return diag.FromErr(errors.Join(errs...))
		}
	}

	dash, err := conn.CreateObservabilityCustomDashboard(&input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dash.ID)

	return nil
}

func mapToDashboardItem(m map[string]any) (*gofastly.DashboardItem, error) {
	var errs []error
	var (
		title, subtitle string
		span            int

		dataSourceSet, sourceConfigSet *schema.Set
		dataSource, sourceConfig       map[string]any
		sourceType                     string
		metrics                        []string

		vizSet, vizConfigSet               *schema.Set
		visualization, visualizationConfig map[string]any
		visualizationType                  string
		plotType                           string
		format, calcMethod                 *string
	)

	var ok bool

	// Top-level fields
	if title, ok = m["title"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid title: %#v", m["title"]))
	}
	if subtitle, ok = m["subtitle"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid subtitle: %#v", m["subtitle"]))
	}
	if span, ok = m["span"].(int); !ok {
		errs = append(errs, fmt.Errorf("invalid span: %#v", m["span"]))
	}
	if dataSourceSet, ok = m["data_source"].(*schema.Set); !ok {
		errs = append(errs, fmt.Errorf("invalid data_source: %#v", m["data_source"]))
	} else {
		if dsl := dataSourceSet.List(); len(dsl) != 1 {
			errs = append(errs, fmt.Errorf("invalid data_source: %#v", m["data_source"]))
		} else {
			if dataSource, ok = dsl[0].(map[string]any); !ok {
				errs = append(errs, fmt.Errorf("invalid data_source: %#v", m["data_source"]))
			}
		}
	}
	if vizSet, ok = m["visualization"].(*schema.Set); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization: %#v", m["visualization"]))
	} else {
		if vizList := vizSet.List(); len(vizList) != 1 {
			errs = append(errs, fmt.Errorf("invalid visualization: %#v", m["visualization"]))
		} else {
			if visualization, ok = vizList[0].(map[string]any); !ok {
				errs = append(errs, fmt.Errorf("invalid visualization: %#v", m["visualization"]))
			}
		}
	}

	// Nested DataSource
	if sourceType, ok = dataSource["type"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid data_source.type: %#v", dataSource["type"]))
	}
	if sourceConfigSet, ok = dataSource["config"].(*schema.Set); !ok {
		errs = append(errs, fmt.Errorf("invalid data_source.config: %#v", dataSource["config"]))
	}
	if scl := sourceConfigSet.List(); len(scl) != 1 {
		errs = append(errs, fmt.Errorf("invalid data_source.config: %#v", dataSource["config"]))
	} else {
		if sourceConfig, ok = scl[0].(map[string]any); !ok {
			errs = append(errs, fmt.Errorf("invalid data_source.config: %#v", dataSource["config"]))
		}
	}
	if metricsList, ok := sourceConfig["metrics"].([]any); !ok {
		errs = append(errs, fmt.Errorf("invalid data_source.config.metrics: %#v", sourceConfig["metrics"]))
	} else {
		for _, m := range metricsList {
			metrics = append(metrics, m.(string))
		}
	}

	// Nested Visualization
	if visualizationType, ok = visualization["type"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization.type: %#v", visualization["type"]))
	}
	if vizConfigSet, ok = visualization["config"].(*schema.Set); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization.config: %#v", visualization["config"]))
	}
	if vcl := vizConfigSet.List(); len(vcl) != 1 {
		errs = append(errs, fmt.Errorf("invalid visualization.config: %#v", visualization["config"]))
	} else {
		if visualizationConfig, ok = vcl[0].(map[string]any); !ok {
			errs = append(errs, fmt.Errorf("invalid visualization.config: %#v", visualization["config"]))
		}
	}
	if plotType, ok = visualizationConfig["plot_type"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization.config.plot_type: %#v", visualizationConfig["plot_type"]))
	}
	format, _ = visualizationConfig["format"].(*string)
	calcMethod, _ = visualizationConfig["calculation_method"].(*string)

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &gofastly.DashboardItem{
		DataSource: gofastly.DashboardDataSource{
			Config: gofastly.DashboardSourceConfig{
				Metrics: metrics,
			},
			Type: gofastly.DashboardSourceType(sourceType),
		},
		Span:     uint8(span),
		Subtitle: subtitle,
		Title:    title,
		Visualization: gofastly.DashboardVisualization{
			Config: gofastly.VisualizationConfig{
				CalculationMethod: (*gofastly.CalculationMethod)(calcMethod),
				Format:            (*gofastly.VisualizationFormat)(format),
				PlotType:          gofastly.PlotType(plotType),
			},
			Type: gofastly.VisualizationType(visualizationType),
		},
	}, nil
}

func dashboardItemToMap(di gofastly.DashboardItem) map[string]interface{} {
	var metrics []any
	for _, m := range di.DataSource.Config.Metrics {
		metrics = append(metrics, m)
	}
	sourceConfigSet := schema.NewSet(schema.HashResource(schemaDataSourceConfig.Elem.(*schema.Resource)), nil)
	sourceConfigSet.Add(map[string]any{"metrics": metrics})

	dataSourceSet := schema.NewSet(schema.HashResource(schemaDataSource.Elem.(*schema.Resource)), nil)
	dataSourceSet.Add(map[string]any{"type": string(di.DataSource.Type), "config": sourceConfigSet})

	var format, calcMethod string
	if di.Visualization.Config.CalculationMethod != nil {
		calcMethod = string(*di.Visualization.Config.CalculationMethod)
	}
	if di.Visualization.Config.Format != nil {
		format = string(*di.Visualization.Config.Format)
	}
	vizConfigSet := schema.NewSet(schema.HashResource(schemaVisualizationConfig.Elem.(*schema.Resource)), nil)
	vizConfigSet.Add(map[string]any{"plot_type": string(di.Visualization.Config.PlotType), "format": format, "calculation_method": calcMethod})

	visualizationSet := schema.NewSet(schema.HashResource(schemaVisualization.Elem.(*schema.Resource)), nil)
	visualizationSet.Add(map[string]any{"type": string(di.Visualization.Type), "config": vizConfigSet})
	return map[string]interface{}{
		"id":            di.ID,
		"title":         di.Title,
		"subtitle":      di.Subtitle,
		"span":          di.Span,
		"data_source":   dataSourceSet,
		"visualization": visualizationSet,
	}
}

func resourceFastlyCustomDashboardRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Custom Dashboard for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	dash, err := conn.GetObservabilityCustomDashboard(&gofastly.GetObservabilityCustomDashboardInput{
		ID: gofastly.ToPointer(d.Id()),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dash.ID)
	if len(dash.Name) > 0 {
		err = d.Set("name", dash.Name)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if len(dash.Description) > 0 {
		err = d.Set("description", dash.Description)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if len(dash.Items) > 0 {
		var itemMaps []map[string]interface{}
		for _, di := range dash.Items {
			itemMaps = append(itemMaps, dashboardItemToMap(di))
		}
		err = d.Set("dashboard_item", itemMaps)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceFastlyCustomDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := gofastly.UpdateObservabilityCustomDashboardInput{
		Description: gofastly.ToPointer(d.Get("description").(string)),
		ID:          gofastly.ToPointer(d.Id()),
		Name:        gofastly.ToPointer(d.Get("name").(string)),
	}
	// if v, ok := d.GetOk("dashboard_items"); ok {
	// 	for i, r := range v.([]any) {
	// 		if _, ok := r.(map[string]any); ok {
	// 			(*input.Items)[i].Title = r.Get("title").(string)
	// 		}
	// 	}
	// }
	_, err := conn.UpdateObservabilityCustomDashboard(&input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyCustomDashboardRead(ctx, d, meta)
}

func resourceFastlyCustomDashboardDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteObservabilityCustomDashboard(&gofastly.DeleteObservabilityCustomDashboardInput{
		ID: gofastly.ToPointer(d.Id()),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
