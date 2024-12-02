package fastly

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/template"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyCustomDashboard_Basic(t *testing.T) {
	rand := acctest.RandString(10)
	dashboardName := fmt.Sprintf("Custom Dashboard %s", rand)
	dashboardDescription := fmt.Sprintf("Created by tf-test-%s", rand)
	dashboardItems := []gofastly.DashboardItem{
		{
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

	createDashboard := gofastly.CreateObservabilityCustomDashboardInput{
		Name:        dashboardName,
		Description: gofastly.ToPointer(dashboardDescription),
		Items:       dashboardItems,
	}

	updatedItems := []gofastly.DashboardItem{}
	updatedItems = append(updatedItems, dashboardItems[0])
	updatedItems[0].Subtitle = "This is STILL the first chart"
	updatedItems = append(updatedItems, gofastly.DashboardItem{
		Title:    "NEW Chart",
		Subtitle: "This is the new Chart #2",
		DataSource: gofastly.DashboardDataSource{
			Type:   gofastly.SourceTypeStatsOrigin,
			Config: gofastly.DashboardSourceConfig{Metrics: []string{"all_status_2xx"}},
		},
		Visualization: gofastly.DashboardVisualization{
			Type:   gofastly.VisualizationTypeChart,
			Config: gofastly.VisualizationConfig{PlotType: gofastly.PlotTypeSingleMetric},
		},
	})
	updatedItems = append(updatedItems, dashboardItems[1])
	updatedItems[2].Title = "Chart #3"
	updatedItems[2].Subtitle = "This is chart, the third now"
	updatedItems[2].Visualization.Config.PlotType = gofastly.PlotTypeDonut
	updatedItems[2].Visualization.Config.CalculationMethod = nil

	updatedName := "This is an updated dashboard"
	updatedDescription := fmt.Sprintf("Updated by tf-test-%s", rand)
	updateDashboard := gofastly.UpdateObservabilityCustomDashboardInput{
		Description: &updatedDescription,
		Items:       &updatedItems,
		Name:        &updatedName,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCustomDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityCustomDashboard(t, createDashboard),
				Check:  testAccCustomDashboardRemoteState(dashboardName, dashboardDescription, dashboardItems),
			},
			{
				Config: testAccObservabilityCustomDashboard(t, updateDashboard),
				Check:  testAccCustomDashboardRemoteState(updatedName, updatedDescription, updatedItems),
			},
			{
				ResourceName:      "fastly_custom_dashboard.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

}

func testAccCustomDashboardRemoteState(dashboardName, dashboardDescription string, dashboardItems []gofastly.DashboardItem) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn

		dashboards, err := conn.ListObservabilityCustomDashboards(&gofastly.ListObservabilityCustomDashboardsInput{})
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

		if got.Name != dashboardName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", dashboardName, got.Name)
		} else if got.Description != dashboardDescription {
			return fmt.Errorf("bad description, expected (%s), got (%s)", dashboardDescription, got.Description)
		} else if len(got.Items) != len(dashboardItems) {
			return fmt.Errorf("bad items, expected (%d items), got (%d items), %#v", len(dashboardItems), len(got.Items), got)
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
				id = ".ID"
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
		dashResp, err := conn.ListObservabilityCustomDashboards(&gofastly.ListObservabilityCustomDashboardsInput{})
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
