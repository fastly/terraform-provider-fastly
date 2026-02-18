package fastly

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
)

func generateDashboardParams(t *testing.T) (name, description string, items []gofastly.DashboardItem) {
	t.Helper()

	rand := acctest.RandString(10)
	name = fmt.Sprintf("Custom Dashboard %s", rand)
	description = fmt.Sprintf("Created by tf-test-%s", rand)
	items = []gofastly.DashboardItem{
		{
			ID: "item1",
			DataSource: gofastly.DashboardDataSource{
				Config: gofastly.DashboardSourceConfig{
					Metrics: []string{"requests"},
				},
				Type: gofastly.SourceTypeStatsEdge,
			},
			Span:     4,
			Subtitle: "This is the first chart",
			Title:    "Chart #1",
			Visualization: gofastly.DashboardVisualization{
				Config: gofastly.VisualizationConfig{
					PlotType: gofastly.PlotTypeBar,
				},
				Type: gofastly.VisualizationTypeChart,
			},
		},
		{
			ID: "item2",
			DataSource: gofastly.DashboardDataSource{
				Config: gofastly.DashboardSourceConfig{
					Metrics: []string{"status_4xx", "status_5xx"},
				},
				Type: gofastly.SourceTypeStatsDomain,
			},
			Span:     12,
			Subtitle: "This is chart, the second",
			Title:    "Chart #2",
			Visualization: gofastly.DashboardVisualization{
				Config: gofastly.VisualizationConfig{
					PlotType:          gofastly.PlotTypeLine,
					CalculationMethod: gofastly.ToPointer(gofastly.CalculationMethodAvg),
				},
				Type: gofastly.VisualizationTypeChart,
			},
		},
	}

	return
}

func TestAccFastlyCustomDashboard_Basic(t *testing.T) {
	dashboardName, dashboardDescription, dashboardItems := generateDashboardParams(t)

	createDashboard := gofastly.CreateObservabilityCustomDashboardInput{
		Name:        dashboardName,
		Description: gofastly.ToPointer(dashboardDescription),
		Items:       dashboardItems,
	}

	// Leave one item alone
	updatedItems := []gofastly.DashboardItem{}
	updatedItems = append(updatedItems, dashboardItems[0])

	// Update one item in place
	updatedItems = append(updatedItems, dashboardItems[1])
	updatedItems[1].Visualization.Config.PlotType = gofastly.PlotTypeDonut
	updatedItems[1].Visualization.Config.CalculationMethod = nil

	// Add a new item
	updatedItems = append(updatedItems, gofastly.DashboardItem{
		ID:       "item3",
		Title:    "NEW Chart",
		Subtitle: "This is the new Chart #3",
		DataSource: gofastly.DashboardDataSource{
			Type:   gofastly.SourceTypeStatsOrigin,
			Config: gofastly.DashboardSourceConfig{Metrics: []string{"all_status_2xx"}},
		},
		Visualization: gofastly.DashboardVisualization{
			Type:   gofastly.VisualizationTypeChart,
			Config: gofastly.VisualizationConfig{PlotType: gofastly.PlotTypeSingleMetric},
		},
	})

	updatedName := "This is an updated dashboard"
	update1 := gofastly.UpdateObservabilityCustomDashboardInput{
		Description: &dashboardDescription,
		Items:       &updatedItems,
		Name:        &updatedName,
	}

	// Reverse `Items` to test item IDs are honored
	update2 := update1
	update2.Items = &[]gofastly.DashboardItem{updatedItems[2], updatedItems[1], updatedItems[0]}

	// Update again, deleting first and last item
	update3 := update1
	update3.Items = &[]gofastly.DashboardItem{updatedItems[1]}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCustomDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityCustomDashboard(t, createDashboard),
				Check: resource.ComposeTestCheckFunc(
					testAccCustomDashboardRemoteState(dashboardName),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "name", dashboardName),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "description", dashboardDescription),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.#", "2"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.id", "item1"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.title", "Chart #1"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.id", "item2"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.title", "Chart #2"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.visualization.0.config.0.plot_type", "line"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.visualization.0.config.0.calculation_method", "avg"),
				),
			},
			{
				Config: testAccObservabilityCustomDashboard(t, update1),
				Check: resource.ComposeTestCheckFunc(
					testAccCustomDashboardRemoteState(updatedName),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "name", updatedName),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "description", dashboardDescription),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.#", "3"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.id", "item1"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.title", "Chart #1"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.id", "item2"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.title", "Chart #2"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.visualization.0.config.0.plot_type", "donut"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.visualization.0.config.0.calculation_method", ""),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.2.id", "item3"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.2.title", "NEW Chart"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.2.span", "4"),
				),
			},
			{
				Config: testAccObservabilityCustomDashboard(t, update2),
				Check: resource.ComposeTestCheckFunc(
					testAccCustomDashboardRemoteState(updatedName),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "name", updatedName),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "description", dashboardDescription),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.#", "3"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.id", "item3"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.title", "NEW Chart"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.span", "4"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.id", "item2"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.title", "Chart #2"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.visualization.0.config.0.plot_type", "donut"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.1.visualization.0.config.0.calculation_method", ""),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.2.id", "item1"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.2.title", "Chart #1"),
				),
			},
			{
				Config: testAccObservabilityCustomDashboard(t, update3),
				Check: resource.ComposeTestCheckFunc(
					testAccCustomDashboardRemoteState(updatedName),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "name", updatedName),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "description", dashboardDescription),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.#", "1"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.id", "item2"),
					resource.TestCheckResourceAttr("fastly_custom_dashboard.example", "dashboard_item.0.title", "Chart #2"),
				),
			},
			{
				ResourceName:      "fastly_custom_dashboard.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCustomDashboardRemoteState(dashboardName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn

		dashboards, err := conn.ListObservabilityCustomDashboards(context.TODO(), &gofastly.ListObservabilityCustomDashboardsInput{})
		if err != nil {
			return fmt.Errorf("error listing all Custom Dashboards: %s", err)
		}

		var got *gofastly.ObservabilityCustomDashboard
		var found bool
		for _, dash := range dashboards.Data {
			if dash.Name == dashboardName {
				found = true
				got = &dash
				break
			}
		}
		if !found || got == nil {
			return fmt.Errorf("error looking up the dashboard")
		}

		return nil
	}
}

