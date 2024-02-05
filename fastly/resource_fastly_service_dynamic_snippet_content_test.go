package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceDynamicSnippetContent_create(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	expectedNumberOfSnippets := "1"
	expectedSnippetName := "dyn_hit_test"
	expectedSnippetType := "hit"
	expectedSnippetPriority := "100"
	expectedSnippetContent := "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDynamicSnippetContentConfigWithDynamicSnippet(serviceName, expectedSnippetName, expectedSnippetContent, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceDynamicSnippetContentRemoteState(&service, serviceName, expectedSnippetName, expectedSnippetContent),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", expectedNumberOfSnippets),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "dynamicsnippet.*", map[string]string{
						"name":     expectedSnippetName,
						"type":     expectedSnippetType,
						"priority": expectedSnippetPriority,
					}),
					resource.TestCheckResourceAttr("fastly_service_dynamic_snippet_content.content", "content", expectedSnippetContent),
				),
			},
			{
				ResourceName:            "fastly_service_dynamic_snippet_content.content",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manage_snippets"},
			},
		},
	})
}

func TestAccFastlyServiceDynamicSnippetContent_update(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	dynamicSnippetName := fmt.Sprintf("dynamic snippet %s", acctest.RandString(10))

	expectedRemoteItems := "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

	expectedRemoteItemsAfterUpdate := "if ( req.url ) {\n set req.http.my-updated-test-header = \"true\";\n}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDynamicSnippetContentConfigWithDynamicSnippet(name, dynamicSnippetName, expectedRemoteItems, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceDynamicSnippetContentRemoteState(&service, name, dynamicSnippetName, expectedRemoteItems),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_dynamic_snippet_content.content", "content", expectedRemoteItems),
				),
			},
			{
				Config: testAccServiceDynamicSnippetContentConfigWithDynamicSnippet(name, dynamicSnippetName, expectedRemoteItemsAfterUpdate, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceDynamicSnippetContentRemoteState(&service, name, dynamicSnippetName, expectedRemoteItemsAfterUpdate),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_dynamic_snippet_content.content", "content", expectedRemoteItemsAfterUpdate),
				),
			},
		},
	})
}

func TestAccFastlyServiceDynamicSnippetContent_external_snippet_is_removed(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	externalDynamicSnippetName := fmt.Sprintf("existing dynamic snippet %s", acctest.RandString(10))
	externalContent := "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

	managedDynamicSnippetName := fmt.Sprintf("dynamic snippet %s", acctest.RandString(10))
	managedContent := "if ( req.url ) {\n set req.http.my-updated-test-header = \"true\";\n}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDynamicSnippetContentConfigWithDynamicSnippet(name, managedDynamicSnippetName, managedContent, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", "1"),
				),
			},
			{
				PreConfig: func() {
					createDynamicSnippetThroughAPI(t, &service, externalDynamicSnippetName, gofastly.SnippetTypeHit, externalContent)
				},
				Config: testAccServiceDynamicSnippetContentConfigWithDynamicSnippet(name, managedDynamicSnippetName, managedContent, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceDynamicSnippetContentRemoteState(&service, name, managedDynamicSnippetName, managedContent),
					testAccCheckFastlyServiceDynamicSnippetContentRemoteStateDoesntExist(&service, name, externalDynamicSnippetName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_dynamic_snippet_content.content", "content", managedContent),
				),
			},
		},
	})
}

func TestAccFastlyServiceDynamicSnippetContent_normal_snippet_is_not_removed(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	normalSnippetName := fmt.Sprintf("normal dynamic snippet %s", acctest.RandString(10))
	normalContent := "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

	dynamicSnippetName := fmt.Sprintf("existing dynamic snippet %s", acctest.RandString(10))
	dynamicContent := "if ( req.url ) {\n set req.http.my-new-content-test-header = \"true\";\n}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDynamicSnippetContentConfigWithSnippet(name, normalSnippetName, normalContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "snippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", "0"),
				),
			},
			{
				Config: testAccServiceDynamicSnippetContentConfigWithSnippetAndDynamicSnippet(name, normalSnippetName, normalContent, dynamicSnippetName, dynamicContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceDynamicSnippetContentRemoteState(&service, name, dynamicSnippetName, dynamicContent),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "snippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_dynamic_snippet_content.content", "content", dynamicContent),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceDynamicSnippetContentRemoteState(service *gofastly.ServiceDetail, name, dynamicSnippetName, expectedContent string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		snippet, err := conn.GetSnippet(&gofastly.GetSnippetInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
			Name:           dynamicSnippetName,
		})
		if err != nil {
			return fmt.Errorf("error looking up snippet records for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		dynamicSnippet, err := conn.GetDynamicSnippet(&gofastly.GetDynamicSnippetInput{
			ServiceID: gofastly.ToValue(service.ServiceID),
			SnippetID: gofastly.ToValue(snippet.SnippetID),
		})
		if err != nil {
			return fmt.Errorf("error looking up Dynamic snippet content for (%s), snippet (%s): %s", gofastly.ToValue(service.Name), gofastly.ToValue(snippet.SnippetID), err)
		}

		if gofastly.ToValue(dynamicSnippet.Content) != expectedContent {
			return fmt.Errorf("error matching:\nexpected: %s\ngot: %s", expectedContent, gofastly.ToValue(dynamicSnippet.Content))
		}

		return nil
	}
}

