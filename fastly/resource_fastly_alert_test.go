package fastly

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

func TestAccFastlyAlert_Basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	createAlert := gofastly.AlertDefinition{
		Dimensions: map[string][]string{
			"domains": {"example.com", "fastly.com"},
		},
		EvaluationStrategy: map[string]any{
			"type":      "above_threshold",
			"period":    "5m",
			"threshold": float64(10),
		},
		Metric: "status_5xx",
		Name:   fmt.Sprintf("alert %s", acctest.RandString(10)),
		Source: "domains",
	}
	updateAlert := gofastly.AlertDefinition{
		Description: "my new description",
		Dimensions: map[string][]string{
			"domains": {"demo.com", "fastly.com"},
		},
		EvaluationStrategy: map[string]any{
			"type":      "below_threshold",
			"period":    "15m",
			"threshold": float64(100),
		},
		Metric: "status_4xx",
		Name:   fmt.Sprintf("alert %s", acctest.RandString(10)),
		Source: "domains",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAlertDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertConfig(serviceName, domainName, createAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.example", &service),
					testAccCheckFastlyAlertsRemoteState(&service, serviceName, createAlert),
				),
			},
			{
				Config: testAccAlertConfig(serviceName, domainName, updateAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.example", &service),
					testAccCheckFastlyAlertsRemoteState(&service, serviceName, updateAlert),
				),
			},
			{
				ResourceName:      "fastly_alert.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyAlert_BasicStats(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	createAlert := gofastly.AlertDefinition{
		Description: "Terraform test",
		Dimensions:  map[string][]string{},
		EvaluationStrategy: map[string]any{
			"type":      "above_threshold",
			"period":    "5m",
			"threshold": float64(10),
		},
		Metric: "status_5xx",
		Name:   fmt.Sprintf("Terraform test alert %s", acctest.RandString(10)),
		Source: "stats",
	}
	updateAlert := gofastly.AlertDefinition{
		Description: "Terraform test with new description",
		Dimensions:  map[string][]string{},
		EvaluationStrategy: map[string]any{
			"type":      "below_threshold",
			"period":    "15m",
			"threshold": float64(100),
		},
		Metric: "status_4xx",
		Name:   fmt.Sprintf("Terraform test alert %s", acctest.RandString(10)),
		Source: "stats",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAlertDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertStatsConfig(serviceName, domainName, createAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.tf_bar", &service),
					testAccCheckFastlyAlertsRemoteState(&service, serviceName, createAlert),
				),
			},
			{
				Config: testAccAlertStatsConfig(serviceName, domainName, updateAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.tf_bar", &service),
					testAccCheckFastlyAlertsRemoteState(&service, serviceName, updateAlert),
				),
			},
			{
				ResourceName:      "fastly_alert.tf_bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyAlert_BasicStatsAggregate(t *testing.T) {
	service := gofastly.ServiceDetail{
		Name:      gofastly.ToPointer(""),
		ServiceID: gofastly.ToPointer(""),
	}

	createAlert := gofastly.AlertDefinition{
		Description: "Terraform test",
		Dimensions:  map[string][]string{},
		EvaluationStrategy: map[string]any{
			"type":      "above_threshold",
			"period":    "5m",
			"threshold": float64(10),
		},
		Metric: "status_5xx",
		Name:   fmt.Sprintf("Terraform test alert %s", acctest.RandString(10)),
		Source: "stats",
	}
	updateAlert := gofastly.AlertDefinition{
		Description: "Terraform test with new description",
		Dimensions:  map[string][]string{},
		EvaluationStrategy: map[string]any{
			"type":      "below_threshold",
			"period":    "15m",
			"threshold": float64(100),
		},
		Metric: "status_4xx",
		Name:   fmt.Sprintf("Terraform test alert %s", acctest.RandString(10)),
		Source: "stats",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAlertDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertAggregateStatsConfig(createAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyAlertsRemoteState(&service, "", createAlert),
				),
			},
			{
				Config: testAccAlertAggregateStatsConfig(updateAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyAlertsRemoteState(&service, "", updateAlert),
				),
			},
			{
				ResourceName:      "fastly_alert.tf_bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyAlert_BasicStatsAggregatePercent(t *testing.T) {
	service := gofastly.ServiceDetail{
		Name:      gofastly.ToPointer(""),
		ServiceID: gofastly.ToPointer(""),
	}

	createAlert := gofastly.AlertDefinition{
		Description: "Terraform percent test",
		Dimensions:  map[string][]string{},
		// 25 percent increase
		EvaluationStrategy: map[string]any{
			"type":         "percent_increase",
			"period":       "2m",
			"threshold":    0.25,
			"ignore_below": float64(10),
		},
		Metric: "status_4xx",
		Name:   fmt.Sprintf("Terraform test percent alert %s", acctest.RandString(10)),
		Source: "stats",
	}
	updateAlert := gofastly.AlertDefinition{
		Description: "Terraform test with new description",
		Dimensions:  map[string][]string{},
		// 10 percent increase
		EvaluationStrategy: map[string]any{
			"type":         "percent_increase",
			"period":       "2m",
			"threshold":    0.1,
			"ignore_below": float64(10),
		},
		Metric: "status_4xx",
		Name:   fmt.Sprintf("Terraform test update percent alert %s", acctest.RandString(10)),
		Source: "stats",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAlertDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertPercentAggregateStatsConfig(createAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyAlertsRemoteState(&service, "", createAlert),
				),
			},
			{
				Config: testAccAlertPercentAggregateStatsConfig(updateAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyAlertsRemoteState(&service, "", updateAlert),
				),
			},
			{
				ResourceName:      "fastly_alert.tf_percent",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyAlert_BasicStatsBadConfig(t *testing.T) {
	createAlert := gofastly.AlertDefinition{
		Description: "Terraform percent test",
		Dimensions:  map[string][]string{},
		// 25 percent increase
		EvaluationStrategy: map[string]any{
			"type":         "percent_increase",
			"period":       "2m",
			"threshold":    0.25,
			"ignore_below": float64(10),
		},
		Metric: "status_4xx",
		Name:   fmt.Sprintf("Terraform test percent alert %s", acctest.RandString(10)),
		Source: "origins",
	}
	updateAlert := gofastly.AlertDefinition{
		Description: "Terraform test with new description",
		Dimensions:  map[string][]string{},
		// 10 percent increase
		EvaluationStrategy: map[string]any{
			"type":         "percent_increase",
			"period":       "2m",
			"threshold":    0.1,
			"ignore_below": float64(10),
		},
		Metric: "status_4xx",
		Name:   fmt.Sprintf("Terraform test update percent alert %s", acctest.RandString(10)),
		Source: "domains",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAlertDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAlertPercentAggregateStatsConfig(createAlert),
				ExpectError: regexp.MustCompile(badAlertSourceServiceIDConfig),
			},
			{
				Config:      testAccAlertPercentAggregateStatsConfig(updateAlert),
				ExpectError: regexp.MustCompile(badAlertSourceServiceIDConfig),
			},
		},
	})
}

