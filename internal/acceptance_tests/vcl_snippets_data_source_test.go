package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFastlyDataSourceVCLSnippets_Config(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-vcl-snippets-ds-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	regularNameOne := fmt.Sprintf("regular_recv_%s", acctest.RandString(10))
	regularNameTwo := fmt.Sprintf("regular_deliver_%s", acctest.RandString(10))
	dynamicName := fmt.Sprintf("dynamic_recv_%s", acctest.RandString(10))
	contentOne := `set req.http.X-Terraform-Snippet-Data-Source = "one";`
	contentTwo := `set resp.http.X-Terraform-Snippet-Data-Source = "two";`
	snippetFileOne := writeSnippetFile(t, "data-source-one.vcl", contentOne)
	snippetFileTwo := writeSnippetFile(t, "data-source-two.vcl", contentTwo)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDataSourceVCLSnippets(serviceName, domainName, backendName, regularNameOne, regularNameTwo, dynamicName, snippetFileOne, snippetFileTwo),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("data.fastly_vcl_snippets.example", "vcl_snippets.#", "3"),
					resource.TestCheckResourceAttrSet("data.fastly_vcl_snippets.example", "id"),
					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_vcl_snippets.example", "vcl_snippets.*", map[string]string{
						"name":     regularNameOne,
						"type":     "recv",
						"priority": "100",
						"dynamic":  "false",
						"content":  contentOne,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_vcl_snippets.example", "vcl_snippets.*", map[string]string{
						"name":     regularNameTwo,
						"type":     "deliver",
						"priority": "50",
						"dynamic":  "false",
						"content":  contentTwo,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_vcl_snippets.example", "vcl_snippets.*", map[string]string{
						"name":     dynamicName,
						"type":     "recv",
						"priority": "25",
						"dynamic":  "true",
					}),
				),
			},
		},
	})
}

func ConfigDataSourceVCLSnippets(serviceName, domainName, backendName, regularNameOne, regularNameTwo, dynamicName, snippetFilePathOne, snippetFilePathTwo string) string {
	serviceConfig := BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":             serviceName,
			"DOMAIN_NAME":              domainName,
			"BACKEND_NAME":             backendName,
			"SNIPPET_NAME_ONE":         regularNameOne,
			"SNIPPET_NAME_TWO":         regularNameTwo,
			"SNIPPET_FILE_PATH_ONE":    snippetFilePathOne,
			"SNIPPET_FILE_PATH_TWO":    snippetFilePathTwo,
			"DYNAMIC_SNIPPET_NAME":     dynamicName,
			"DYNAMIC_SNIPPET_TYPE":     "recv",
			"DYNAMIC_SNIPPET_PRIORITY": "25",
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_single.tf",
		"internal/acceptance_tests/blocks/snippet_nested_multiple.tf",
		"internal/acceptance_tests/blocks/dynamic_snippet_nested.tf",
	)

	dataSource := `
data "fastly_vcl_snippets" "example" {
  depends_on      = [fastly_service_cdn_auto.test]
  service_id      = fastly_service_cdn_auto.test.id
  service_version = fastly_service_cdn_auto.test.active_version
}
`

	return joinBlocks(serviceConfig, dataSource)
}
