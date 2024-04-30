package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyIntegration_mailinglist(t *testing.T) {
	createIntegration := gofastly.Integration{
		Config: map[string]string{
			"address": fmt.Sprintf("noreply-%s@fastly.com", acctest.RandString(10)),
		},
		Description: gofastly.ToPointer("my description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("mailinglist"),
	}
	updateIntegration := gofastly.Integration{
		Config: map[string]string{
			"address": fmt.Sprintf("noreply-%s@fastly.com", acctest.RandString(10)),
		},
		Description: gofastly.ToPointer("my new description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("mailinglist"),
	}
	testAccFastlyIntegration(createIntegration, updateIntegration, t)
}

func TestAccFastlyIntegration_microsoftteams(t *testing.T) {
	createIntegration := gofastly.Integration{
		Config: map[string]string{
			"webhook": fmt.Sprintf("https://foo.com/bar-%s", acctest.RandString(10)),
		},
		Description: gofastly.ToPointer("my description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("microsoftteams"),
	}
	updateIntegration := gofastly.Integration{
		Config: map[string]string{
			"webhook": fmt.Sprintf("https://foo.com/bar-%s", acctest.RandString(10)),
		},
		Description: gofastly.ToPointer("my new description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("microsoftteams"),
	}
	testAccFastlyIntegration(createIntegration, updateIntegration, t)
}

func TestAccFastlyIntegration_newrelic(t *testing.T) {
	createIntegration := gofastly.Integration{
		Config: map[string]string{
			"key":     acctest.RandString(10),
			"account": acctest.RandString(10),
		},
		Description: gofastly.ToPointer("my description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("newrelic"),
	}
	updateIntegration := gofastly.Integration{
		Config: map[string]string{
			"key":     acctest.RandString(10),
			"account": acctest.RandString(10),
		},
		Description: gofastly.ToPointer("my new description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("newrelic"),
	}
	testAccFastlyIntegration(createIntegration, updateIntegration, t)
}

func TestAccFastlyIntegration_pagerduty(t *testing.T) {
	createIntegration := gofastly.Integration{
		Config: map[string]string{
			"key": acctest.RandString(10),
		},
		Description: gofastly.ToPointer("my description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("pagerduty"),
	}
	updateIntegration := gofastly.Integration{
		Config: map[string]string{
			"key": acctest.RandString(10),
		},
		Description: gofastly.ToPointer("my new description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("pagerduty"),
	}
	testAccFastlyIntegration(createIntegration, updateIntegration, t)
}

func TestAccFastlyIntegration_slack(t *testing.T) {
	createIntegration := gofastly.Integration{
		Config: map[string]string{
			"webhook": fmt.Sprintf("https://foo.com/bar-%s", acctest.RandString(10)),
		},
		Description: gofastly.ToPointer("my description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("slack"),
	}
	updateIntegration := gofastly.Integration{
		Config: map[string]string{
			"webhook": fmt.Sprintf("https://foo.com/bar-%s", acctest.RandString(10)),
		},
		Description: gofastly.ToPointer("my new description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("slack"),
	}
	testAccFastlyIntegration(createIntegration, updateIntegration, t)
}

func TestAccFastlyIntegration_webhook(t *testing.T) {
	createIntegration := gofastly.Integration{
		Config: map[string]string{
			"webhook": fmt.Sprintf("https://foo.com/bar-%s", acctest.RandString(10)),
		},
		Description: gofastly.ToPointer("my description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("webhook"),
	}
	updateIntegration := gofastly.Integration{
		Config: map[string]string{
			"webhook": fmt.Sprintf("https://foo.com/bar-%s", acctest.RandString(10)),
		},
		Description: gofastly.ToPointer("my new description"),
		Name:        gofastly.ToPointer(fmt.Sprintf("integration %s", acctest.RandString(10))),
		Type:        gofastly.ToPointer("webhook"),
	}
	testAccFastlyIntegration(createIntegration, updateIntegration, t)
}

func testAccFastlyIntegration(createIntegration, updateIntegration gofastly.Integration, t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig(createIntegration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyIntegrationsRemoteState(createIntegration),
				),
			},
			{
				Config: testAccIntegrationConfig(updateIntegration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyIntegrationsRemoteState(updateIntegration),
				),
			},
		},
	})
}

func testAccCheckFastlyIntegrationsRemoteState(expected gofastly.Integration) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn

		integrations := []gofastly.Integration{}

		var cursor *string

		for {
			sir, err := conn.SearchIntegrations(&gofastly.SearchIntegrationsInput{
				Cursor: cursor,
			})
			if err != nil {
				return fmt.Errorf("error searching integrations: %s", err)
			}

			integrations = append(integrations, sir.Data...)

			cursor = sir.Meta.NextCursor

			if cursor == nil {
				break
			}
		}

		var got *gofastly.Integration
		for _, integration := range integrations {
			if *integration.Name == *expected.Name {
				got = &integration
				break
			}
		}
		if got == nil {
			return fmt.Errorf("error looking up the integration")
		}
		if diff := cmp.Diff(expected.Config, got.Config); *expected.Type == "mailinglist" && diff != "" {
			return fmt.Errorf("bad config -expected +got\n%v", diff)
		}
		if *expected.Description != *got.Description {
			return fmt.Errorf("bad description, expected (%s), got (%s)", *expected.Description, *got.Description)
		}
		if *expected.Type != *got.Type {
			return fmt.Errorf("bad type, expected (%s), got (%s)", *expected.Type, *got.Type)
		}

		return nil
	}
}

func testAccCheckIntegrationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_integration" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		sir, err := conn.SearchIntegrations(&gofastly.SearchIntegrationsInput{})
		if err != nil {
			return fmt.Errorf("error searching integrations when checking integration destroy (%s): %s", rs.Primary.ID, err)
		}

		for _, i := range sir.Data {
			if *i.ID == rs.Primary.ID {
				// integration still found
				return fmt.Errorf("tried deleting integration (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccIntegrationConfig(integration gofastly.Integration) string {
	var config string
	for key, value := range integration.Config {
		config += fmt.Sprintf(`
    %s = "%s"`, key, value)
	}
	return fmt.Sprintf(`
resource "fastly_integration" "foo" {
  name = "%s"
  description = "%s"
  type = "%s"
  
  config = {%s
  }
}`, *integration.Name, *integration.Description, *integration.Type, config)
}
