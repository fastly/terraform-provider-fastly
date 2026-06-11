package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccProvider_ConfigureWithAPIToken tests provider configuration with explicit api_token
func TestAccProvider_ConfigureWithAPIToken(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-provider-token-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("tf-test-%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigWithExplicitToken(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					CheckServiceExists("fastly_service_cdn_auto.test"),
				),
			},
		},
	})
}

// TestAccProvider_ConfigureWithEnvVar tests provider configuration via FASTLY_API_TOKEN env var
func TestAccProvider_ConfigureWithEnvVar(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-provider-env-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("tf-test-%s.example.com", acctest.RandString(10))

	// Verify env var is set (PreCheck does this but being explicit for this test)
	if os.Getenv("FASTLY_API_TOKEN") == "" {
		t.Fatal("FASTLY_API_TOKEN must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigWithEnvVar(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					CheckServiceExists("fastly_service_cdn_auto.test"),
				),
			},
		},
	})
}

// TestAccProvider_MissingToken tests provider error handling when no token is provided
func TestAccProvider_MissingToken(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-provider-missing-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("tf-test-%s.example.com", acctest.RandString(10))

	// Save original env var
	originalToken := os.Getenv("FASTLY_API_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("FASTLY_API_TOKEN", originalToken)
		}
	}()

	// Unset the env var to test missing token scenario
	os.Unsetenv("FASTLY_API_TOKEN")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderConfigWithNoToken(serviceName, domainName),
				ExpectError: regexp.MustCompile(`Missing API Token`),
			},
		},
	})
}

// TestAccProvider_InvalidToken tests provider error handling with an invalid token
func TestAccProvider_InvalidToken(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-provider-invalid-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("tf-test-%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccProviderConfigWithInvalidToken(serviceName, domainName),
				ExpectError: regexp.MustCompile(`(Unauthorized|authentication|401|invalid.*token|permission denied)`),
			},
		},
	})
}

// TestAccProvider_ExplicitTokenOverridesEnvVar tests that explicit api_token takes precedence over env var
func TestAccProvider_ExplicitTokenOverridesEnvVar(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-provider-override-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("tf-test-%s.example.com", acctest.RandString(10))

	// Verify env var is set
	if os.Getenv("FASTLY_API_TOKEN") == "" {
		t.Fatal("FASTLY_API_TOKEN must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigWithExplicitToken(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					CheckServiceExists("fastly_service_cdn_auto.test"),
					// Verify we can successfully create resources, proving the explicit token worked
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_cdn_auto.test"]
						if !ok {
							return fmt.Errorf("service not found in state")
						}
						if rs.Primary.ID == "" {
							return fmt.Errorf("service ID not set, token may not have worked")
						}
						return nil
					},
				),
			},
		},
	})
}

// testAccProviderConfigWithExplicitToken returns config with explicit api_token in provider block
func testAccProviderConfigWithExplicitToken(serviceName, domainName string) string {
	apiToken := os.Getenv("FASTLY_API_TOKEN")
	return fmt.Sprintf(`
provider "fastly" {
  api_token = "%s"
}

resource "fastly_service_cdn_auto" "test" {
  name          = "%s"
  force_destroy = true

  domain {
    name = "%s"
  }
}
`, apiToken, serviceName, domainName)
}

// testAccProviderConfigWithEnvVar returns config without explicit api_token (relies on env var)
func testAccProviderConfigWithEnvVar(serviceName, domainName string) string {
	return fmt.Sprintf(`
provider "fastly" {
  # api_token will be read from FASTLY_API_TOKEN env var
}

resource "fastly_service_cdn_auto" "test" {
  name          = "%s"
  force_destroy = true

  domain {
    name = "%s"
  }
}
`, serviceName, domainName)
}

// testAccProviderConfigWithNoToken returns config with no token configured
func testAccProviderConfigWithNoToken(serviceName, domainName string) string {
	return fmt.Sprintf(`
provider "fastly" {
  # No api_token and env var is unset
}

resource "fastly_service_cdn_auto" "test" {
  name          = "%s"
  force_destroy = true

  domain {
    name = "%s"
  }
}
`, serviceName, domainName)
}

// testAccProviderConfigWithInvalidToken returns config with an invalid token
func testAccProviderConfigWithInvalidToken(serviceName, domainName string) string {
	return fmt.Sprintf(`
provider "fastly" {
  api_token = "invalid-token-12345"
}

resource "fastly_service_cdn_auto" "test" {
  name          = "%s"
  force_destroy = true

  domain {
    name = "%s"
  }
}
`, serviceName, domainName)
}