func testAccObservabilityCustomDashboard(t *testing.T, input any) string {
	t.Helper()
	f := template.FuncMap{"join": strings.Join, "quote": func(s []string) []string {
		final := make([]string, len(s))
		for i, q := range s {
			final[i] = fmt.Sprintf("%q", q)
		}
		return final
	}}
	tmpl := template.Must(template.New("dashboard_items").Funcs(f).Parse(`
	resource "fastly_custom_dashboard" "example" {
		name = "{{ .Name }}"
		description = "{{ .Description }}"

		{{ range .Items -}}
		dashboard_item {
			{{if .ID -}}
				id = "{{- .ID -}}"
			{{- end}}
			title = "{{- .Title -}}"
			subtitle = "{{- .Subtitle -}}"
			{{if .Span -}}
				span = "{{- .Span -}}"
			{{- end}}
			data_source {
				type = "{{- .DataSource.Type -}}"
				config {
					metrics = [{{- join (quote .DataSource.Config.Metrics) "," -}}]
				}
			}
			visualization {
				type = "chart"
				config {
					plot_type = "{{- .Visualization.Config.PlotType -}}"
					{{if .Visualization.Config.CalculationMethod -}}
						calculation_method = "{{- .Visualization.Config.CalculationMethod -}}"
					{{- end}}
					{{if .Visualization.Config.Format -}}
						format = "{{- .Visualization.Config.Format -}}"
					{{- end}}
				}
			}
		}
		{{ end }}
	}
	`))
	b := new(bytes.Buffer)
	if err := tmpl.Execute(b, input); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func testAccCheckCustomDashboardDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_custom_dashboard" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		dashResp, err := conn.ListObservabilityCustomDashboards(context.TODO(), &gofastly.ListObservabilityCustomDashboardsInput{})
		if err != nil {
			return fmt.Errorf("error listing custom dashboards when checking dashboard destroy (%s): %s", rs.Primary, err)
		}

		for _, dash := range dashResp.Data {
			if dash.ID == rs.Primary.ID {
				return fmt.Errorf("tried deleting dashboard (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}
