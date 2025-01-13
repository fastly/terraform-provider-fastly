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
		Type:        schema.TypeList,
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
		Type:        schema.TypeList,
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
		Type:        schema.TypeList,
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
		Type:        schema.TypeList,
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
						[]string{"number", "bytes", "percent", "requests", "responses", "seconds", "milliseconds", "ratio", "bitrate"},
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
						"data_source": &schemaDataSource,
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Dashboard item identifier (alphanumeric). Must be unique, relative to other items in the same dashboard.",
						},
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
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Dashboard identifier (UUID).",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A human-readable name.",
			},
		},
	}
}

func resourceFastlyCustomDashboardCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := gofastly.CreateObservabilityCustomDashboardInput{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}

	items, err := resourceItems(d)
	if err != nil {
		return diag.FromErr(err)
	}
	input.Items = items

	dash, err := conn.CreateObservabilityCustomDashboard(&input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dash.ID)

	return resourceFastlyCustomDashboardRead(ctx, d, meta)
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
		itemList := flattenDashboardItems(dash.Items)
		err = d.Set("dashboard_item", itemList)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceFastlyCustomDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	var items = make([]gofastly.DashboardItem, 0)
	input := gofastly.UpdateObservabilityCustomDashboardInput{
		Description: gofastly.ToPointer(d.Get("description").(string)),
		ID:          gofastly.ToPointer(d.Id()),
		Name:        gofastly.ToPointer(d.Get("name").(string)),
	}

	items, err := resourceItems(d)
	if err != nil {
		return diag.FromErr(err)
	}
	input.Items = &items

	_, err = conn.UpdateObservabilityCustomDashboard(&input)
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

func flattenDashboardItems(remoteState []gofastly.DashboardItem) []map[string]interface{} {
	var result []map[string]any
	for _, di := range remoteState {
		dataSource := map[string]any{
			"type":   di.DataSource.Type,
			"config": []any{map[string]any{"metrics": di.DataSource.Config.Metrics}},
		}

		vizConfig := map[string]any{
			"plot_type": di.Visualization.Config.PlotType,
		}
		if di.Visualization.Config.CalculationMethod != nil {
			vizConfig["calculation_method"] = string(*di.Visualization.Config.CalculationMethod)
		}
		if di.Visualization.Config.Format != nil {
			vizConfig["format"] = string(*di.Visualization.Config.Format)
		}
		visualization := map[string]any{
			"type":   di.Visualization.Type,
			"config": []any{vizConfig},
		}

		result = append(result, map[string]interface{}{
			"id":            di.ID,
			"title":         di.Title,
			"subtitle":      di.Subtitle,
			"span":          di.Span,
			"data_source":   []any{dataSource},
			"visualization": []any{visualization},
		})
	}
	return result
}

func resourceItems(d *schema.ResourceData) ([]gofastly.DashboardItem, error) {
	var items []gofastly.DashboardItem
	var errs []error
	if v, ok := d.GetOk("dashboard_item"); ok {
		for i, r := range v.([]any) {
			if m, ok := r.(map[string]any); ok {
				item, err := mapToDashboardItem(m)
				if err != nil {
					errs = append(errs, fmt.Errorf("item #%d is invalid: %w", i, err))
				}
				items = append(items, *item)
			}
		}
	}
	if errs != nil {
		return nil, errors.Join(errs...)
	}
	return items, nil
}

func mapToDashboardItem(m map[string]any) (*gofastly.DashboardItem, error) {
	var errs []error
	var (
		id, title, subtitle string
		span                int

		dataSourceList, sourceConfigList []any
		dataSource, sourceConfig         map[string]any
		sourceType                       string
		metrics                          []string

		vizList, vizConfigList             []any
		visualization, visualizationConfig map[string]any
		visualizationType                  string
		plotType, format, calcMethod       string
	)

	var ok bool

	// Top-level fields
	if id, ok = m["id"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid id: %#v", m["id"]))
	}
	if title, ok = m["title"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid title: %#v", m["title"]))
	}
	if subtitle, ok = m["subtitle"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid subtitle: %#v", m["subtitle"]))
	}
	if span, ok = m["span"].(int); !ok {
		errs = append(errs, fmt.Errorf("invalid span: %#v", m["span"]))
	}
	if dataSourceList, ok = m["data_source"].([]any); !ok {
		errs = append(errs, fmt.Errorf("invalid data_source: %#v", m["data_source"]))
	}
	if len(dataSourceList) != 1 {
		errs = append(errs, fmt.Errorf("invalid data_source: %#v", m["data_source"]))
	} else {
		if dataSource, ok = dataSourceList[0].(map[string]any); !ok {
			errs = append(errs, fmt.Errorf("invalid data_source: %#v", m["data_source"]))
		}
	}

	if vizList, ok = m["visualization"].([]any); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization: %#v", m["visualization"]))
	}
	if len(vizList) != 1 {
		errs = append(errs, fmt.Errorf("invalid visualization: %#v", m["visualization"]))
	} else {
		if visualization, ok = vizList[0].(map[string]any); !ok {
			errs = append(errs, fmt.Errorf("invalid visualization: %#v", m["visualization"]))
		}
	}

	// Nested DataSource
	if sourceType, ok = dataSource["type"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid data_source.type: %#v", dataSource["type"]))
	}
	if sourceConfigList, ok = dataSource["config"].([]any); !ok {
		errs = append(errs, fmt.Errorf("invalid data_source.config: %#v", dataSource["config"]))
	}
	if len(sourceConfigList) != 1 {
		errs = append(errs, fmt.Errorf("invalid data_source.config: %#v", dataSource["config"]))
	} else {
		if sourceConfig, ok = sourceConfigList[0].(map[string]any); !ok {
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
	if vizConfigList, ok = visualization["config"].([]any); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization.config: %#v", visualization["config"]))
	}
	if len(vizConfigList) != 1 {
		errs = append(errs, fmt.Errorf("invalid visualization.config: %#v", visualization["config"]))
	} else {
		if visualizationConfig, ok = vizConfigList[0].(map[string]any); !ok {
			errs = append(errs, fmt.Errorf("invalid visualization.config: %#v", visualization["config"]))
		}
	}
	if plotType, ok = visualizationConfig["plot_type"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization.config.plot_type: %#v", visualizationConfig["plot_type"]))
	}
	if format, ok = visualizationConfig["format"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization.config.format: %#v", visualizationConfig["format"]))
	}
	if calcMethod, ok = visualizationConfig["calculation_method"].(string); !ok {
		errs = append(errs, fmt.Errorf("invalid visualization.config.calculation_method: %#v", visualizationConfig["calculation_method"]))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &gofastly.DashboardItem{
		ID: id,
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
				CalculationMethod: gofastly.ToPointer(gofastly.CalculationMethod(calcMethod)),
				Format:            gofastly.ToPointer(gofastly.VisualizationFormat(format)),
				PlotType:          gofastly.PlotType(plotType),
			},
			Type: gofastly.VisualizationType(visualizationType),
		},
	}, nil
}
