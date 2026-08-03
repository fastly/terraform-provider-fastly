package acceptancetests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceVCL_basicAndUpdate(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-vcl-%s", acctest.RandString(10))
	vclName := fmt.Sprintf("main_%s", acctest.RandString(10))
	contentOne := readVCLFixture(t, "vcl/main.vcl")
	contentTwo := vclBoilerplate("two")
	vclFileOne := vclFixturePath(t, "vcl/main.vcl")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceVCLWithFile(serviceName, vclName, vclFileOne),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					CheckVCLExistsInFastly("fastly_service_cdn.test", vclName, 1),
					resource.TestCheckResourceAttr("fastly_service_vcl.main", "name", vclName),
					resource.TestCheckResourceAttr("fastly_service_vcl.main", "main", "true"),
					resource.TestCheckResourceAttr("fastly_service_vcl.main", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_vcl.main", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_vcl.main", "id"),
					resource.TestCheckResourceAttr("fastly_service_vcl.main", "content", contentOne),
				),
			},
			{
				Config: ConfigServiceVCLInline(serviceName, vclName, contentTwo),
				Check: resource.ComposeTestCheckFunc(
					CheckVCLExistsInFastly("fastly_service_cdn.test", vclName, 1),
					resource.TestCheckResourceAttr("fastly_service_vcl.main", "name", vclName),
					resource.TestCheckResourceAttr("fastly_service_vcl.main", "main", "true"),
					resource.TestCheckResourceAttr("fastly_service_vcl.main", "content", contentTwo),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withVCLAndContentUpdate(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-vcl-auto-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	vclName := fmt.Sprintf("main_%s", acctest.RandString(10))
	contentOne := readVCLFixture(t, "vcl/main.vcl")
	contentTwo := vclBoilerplate("two")
	vclFileOne := vclFixturePath(t, "vcl/main.vcl")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithVCLFile(serviceName, domainName, backendName, vclName, vclFileOne),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckVCLExistsInFastly("fastly_service_cdn_auto.test", vclName, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.name", vclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.main", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.content", contentOne),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithVCLInline(serviceName, domainName, backendName, vclName, contentTwo),
				Check: resource.ComposeTestCheckFunc(
					CheckVCLExistsInFastly("fastly_service_cdn_auto.test", vclName, 2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.name", vclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.main", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.content", contentTwo),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withMultipleVCLFiles(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-vcl-auto-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	mainName := fmt.Sprintf("main_%s", acctest.RandString(10))
	includeName := fmt.Sprintf("include_%s", acctest.RandString(10))
	mainContent := fmt.Sprintf("include %q;\n%s", includeName, vclBoilerplate("multi"))
	includeContent := `sub set_test_header {
  set resp.http.X-Terraform-Include = "true";
}
`
	mainFile := writeVCLFile(t, "auto-main-with-include.vcl", mainContent)
	includeFile := writeVCLFile(t, "auto-include.vcl", includeContent)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleVCLFiles(serviceName, domainName, backendName, mainName, includeName, mainFile, includeFile),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckVCLExistsInFastly("fastly_service_cdn_auto.test", mainName, 1),
					CheckVCLExistsInFastly("fastly_service_cdn_auto.test", includeName, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.name", mainName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.main", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.content", mainContent),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.1.name", includeName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.1.main", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.1.content", includeContent),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withVCLHeredocContent(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-vcl-auto-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	vclName := fmt.Sprintf("main_%s", acctest.RandString(10))
	content := vclBoilerplate("heredoc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithVCLHeredoc(serviceName, domainName, backendName, vclName, content),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckVCLExistsInFastly("fastly_service_cdn_auto.test", vclName, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.name", vclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.main", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "vcl.0.content", content),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config:   ConfigCDNAutoWithVCLHeredoc(serviceName, domainName, backendName, vclName, content),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_VCLValidation(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-vcl-invalid-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	mainFile := writeVCLFile(t, "invalid-main-one.vcl", vclBoilerplate("one"))
	secondMainFile := writeVCLFile(t, "invalid-main-two.vcl", vclBoilerplate("two"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigCDNAutoWithInvalidMultipleMainVCLFiles(serviceName, domainName, backendName, mainFile, secondMainFile),
				ExpectError: regexp.MustCompile(`only one custom VCL file can have main = true`),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_VCLValidationNoMain(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-vcl-invalid-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	includeFile := writeVCLFile(t, "invalid-include-only.vcl", `sub set_test_header {
  set resp.http.X-Terraform-Include = "true";
}
`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigCDNAutoWithInvalidNoMainVCLFile(serviceName, domainName, backendName, includeFile),
				ExpectError: regexp.MustCompile(`one custom VCL file must have main = true`),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_VCLValidationInvalidSyntax(t *testing.T) {
	t.Parallel()

	serviceName := fmt.Sprintf("tf-test-vcl-invalid-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	vclName := fmt.Sprintf("main_%s", acctest.RandString(10))
	invalidContent := `sub vcl_recv {
#FASTLY recv
  this is not valid vcl
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigCDNAutoWithVCLInline(serviceName, domainName, backendName, vclName, invalidContent),
				ExpectError: regexp.MustCompile(`Error validating service version|VCL|invalid`),
			},
		},
	})
}

func vclFixturePath(t *testing.T, name string) string {
	t.Helper()

	path, err := filepath.Abs(filepath.Join("fixtures", name))
	if err != nil {
		t.Fatalf("failed to resolve VCL fixture path: %v", err)
	}
	return path
}

func readVCLFixture(t *testing.T, name string) string {
	t.Helper()

	content, err := os.ReadFile(vclFixturePath(t, name))
	if err != nil {
		t.Fatalf("failed to read VCL fixture file: %v", err)
	}
	return string(content)
}

func writeVCLFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write VCL fixture file: %v", err)
	}
	return path
}

func CheckVCLExistsInFastly(serviceResourceName, vclName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceResourceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceResourceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		vcl, err := client.GetVCL(context.Background(), &fastly.GetVCLInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           vclName,
		})
		if err != nil {
			return fmt.Errorf("error fetching custom VCL from Fastly: %w", err)
		}

		if vcl == nil {
			return fmt.Errorf("custom VCL %s not found in Fastly", vclName)
		}

		return nil
	}
}

func vclBoilerplate(marker string) string {
	return fmt.Sprintf(`sub vcl_recv {
#FASTLY recv
  return(lookup);
}

sub vcl_hash {
  set req.hash += req.url;
  set req.hash += req.http.host;
#FASTLY hash
  return(hash);
}

sub vcl_hit {
#FASTLY hit
  return(deliver);
}

sub vcl_miss {
#FASTLY miss
  return(fetch);
}

sub vcl_pass {
#FASTLY pass
  return(pass);
}

sub vcl_fetch {
#FASTLY fetch
  return(deliver);
}

sub vcl_error {
#FASTLY error
  return(deliver);
}

sub vcl_deliver {
#FASTLY deliver
  set resp.http.X-Terraform-VCL-Test = %[1]q;
  return(deliver);
}

sub vcl_log {
#FASTLY log
}
`, marker)
}