func testAccCheckFastlyAlertsRemoteState(service *gofastly.ServiceDetail, serviceName string, expected gofastly.AlertDefinition) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != serviceName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", serviceName, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		alerts := []gofastly.AlertDefinition{}

		var cursor string

		for {
			adr, err := conn.ListAlertDefinitions(context.TODO(), &gofastly.ListAlertDefinitionsInput{
				Cursor: gofastly.ToPointer(cursor),
			})
			if err != nil {
				return fmt.Errorf("error listing all alert definitions: %s", err)
			}

			alerts = append(alerts, adr.Data...)

			cursor = adr.Meta.NextCursor

			if cursor == "" {
				break
			}
		}

		var got *gofastly.AlertDefinition
		for _, alert := range alerts {
			if alert.Name == expected.Name {
				got = &alert
				break
			}
		}
		if got == nil {
			return fmt.Errorf("error looking up the alert")
		}
		expectedDescription := strings.TrimSpace(expected.Description + " " + ManagedByTerraform)
		if expectedDescription != got.Description {
			return fmt.Errorf("bad description, expected (%s), got (%s)", expectedDescription, got.Description)
		}
		if diff := cmp.Diff(expected.Dimensions, got.Dimensions); diff != "" {
			return fmt.Errorf("bad dimensions -expected +got\n%v", diff)
		}
		if diff := cmp.Diff(expected.EvaluationStrategy, got.EvaluationStrategy); diff != "" {
			return fmt.Errorf("bad evaluation_strategy -expected +got\n%v", diff)
		}
		if expected.Metric != got.Metric {
			return fmt.Errorf("bad metric, expected (%s), got (%s)", expected.Metric, got.Metric)
		}
		if gofastly.ToValue(service.ServiceID) != got.ServiceID {
			return fmt.Errorf("bad service_id, expected (%s), got (%s)", gofastly.ToValue(service.ServiceID), got.ServiceID)
		}
		if expected.Source != got.Source {
			return fmt.Errorf("bad source, expected (%s), got (%s)", expected.Source, got.Source)
		}

		return nil
	}
}

func testAccCheckAlertDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_alert" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		adr, err := conn.ListAlertDefinitions(context.TODO(), &gofastly.ListAlertDefinitionsInput{})
		if err != nil {
			return fmt.Errorf("error listing alert definitions when checking alert destroy (%s): %s", rs.Primary.ID, err)
		}

		for _, ad := range adr.Data {
			if ad.ID == rs.Primary.ID {
				// alert definition still found
				return fmt.Errorf("tried deleting alert (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAlertConfig(serviceName, domainName string, alert gofastly.AlertDefinition) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "example" {
  name = "%s"

  domain {
    name = "%s"
  }

  product_enablement {
    domain_inspector = true
  }

  force_destroy = true
}

resource "fastly_alert" "foo" {
  name = "%s"
  description = "%s"
  service_id = fastly_service_vcl.example.id
  source = "%s"
  metric = "%s"

  dimensions {
    %s = ["%s"]
  }
  
  evaluation_strategy {
    type = "%s"
    period = "%s"
    threshold = %v
  }
}`, serviceName, domainName, alert.Name, alert.Description, alert.Source, alert.Metric, alert.Source, strings.Join(alert.Dimensions[alert.Source], "\", \""), alert.EvaluationStrategy["type"], alert.EvaluationStrategy["period"], alert.EvaluationStrategy["threshold"])
}

func testAccAlertStatsConfig(serviceName, domainName string, alert gofastly.AlertDefinition) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "tf_bar" {
  name = "%s"

  domain {
    name = "%s"
  }

  product_enablement {
    domain_inspector = false
  }

  force_destroy = true
}

resource "fastly_alert" "tf_bar" {
  name = "%s"
  description = "%s"
  service_id = fastly_service_vcl.tf_bar.id
  source = "%s"
  metric = "%s"

  evaluation_strategy {
    type = "%s"
    period = "%s"
    threshold = %v
  }
}`, serviceName, domainName, alert.Name, alert.Description, alert.Source, alert.Metric, alert.EvaluationStrategy["type"], alert.EvaluationStrategy["period"], alert.EvaluationStrategy["threshold"])
}

func testAccAlertAggregateStatsConfig(alert gofastly.AlertDefinition) string {
	return fmt.Sprintf(`
resource "fastly_alert" "tf_bar" {
  name = "%s"
  description = "%s"
  service_id = ""
  source = "%s"
  metric = "%s"

  evaluation_strategy {
    type = "%s"
    period = "%s"
    threshold = %v
  }
}`, alert.Name, alert.Description, alert.Source, alert.Metric, alert.EvaluationStrategy["type"], alert.EvaluationStrategy["period"], alert.EvaluationStrategy["threshold"])
}

func testAccAlertPercentAggregateStatsConfig(alert gofastly.AlertDefinition) string {
	return fmt.Sprintf(`
resource "fastly_alert" "tf_percent" {
  name = "%s"
  description = "%s"
  source = "%s"
  metric = "%s"

  evaluation_strategy {
    type = "%s"
    period = "%s"
    threshold = %v
    ignore_below = %v
  }
}`, alert.Name, alert.Description, alert.Source, alert.Metric, alert.EvaluationStrategy["type"], alert.EvaluationStrategy["period"], alert.EvaluationStrategy["threshold"], alert.EvaluationStrategy["ignore_below"])
}
