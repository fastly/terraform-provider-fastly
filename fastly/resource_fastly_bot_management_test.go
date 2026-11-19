package fastly

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/fastly/go-fastly/v19/fastly/botmanagement/v1/workspaces/policy"
)

// testBotManagementBotID must belong to testBotManagementCategoryID.
const (
	testBotManagementCategoryID = "cat_ai_crawler"
	testBotManagementBotID      = "bot_gptbot"
)

func TestParseBotManagementImportID(t *testing.T) {
	cases := []struct {
		id      string
		names   []string
		want    []string
		wantErr bool
	}{
		{id: "ws/cat", names: []string{"workspace_id", "category_id"}, want: []string{"ws", "cat"}},
		{id: "ws/cat/bot", names: []string{"workspace_id", "category_id", "bot_id"}, want: []string{"ws", "cat", "bot"}},
		{id: "ws", names: []string{"workspace_id", "category_id"}, wantErr: true},
		{id: "ws/cat/bot", names: []string{"workspace_id", "category_id"}, wantErr: true},
		{id: "ws/", names: []string{"workspace_id", "category_id"}, wantErr: true},
		{id: "/cat/bot", names: []string{"workspace_id", "category_id", "bot_id"}, wantErr: true},
	}

	for _, tc := range cases {
		got, err := parseBotManagementImportID(tc.id, tc.names...)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseBotManagementImportID(%q) = %v, want error", tc.id, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseBotManagementImportID(%q) unexpected error: %v", tc.id, err)
			continue
		}
		if fmt.Sprint(got) != fmt.Sprint(tc.want) {
			t.Errorf("parseBotManagementImportID(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestResourceFastlyBotManagementBot_actionValidation(t *testing.T) {
	validate := resourceFastlyBotManagementBot().Schema["action"].ValidateDiagFunc
	for _, v := range []string{"allow", "block", "challenge", "inherit"} {
		if diags := validate(v, cty.GetAttrPath("action")); diags.HasError() {
			t.Errorf("action %q: unexpected error: %v", v, diags)
		}
	}
	if diags := validate("log", cty.GetAttrPath("action")); !diags.HasError() {
		t.Errorf("action %q: expected error, got none", "log")
	}
}

// TestAccFastlyBotManagement_lifecycle enables ContentGuard in its own step
// before looking up the workspace, since the workspace must exist before the
// data source reads it.
func TestAccFastlyBotManagement_lifecycle(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-tf-%s", acctest.RandString(10))
	workspaceName := fmt.Sprintf("tf-test-bot-management-%s", acctest.RandString(10))

	const (
		workspace = "fastly_bot_management_workspace.test"
		category  = "fastly_bot_management_category.test"
		bot       = "fastly_bot_management_bot.test"
		byService = "data.fastly_bot_management_workspace.by_service"
		byID      = "data.fastly_bot_management_workspace.by_id"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBotManagementServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBotManagementServiceConfig(serviceName, domainName, backendName),
			},
			{
				Config: testAccBotManagementConfig(serviceName, domainName, backendName, workspaceName, "log", "block", "allow"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(byService, "id"),
					resource.TestCheckResourceAttrPair(byID, "id", byService, "id"),
					resource.TestCheckTypeSetElemAttrPair(byService, "services.*", "fastly_service_vcl.test", "id"),

					resource.TestCheckResourceAttrPair(workspace, "workspace_id", byService, "id"),
					resource.TestCheckResourceAttrPair(workspace, "id", workspace, "workspace_id"),
					resource.TestCheckResourceAttr(workspace, "name", workspaceName),
					resource.TestCheckResourceAttr(workspace, "description", "Managed by Terraform acceptance tests"),
					resource.TestCheckResourceAttr(workspace, "protection_mode", "log"),

					resource.TestCheckResourceAttr(category, "category_id", testBotManagementCategoryID),
					resource.TestCheckResourceAttr(category, "action", "block"),
					resource.TestCheckResourceAttrSet(category, "name"),

					resource.TestCheckResourceAttr(bot, "bot_id", testBotManagementBotID),
					resource.TestCheckResourceAttr(bot, "category_id", testBotManagementCategoryID),
					resource.TestCheckResourceAttr(bot, "action", "allow"),
					resource.TestCheckResourceAttrSet(bot, "name"),

					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_bot_management_categories.test", "categories.*", map[string]string{
						"category_id": testBotManagementCategoryID,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_bot_management_bots.test", "bots.*", map[string]string{
						"bot_id":      testBotManagementBotID,
						"category_id": testBotManagementCategoryID,
					}),
				),
			},
			{
				Config: testAccBotManagementConfig(serviceName, domainName, backendName, workspaceName, "log", "block", "inherit"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(bot, "action", "inherit"),
					testAccCheckBotManagementBotAction(category, testBotManagementBotID, "inherit"),
				),
			},
			{
				Config: testAccBotManagementConfig(serviceName, domainName, backendName, workspaceName, "block", "challenge", "block"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(workspace, "protection_mode", "block"),
					resource.TestCheckResourceAttr(category, "action", "challenge"),
					resource.TestCheckResourceAttr(bot, "action", "block"),
				),
			},
			{
				ResourceName:      workspace,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      category,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      bot,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Removing the bot resource leaves the bot's action unchanged.
				Config: testAccBotManagementConfig(serviceName, domainName, backendName, workspaceName, "block", "challenge", ""),
				Check:  testAccCheckBotManagementBotAction(category, testBotManagementBotID, "block"),
			},
		},
	})
}

