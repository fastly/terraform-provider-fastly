package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	ws "github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/workspaces/virtualpatches"
)

func TestAccFastlyNGWAFVirtualPatch_validate(t *testing.T) {
	newWorkspaceName := fmt.Sprintf("Test Virtual Patch WS %s", acctest.RandString(10))
	virtualPatchID := "CVE-2017-5638"

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
				Config: testAccNGWAFVirtualPatchConfig(virtualPatchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "action", "block"),
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "enabled", "true"),
					resource.TestCheckResourceAttrPair("fastly_ngwaf_virtual_patches.sample", "workspace_id", "fastly_ngwaf_workspace.example", "id"),
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "virtual_patch_id", virtualPatchID),
					testAccNGWAFVirtualPatchExists("fastly_ngwaf_virtual_patches.sample"),
				),
			},
			{
				Config: testAccNGWAFVirtualPatchConfigUpdate(virtualPatchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "action", "log"),
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "enabled", "false"),
					resource.TestCheckResourceAttrPair("fastly_ngwaf_virtual_patches.sample", "workspace_id", "fastly_ngwaf_workspace.example", "id"),
					resource.TestCheckResourceAttr("fastly_ngwaf_virtual_patches.sample", "virtual_patch_id", virtualPatchID),
					testAccNGWAFVirtualPatchExists("fastly_ngwaf_virtual_patches.sample"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "fastly_ngwaf_virtual_patches.sample",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNGWAFVirtualPatchImportID("fastly_ngwaf_virtual_patches.sample"),
			},
		},
	})
}

func testAccNGWAFVirtualPatchImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["workspace_id"], rs.Primary.Attributes["virtual_patch_id"]), nil
	}
}

func testAccNGWAFVirtualPatchExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		virtualpatch, err := ws.Get(context.TODO(), conn, &ws.GetInput{
			WorkspaceID:    gofastly.ToPointer(rs.Primary.Attributes["workspace_id"]),
			VirtualPatchID: gofastly.ToPointer(rs.Primary.Attributes["virtual_patch_id"]),
		})
		if err != nil {
			return fmt.Errorf("Unable to retrieve NGWAF Virtual Patch %s: %v", rs.Primary.ID, err)
		}
		if virtualpatch == nil {
			return fmt.Errorf("NGWAF Virtual Patch %s not found in API", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNGWAFVirtualPatchConfig(virtualPatchID string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
    name                         = "Test Virtual Patch WS"
    description                  = "Test NGWAF Workspace"
    mode                         = "log"
  }

  resource "fastly_ngwaf_virtual_patches" "sample" {
    action            = "block"
    enabled           = true
    virtual_patch_id  = "%s"
    workspace_id      = fastly_ngwaf_workspace.example.id
  }
  `, virtualPatchID)
}

func testAccNGWAFVirtualPatchConfigUpdate(virtualPatchID string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
    name                         = "Test Virtual Patch WS"
    description                  = "Test NGWAF Workspace"
    mode                         = "log"
  }

  resource "fastly_ngwaf_virtual_patches" "sample" {
    action            = "log"
    enabled           = false
    virtual_patch_id  = "%s"
    workspace_id      = fastly_ngwaf_workspace.example.id
  }
  `, virtualPatchID)
}