func testAccCheckFastlyServiceDynamicSnippetContentRemoteStateDoesntExist(service *gofastly.ServiceDetail, name, dynamicSnippetName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		snippets, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up snippet records for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		for _, snippet := range snippets {
			if gofastly.ToValue(snippet.Name) == dynamicSnippetName {
				return fmt.Errorf("dynamic snippet (%s) exists in service (%s)", dynamicSnippetName, gofastly.ToValue(service.Name))
			}
		}

		return nil
	}
}

func createDynamicSnippetThroughAPI(t *testing.T, service *gofastly.ServiceDetail, dynamicSnippetName string, snippetType gofastly.SnippetType, content string) {
	conn := testAccProvider.Meta().(*APIClient).conn

	newVersion, err := conn.CloneVersion(&gofastly.CloneVersionInput{
		ServiceID:      gofastly.ToValue(service.ServiceID),
		ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
	})
	if err != nil {
		t.Fatalf("[ERR] Error cloning service version (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
	}

	dynamicSnippet, err := conn.CreateSnippet(&gofastly.CreateSnippetInput{
		ServiceID:      gofastly.ToValue(service.ServiceID),
		ServiceVersion: gofastly.ToValue(newVersion.Number),
		Name:           gofastly.ToPointer(dynamicSnippetName),
		Type:           gofastly.ToPointer(snippetType),
		Dynamic:        gofastly.ToPointer(1),
		Content:        gofastly.ToPointer("// vcl"),
	})
	if err != nil {
		t.Fatalf("[ERR] Error creating Dynamic snippet records for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
	}

	_, err = conn.ActivateVersion(&gofastly.ActivateVersionInput{
		ServiceID:      gofastly.ToValue(service.ServiceID),
		ServiceVersion: gofastly.ToValue(newVersion.Number),
	})
	if err != nil {
		t.Fatalf("[ERR] Error activating service version (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(newVersion.Number), err)
	}

	_, err = conn.UpdateDynamicSnippet(&gofastly.UpdateDynamicSnippetInput{
		ServiceID: gofastly.ToValue(service.ServiceID),
		SnippetID: gofastly.ToValue(dynamicSnippet.SnippetID),
		Content:   gofastly.ToPointer(content),
	})
	if err != nil {
		t.Fatalf("[ERR] Error update content for Dynamic snippet records for (%s), snippet id (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(dynamicSnippet.SnippetID), err)
	}

	latest, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ServiceID: gofastly.ToValue(service.ServiceID),
	})
	if err != nil {
		t.Fatalf("[ERR] Error retrieving service details for (%s): %s", gofastly.ToValue(service.ServiceID), err)
	}

	*service = *latest
}

func testAccServiceDynamicSnippetContentConfigWithSnippet(serviceName, snippetName, content string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  snippet {
	name = "%s"
	type = "hit"
	content = %q
  }

  force_destroy = true
}`, serviceName, domainName, backendName, snippetName, content)
}

func testAccServiceDynamicSnippetContentConfigWithSnippetAndDynamicSnippet(serviceName, snippetName, snippetContent, dynamicSnippetName, dynamicSnippetContent string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
variable "mydynamicsnippet" {
	type = object({ name=string, content=string })
	default = {
		name = "%s"
		content = %q
	}
}

resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  snippet {
	name = "%s"
	type = "hit"
	content = %q
  }

  dynamicsnippet {
	name = var.mydynamicsnippet.name
	type = "hit"
  }

  force_destroy = true
}

resource "fastly_service_dynamic_snippet_content" "content" {
    service_id = fastly_service_vcl.foo.id
    snippet_id = {for s in fastly_service_vcl.foo.dynamicsnippet : s.name => s.snippet_id}[var.mydynamicsnippet.name]
    content = var.mydynamicsnippet.content
}`, dynamicSnippetName, dynamicSnippetContent, serviceName, domainName, backendName, snippetName, snippetContent)
}

func testAccServiceDynamicSnippetContentConfigWithDynamicSnippet(serviceName, dynamicSnippetName, content string, manageSnippets bool) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
variable "mydynamicsnippet" {
	type = object({ name=string, content=string })
	default = {
		name = "%s"
		content = %q
	}
}

resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  dynamicsnippet {
	name = var.mydynamicsnippet.name
	type = "hit"
  }

  force_destroy = true
}

resource "fastly_service_dynamic_snippet_content" "content" {
  service_id      = fastly_service_vcl.foo.id
  snippet_id      = { for s in fastly_service_vcl.foo.dynamicsnippet : s.name => s.snippet_id }[var.mydynamicsnippet.name]
  manage_snippets = %t
  content         = var.mydynamicsnippet.content
}`, dynamicSnippetName, content, serviceName, domainName, backendName, manageSnippets)
}
