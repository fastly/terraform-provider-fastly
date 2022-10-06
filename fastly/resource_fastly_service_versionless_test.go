package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyServiceVCL_creation_with_versionless_resources(t *testing.T) {
	var service gofastly.ServiceDetail

	serviceName := "tf-test-service-versionless"
	dictionaryName := "tf_test_dictionary_versionless"
	aclName := "tf_test_acl_versionless"
	dynamicSnippetName := "tf_test_dynamic_snippet"

	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigCreateServiceWithOneACLDictionaryAndDynamicSnippet(serviceName, dictionaryName, aclName, dynamicSnippetName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.service", &service),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_dictionary_items.items", "items.%", "3"),
					resource.TestCheckResourceAttrSet("fastly_service_dynamic_snippet_content.dyn_content", "content"),
				),
			},
		},
	})
}

func testAccServiceVCLConfigCreateServiceWithOneACLDictionaryAndDynamicSnippet(serviceName, dictionaryName, aclName, dynamicSnippetName, domain string) string {
	return fmt.Sprintf(`
locals {
  service_name         = "%s"
  dictionary_name      = "%s"
  acl_name             = "%s"
  dynamic_snippet_name = "%s"
  domain               = "%s"
}

resource "fastly_service_vcl" "service" {
  name = local.service_name

  domain {
    name = local.domain
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  dictionary {
    name = local.dictionary_name
  }

  acl {
    name = local.acl_name
  }

  dynamicsnippet {
    name     = local.dynamic_snippet_name
    type     = "recv"
    priority = 110
  }

  force_destroy = true

}

resource "fastly_service_dictionary_items" "items" {
  service_id    = fastly_service_vcl.service.id
  dictionary_id = { for s in fastly_service_vcl.service.dictionary : s.name => s.dictionary_id }[local.dictionary_name]

  items = {
    "US" : "en-US",
    "FR" : "fr-FR",
    "NL" : "nl-NL",
  }
}

resource "fastly_service_acl_entries" "entries" {
  service_id = "${fastly_service_vcl.service.id}"
  acl_id     = { for d in fastly_service_vcl.service.acl : d.name => d.acl_id }[local.acl_name]

  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
  }
}

resource "fastly_service_dynamic_snippet_content" "dyn_content" {
  service_id = fastly_service_vcl.service.id
  snippet_id = { for s in fastly_service_vcl.service.dynamicsnippet : s.name => s.snippet_id }[local.dynamic_snippet_name]

  content = <<EOT
if (!req.http.Accept-Language) {
  set req.http.Accept-Language = table.lookup(${local.dictionary_name}, client.geo.country_code, "en-US");
}

# block all requests to Admin pages from IP addresses not in office_ip_ranges
if (req.url ~ "^/admin" && ! (client.ip ~ ${local.acl_name})) {
  error 403 "Forbidden";
}
EOT
}`, serviceName, dictionaryName, aclName, dynamicSnippetName, domain)
}
