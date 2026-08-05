package acceptancetests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceVCLSnippet_basicAndUpdate(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-snippet-%s", acctest.RandString(10))
	snippetName := fmt.Sprintf("snippet_%s", acctest.RandString(10))
	contentOne := snippetBoilerplate("one")
	contentTwo := snippetBoilerplate("two")
	snippetFileOne := writeSnippetFile(t, "explicit-one.vcl", contentOne)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceVCLSnippetWithFile(serviceName, snippetName, "deliver", 100, snippetFileOne),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					CheckSnippetExistsInFastly("fastly_service_cdn.test", snippetName, 1, contentOne),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "name", snippetName),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "type", "deliver"),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "priority", "100"),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_vcl_snippet.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_vcl_snippet.test", "id"),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "content", contentOne),
				),
			},
			{
				Config: ConfigServiceVCLSnippetInline(serviceName, snippetName, "deliver", 50, contentTwo),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					CheckSnippetExistsInFastly("fastly_service_cdn.test", snippetName, 1, contentTwo),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "name", snippetName),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "type", "deliver"),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "priority", "50"),
					resource.TestCheckResourceAttr("fastly_service_vcl_snippet.test", "content", contentTwo),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCLSnippet_import(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-snippet-%s", acctest.RandString(10))
	snippetName := fmt.Sprintf("snippet_%s", acctest.RandString(10))
	content := snippetBoilerplate("import")
	snippetFile := writeSnippetFile(t, "explicit-import.vcl", content)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceVCLSnippetWithFile(serviceName, snippetName, "deliver", 100, snippetFile),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					CheckSnippetExistsInFastly("fastly_service_cdn.test", snippetName, 1, content),
				),
			},
			{
				ResourceName:      "fastly_service_vcl_snippet.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importSnippetID("fastly_service_vcl_snippet.test"),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withSnippet(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-snippet-auto-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	snippetName := fmt.Sprintf("snippet_%s", acctest.RandString(10))
	content := snippetBoilerplate("one")
	snippetFile := writeSnippetFile(t, "auto-one.vcl", content)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithSnippetFile(serviceName, domainName, backendName, snippetName, "deliver", 100, snippetFile),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckSnippetExistsInFastlyAtActiveVersion("fastly_service_cdn_auto.test", snippetName, content),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.name", snippetName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.type", "deliver"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.priority", "100"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.content", content),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withSnippetContentUpdate(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-snippet-auto-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	snippetName := fmt.Sprintf("snippet_%s", acctest.RandString(10))
	contentOne := snippetBoilerplate("one")
	contentTwo := snippetBoilerplate("two")
	snippetFileOne := writeSnippetFile(t, "auto-one.vcl", contentOne)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithSnippetFile(serviceName, domainName, backendName, snippetName, "deliver", 100, snippetFileOne),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckSnippetExistsInFastlyAtActiveVersion("fastly_service_cdn_auto.test", snippetName, contentOne),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.content", contentOne),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithSnippetInline(serviceName, domainName, backendName, snippetName, "deliver", 100, contentTwo),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
					CheckSnippetExistsInFastlyAtActiveVersion("fastly_service_cdn_auto.test", snippetName, contentTwo),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.name", snippetName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.type", "deliver"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.priority", "100"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.0.content", contentTwo),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withMultipleSnippets(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-snippet-auto-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	snippetNameOne := fmt.Sprintf("recv_%s", acctest.RandString(10))
	snippetNameTwo := fmt.Sprintf("deliver_%s", acctest.RandString(10))
	contentOne := `set req.http.X-Terraform-Snippet = "recv";`
	contentTwo := `set resp.http.X-Terraform-Snippet = "deliver";`
	snippetFileOne := writeSnippetFile(t, "multi-one.vcl", contentOne)
	snippetFileTwo := writeSnippetFile(t, "multi-two.vcl", contentTwo)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleSnippets(serviceName, domainName, backendName, snippetNameOne, snippetNameTwo, snippetFileOne, snippetFileTwo),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckSnippetExistsInFastlyAtActiveVersion("fastly_service_cdn_auto.test", snippetNameOne, contentOne),
					CheckSnippetExistsInFastlyAtActiveVersion("fastly_service_cdn_auto.test", snippetNameTwo, contentTwo),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "snippet.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withDuplicateSnippetNames(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-snippet-auto-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	fileOne := writeSnippetFile(t, "duplicate-one.vcl", snippetBoilerplate("duplicate-one"))
	fileTwo := writeSnippetFile(t, "duplicate-two.vcl", snippetBoilerplate("duplicate-two"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigCDNAutoWithDuplicateSnippets(serviceName, domainName, backendName, fileOne, fileTwo),
				ExpectError: regexp.MustCompile("multiple snippets with the same name"),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withInvalidSnippetType(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-snippet-auto-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	filePath := writeSnippetFile(t, "invalid-type.vcl", snippetBoilerplate("invalid-type"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigCDNAutoWithInvalidSnippetType(serviceName, domainName, backendName, filePath),
				ExpectError: regexp.MustCompile(`Attribute snippet\[0\]\.type value must be one of`),
			},
		},
	})
}

func CheckSnippetExistsInFastlyAtActiveVersion(resourceName, snippetName string, expectedContent string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		version, err := strconv.Atoi(rs.Primary.Attributes["active_version"])
		if err != nil {
			return fmt.Errorf("error parsing active_version for %s: %w", resourceName, err)
		}

		return checkSnippetExistsInFastly(rs.Primary.ID, snippetName, version, expectedContent)
	}
}

func CheckSnippetExistsInFastly(resourceName, snippetName string, version int, expectedContent string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		return checkSnippetExistsInFastly(rs.Primary.ID, snippetName, version, expectedContent)
	}
}

func checkSnippetExistsInFastly(serviceID, snippetName string, version int, expectedContent string) error {
	client, err := NewFastlyClient()
	if err != nil {
		return fmt.Errorf("error creating Fastly client: %w", err)
	}

	snippet, err := client.GetSnippet(context.Background(), &fastly.GetSnippetInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           snippetName,
	})
	if err != nil {
		return fmt.Errorf("error fetching VCL snippet %q in service %s version %d: %w", snippetName, serviceID, version, err)
	}

	if fastly.ToValue(snippet.Dynamic) != 0 {
		return fmt.Errorf("snippet %q is dynamic; expected regular snippet", snippetName)
	}

	if fastly.ToValue(snippet.Content) != expectedContent {
		return fmt.Errorf("snippet %q content mismatch in service %s version %d:\nexpected: %q\nactual:   %q", snippetName, serviceID, version, expectedContent, fastly.ToValue(snippet.Content))
	}

	return nil
}

func importSnippetID(resourceName string) resource.ImportStateIdFunc {
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

func snippetBoilerplate(label string) string {
	return fmt.Sprintf(`set resp.http.X-Terraform-Snippet-Test = %q;`, label)
}

func writeSnippetFile(t *testing.T, name, content string) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "snippets")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("error creating snippet temp dir: %s", err)
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("error writing snippet fixture: %s", err)
	}

	return path
}
