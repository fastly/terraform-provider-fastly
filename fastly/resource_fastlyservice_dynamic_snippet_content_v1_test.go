package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccFastlyServiceDynamicSnippetContentV1_create(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	dynamicSnippetName := fmt.Sprintf("dynamic snippet %s", acctest.RandString(10))

	expectedRemoteItems := "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDynamicSnippetContentV1Config(name, dynamicSnippetName, expectedRemoteItems),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceDynamicSnippetContentV1RemoteState(&service, name, dynamicSnippetName, expectedRemoteItems),
					resource.TestCheckResourceAttr("fastly_service_dynamic_snippet_content_v1.content", "content", expectedRemoteItems),
				),
			},
		},
	})
}

func TestAccFastlyServiceDynamicSnippetContentV1_update(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	dynamicSnippetName := fmt.Sprintf("dynamic snippet %s", acctest.RandString(10))

	expectedRemoteItems := "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

	expectedRemoteItemsAfterUpdate := "if ( req.url ) {\n set req.http.my-updated-test-header = \"true\";\n}"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDynamicSnippetContentV1Config(name, dynamicSnippetName, expectedRemoteItems),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceDynamicSnippetContentV1RemoteState(&service, name, dynamicSnippetName, expectedRemoteItems),
					resource.TestCheckResourceAttr("fastly_service_dynamic_snippet_content_v1.content", "content", expectedRemoteItems),
				),
			},
			{
				Config: testAccServiceDynamicSnippetContentV1Config(name, dynamicSnippetName, expectedRemoteItemsAfterUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceDynamicSnippetContentV1RemoteState(&service, name, dynamicSnippetName, expectedRemoteItemsAfterUpdate),
					resource.TestCheckResourceAttr("fastly_service_dynamic_snippet_content_v1.content", "content", expectedRemoteItemsAfterUpdate),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceDynamicSnippetContentV1RemoteState(service *gofastly.ServiceDetail, name, dynamicSnippetName string, expectedContent string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		snippet, err := conn.GetSnippet(&gofastly.GetSnippetInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
			Name:    dynamicSnippetName,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up snippet records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		dynamicSnippet, err := conn.GetDynamicSnippet(&gofastly.GetDynamicSnippetInput{
			Service: service.ID,
			ID:      snippet.ID,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dynamic snippet content for (%s), snippet (%s): %s", service.Name, snippet.ID, err)
		}

		if dynamicSnippet.Content != expectedContent {
			return fmt.Errorf("[ERR] Error matching:\nexpected: %s\ngot: %s", expectedContent, dynamicSnippet.Content)
		}

		return nil
	}
}

func testAccServiceDynamicSnippetContentV1Config(serviceName, dynamicSnippetName, content string) string {
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

resource "fastly_service_v1" "foo" {
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

resource "fastly_service_dynamic_snippet_content_v1" "content" {
    service_id = fastly_service_v1.foo.id
    snippet_id = {for s in fastly_service_v1.foo.dynamicsnippet : s.name => s.snippet_id}[var.mydynamicsnippet.name]
    content = var.mydynamicsnippet.content
}`, dynamicSnippetName, content, serviceName, domainName, backendName)
}