// testAccCheckBotManagementServiceDestroy retries the service destroy check,
// since the service can briefly remain in the service list after deletion.
func testAccCheckBotManagementServiceDestroy(s *terraform.State) error {
	return retry.RetryContext(context.TODO(), 2*time.Minute, func() *retry.RetryError {
		if err := testAccCheckServiceVCLDestroy(s); err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})
}

// testAccCheckBotManagementBotAction uses go-fastly to check a bot's action,
// including after the bot resource has been removed from state.
func testAccCheckBotManagementBotAction(categoryResource, botID, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[categoryResource]
		if !ok {
			return fmt.Errorf("not found: %s", categoryResource)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		got, err := policy.GetBot(context.TODO(), conn, &policy.GetBotInput{
			WorkspaceID: new(rs.Primary.Attributes["workspace_id"]),
			CategoryID:  new(rs.Primary.Attributes["category_id"]),
			BotID:       new(botID),
		})
		if err != nil {
			return fmt.Errorf("error reading bot %s: %w", botID, err)
		}
		if got.Action.Type != want {
			return fmt.Errorf("bot %s action = %q, want %q", botID, got.Action.Type, want)
		}
		return nil
	}
}

func testAccBotManagementServiceConfig(serviceName, domainName, backendName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "test" {
  name = "%s"

  domain {
    name = "%s"
  }

  backend {
    address = "httpbin.org"
    name    = "%s"
    port    = 443
  }

  product_enablement {
    bot_management {
      enabled      = true
      contentguard = "on"
    }
  }

  force_destroy = true
}
`, serviceName, domainName, backendName)
}

// testAccBotManagementConfig omits the bot resource when botAction is empty.
func testAccBotManagementConfig(serviceName, domainName, backendName, workspaceName, protectionMode, categoryAction, botAction string) string {
	config := testAccBotManagementServiceConfig(serviceName, domainName, backendName) + fmt.Sprintf(`
data "fastly_bot_management_workspace" "by_service" {
  service_id = fastly_service_vcl.test.id
}

data "fastly_bot_management_workspace" "by_id" {
  workspace_id = data.fastly_bot_management_workspace.by_service.id
}

resource "fastly_bot_management_workspace" "test" {
  workspace_id    = data.fastly_bot_management_workspace.by_service.id
  name            = "%s"
  description     = "Managed by Terraform acceptance tests"
  protection_mode = "%s"
}

resource "fastly_bot_management_category" "test" {
  workspace_id = data.fastly_bot_management_workspace.by_service.id
  category_id  = "%s"
  action       = "%s"
}

data "fastly_bot_management_categories" "test" {
  workspace_id = data.fastly_bot_management_workspace.by_service.id
}

data "fastly_bot_management_bots" "test" {
  workspace_id = data.fastly_bot_management_workspace.by_service.id
  category_id  = "%s"
}
`, workspaceName, protectionMode, testBotManagementCategoryID, categoryAction, testBotManagementCategoryID)

	if botAction != "" {
		config += fmt.Sprintf(`
resource "fastly_bot_management_bot" "test" {
  workspace_id = data.fastly_bot_management_workspace.by_service.id
  category_id  = fastly_bot_management_category.test.category_id
  bot_id       = "%s"
  action       = "%s"
}
`, testBotManagementBotID, botAction)
	}

	return config
}
