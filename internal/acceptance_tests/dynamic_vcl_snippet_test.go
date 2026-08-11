package acceptancetests

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceDynamicSnippetMetadataExplicit_basicAndUpdate(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-dynamic-vcl-snippet-%s", acctest.RandString(10))
	snippetName := fmt.Sprintf("dynamic_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceDynamicVCLSnippet(serviceName, snippetName, "recv", 100),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					CheckDynamicVCLSnippetExistsInFastly("fastly_service_cdn.test", snippetName, 1, "recv", 100),
					resource.TestCheckResourceAttr("fastly_service_dynamic_vcl_snippet.test", "name", snippetName),
					resource.TestCheckResourceAttr("fastly_service_dynamic_vcl_snippet.test", "type", "recv"),
					resource.TestCheckResourceAttr("fastly_service_dynamic_vcl_snippet.test", "priority", "100"),
					resource.TestCheckResourceAttr("fastly_service_dynamic_vcl_snippet.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_dynamic_vcl_snippet.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_dynamic_vcl_snippet.test", "snippet_id"),
					resource.TestCheckResourceAttrSet("fastly_service_dynamic_vcl_snippet.test", "id"),
				),
			},
			{
				Config: ConfigServiceDynamicVCLSnippet(serviceName, snippetName, "deliver", 50),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					CheckDynamicVCLSnippetExistsInFastly("fastly_service_cdn.test", snippetName, 1, "deliver", 50),
					resource.TestCheckResourceAttr("fastly_service_dynamic_vcl_snippet.test", "name", snippetName),
					resource.TestCheckResourceAttr("fastly_service_dynamic_vcl_snippet.test", "type", "deliver"),
					resource.TestCheckResourceAttr("fastly_service_dynamic_vcl_snippet.test", "priority", "50"),
					resource.TestCheckResourceAttrSet("fastly_service_dynamic_vcl_snippet.test", "snippet_id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceDynamicSnippetMetadataExplicit_import(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-dynamic-vcl-snippet-%s", acctest.RandString(10))
	snippetName := fmt.Sprintf("dynamic_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceDynamicVCLSnippet(serviceName, snippetName, "recv", 100),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					CheckDynamicVCLSnippetExistsInFastly("fastly_service_cdn.test", snippetName, 1, "recv", 100),
				),
			},
			{
				ResourceName:      "fastly_service_dynamic_vcl_snippet.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importDynamicVCLSnippetID("fastly_service_dynamic_vcl_snippet.test"),
			},
		},
	})
}

func TestAccFastlyServiceDynamicSnippetMetadataExplicit_rejectsRegularSnippetOnImport(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-dynamic-vcl-snippet-%s", acctest.RandString(10))
	snippetName := fmt.Sprintf("regular_%s", acctest.RandString(10))
	content := snippetBoilerplate("regular")
	snippetFile := writeSnippetFile(t, "dynamic-import-regular.vcl", content)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceVCLSnippetWithFile(serviceName, snippetName, "recv", 100, snippetFile),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					CheckSnippetExistsInFastly("fastly_service_cdn.test", snippetName, 1, content),
				),
			},
			{
				Config:            ConfigServiceDynamicVCLSnippet(serviceName, snippetName, "recv", 100),
				ResourceName:      "fastly_service_dynamic_vcl_snippet.test",
				ImportState:       true,
				ImportStateIdFunc: importRegularSnippetAsDynamicVCLSnippetID("fastly_service_vcl_snippet.test"),
				ExpectError:       regexp.MustCompile(`regular.*expected dynamic snippet`),
			},
		},
	})
}

func CheckDynamicVCLSnippetExistsInFastly(resourceName, snippetName string, version int, expectedType string, expectedPriority int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		snippet, err := client.GetSnippet(context.Background(), &fastly.GetSnippetInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           snippetName,
		})
		if err != nil {
			return fmt.Errorf("error fetching dynamic VCL snippet %q in service %s version %d: %w", snippetName, rs.Primary.ID, version, err)
		}

		if fastly.ToValue(snippet.Dynamic) != 1 {
			return fmt.Errorf("snippet %q is regular; expected dynamic snippet", snippetName)
		}

		if got := string(fastly.ToValue(snippet.Type)); got != expectedType {
			return fmt.Errorf("dynamic snippet %q type = %q, want %q", snippetName, got, expectedType)
		}

		priority, err := strconv.Atoi(fastly.ToValue(snippet.Priority))
		if err != nil {
			return fmt.Errorf("error parsing dynamic snippet priority %q: %w", fastly.ToValue(snippet.Priority), err)
		}
		if priority != expectedPriority {
			return fmt.Errorf("dynamic snippet %q priority = %d, want %d", snippetName, priority, expectedPriority)
		}

		if fastly.ToValue(snippet.SnippetID) == "" {
			return fmt.Errorf("dynamic snippet %q has empty snippet_id", snippetName)
		}

		return nil
	}
}

func CheckDynamicSnippetContentResourceInFastly(resourceName string, expectedContent string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		serviceID := rs.Primary.Attributes["service_id"]
		snippetID := rs.Primary.Attributes["snippet_id"]
		if serviceID == "" || snippetID == "" {
			return fmt.Errorf("missing dynamic snippet content identity fields for %s: service_id=%q snippet_id=%q", resourceName, serviceID, snippetID)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		dynamicSnippet, err := client.GetDynamicSnippet(context.Background(), &fastly.GetDynamicSnippetInput{
			ServiceID: serviceID,
			SnippetID: snippetID,
		})
		if err != nil {
			return fmt.Errorf("error fetching dynamic VCL snippet content %q in service %s: %w", snippetID, serviceID, err)
		}

		if fastly.ToValue(dynamicSnippet.Content) != expectedContent {
			return fmt.Errorf("dynamic snippet content mismatch for snippet %q in service %s:\nexpected: %q\nactual:   %q", snippetID, serviceID, expectedContent, fastly.ToValue(dynamicSnippet.Content))
		}

		return nil
	}
}

func importDynamicVCLSnippetID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		serviceID := rs.Primary.Attributes["service_id"]
		version := rs.Primary.Attributes["version"]
		name := rs.Primary.Attributes["name"]

		if serviceID == "" || version == "" || name == "" {
			return "", fmt.Errorf("missing import identity fields for %s: service_id=%q version=%q name=%q", resourceName, serviceID, version, name)
		}

		return fmt.Sprintf("%s/%s/%s", serviceID, version, name), nil
	}
}

func importRegularSnippetAsDynamicVCLSnippetID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		serviceID := rs.Primary.Attributes["service_id"]
		version := rs.Primary.Attributes["version"]
		name := rs.Primary.Attributes["name"]

		if serviceID == "" || version == "" || name == "" {
			return "", fmt.Errorf("missing import identity fields for %s: service_id=%q version=%q name=%q", resourceName, serviceID, version, name)
		}

		return fmt.Sprintf("%s/%s/%s", serviceID, version, name), nil
	}
}
