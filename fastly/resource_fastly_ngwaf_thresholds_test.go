package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	ws "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/thresholds"
)

func TestAccFastlyNGWAFThresholds_validate(t *testing.T) {
	newWorkspaceName := fmt.Sprintf("Test Thresholds WS %s", acctest.RandString(10))
	thresholdName := fmt.Sprintf("threshold-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFWorkspaceConfig(newWorkspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", newWorkspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "description", "Test NGWAF Workspace"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "mode", "block"),
					testAccNGWAFWorkspaceExists("fastly_ngwaf_workspace.example"),
				),
			},
			{
				Config: testAccNGWAFThresholdsConfig(thresholdName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "action", "block"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "dont_notify", "false"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "duration", "86400"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "enabled", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "interval", "3600"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "limit", "10"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "name", thresholdName),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "signal", "SQLI"),
					resource.TestCheckResourceAttrPair("fastly_ngwaf_thresholds.sample", "workspace_id", "fastly_ngwaf_workspace.example", "id"),
					testAccNGWAFThresholdsExists("fastly_ngwaf_thresholds.sample"),
				),
			},
			{
				Config: testAccNGWAFThresholdsConfigUpdate(thresholdName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "action", "log"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "dont_notify", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "duration", "43200"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "enabled", "false"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "interval", "600"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "limit", "50"),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "name", thresholdName),
					resource.TestCheckResourceAttr("fastly_ngwaf_thresholds.sample", "signal", "BHH"),
					resource.TestCheckResourceAttrPair("fastly_ngwaf_thresholds.sample", "workspace_id", "fastly_ngwaf_workspace.example", "id"),
					testAccNGWAFThresholdsExists("fastly_ngwaf_thresholds.sample"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "fastly_ngwaf_thresholds.sample",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNGWAFThresholdsImportID("fastly_ngwaf_thresholds.sample"),
			},
		},
	})
}

func testAccNGWAFThresholdsImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["workspace_id"], rs.Primary.ID), nil
	}
}

func testAccNGWAFThresholdsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		threshold, err := ws.Get(context.TODO(), conn, &ws.GetInput{
			WorkspaceID: gofastly.ToPointer(rs.Primary.Attributes["workspace_id"]),
			ThresholdID: gofastly.ToPointer(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Thresholds %s: %v", rs.Primary.ID, err)
		}
		if threshold == nil {
			return fmt.Errorf("NGWAF Thresholds %s not found in API", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNGWAFThresholdsConfig(thresholdName string) string {
	return fmt.Sprintf(`
%s

resource "fastly_ngwaf_thresholds" "sample" {
    action       = "block"
    dont_notify  = false
    duration     = 86400
    enabled      = true
    interval     = 3600
    limit        = 10
    name         = "%s"
    signal       = "SQLI"
    workspace_id = fastly_ngwaf_workspace.example.id
  }
  `, testAccNGWAFWorkspaceConfig("Test Thresholds WS"), thresholdName)
}

func testAccNGWAFThresholdsConfigUpdate(thresholdName string) string {
	return fmt.Sprintf(`
%s

resource "fastly_ngwaf_thresholds" "sample" {
    action       = "log"
    dont_notify  = true
    duration     = 43200
    enabled      = false
    interval     = 600
    limit        = 50
    name         = "%s"
    signal       = "BHH"
    workspace_id = fastly_ngwaf_workspace.example.id
  }
  `, testAccNGWAFWorkspaceConfig("Test Thresholds WS"), thresholdName)
}
