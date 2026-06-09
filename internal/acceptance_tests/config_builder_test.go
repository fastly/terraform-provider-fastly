package acceptancetests

import (
	"strings"
	"testing"
)

func TestBuildConfig(t *testing.T) {
	config := BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": "test-service",
			"DOMAIN_NAME":  "example.com",
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
	)

	// Check that the service type is correct
	if !strings.Contains(config, `resource "fastly_service_cdn_auto" "test"`) {
		t.Errorf("Config should contain service resource declaration")
	}

	// Check that attributes are present
	if !strings.Contains(config, `name          = "test-service"`) {
		t.Errorf("Config should contain replaced service name")
	}

	if !strings.Contains(config, `force_destroy = true`) {
		t.Errorf("Config should contain force_destroy attribute")
	}

	// Check that domain block is present
	if !strings.Contains(config, `domain {`) {
		t.Errorf("Config should contain domain block")
	}

	if !strings.Contains(config, `name = "example.com"`) {
		t.Errorf("Config should contain replaced domain name")
	}

	// Check that config starts with resource and ends with closing brace
	if !strings.HasPrefix(strings.TrimSpace(config), "resource") {
		t.Errorf("Config should start with resource declaration")
	}

	if !strings.HasSuffix(strings.TrimSpace(config), "}") {
		t.Errorf("Config should end with closing brace")
	}
}

func TestBuildConfigMultipleBlocks(t *testing.T) {
	config := BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": "test-service",
			"DOMAIN_NAME":  "example.com",
			"BACKEND_NAME": "test-backend",
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_single.tf",
	)

	// Check that both blocks are present
	if !strings.Contains(config, `domain {`) {
		t.Errorf("Config should contain domain block")
	}

	if !strings.Contains(config, `backend {`) {
		t.Errorf("Config should contain backend block")
	}

	if !strings.Contains(config, `name              = "test-backend"`) {
		t.Errorf("Config should contain replaced backend name")
	}
}

func TestBuildConfigOutput(t *testing.T) {
	config := BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": "my-service",
			"DOMAIN_NAME":  "example.com",
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
	)

	expected := `resource "fastly_service_cdn_auto" "test" {
  name          = "my-service"
  force_destroy = true
  domain {
    name = "example.com"
  }
}`

	if strings.TrimSpace(config) != expected {
		t.Errorf("Config mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, config)
	}
}

func TestBuildConfigMultipleBlocksOutput(t *testing.T) {
	config := BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": "my-service",
			"DOMAIN_NAME":  "example.com",
			"BACKEND_NAME": "my-backend",
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_single.tf",
	)

	expected := `resource "fastly_service_cdn_auto" "test" {
  name          = "my-service"
  force_destroy = true
  domain {
    name = "example.com"
  }
  backend {
    name              = "my-backend"
    address           = "api.example.com"
    port              = 443
    use_ssl           = true
    ssl_cert_hostname = "api.example.com"
    ssl_sni_hostname  = "api.example.com"
  }
}`

	if strings.TrimSpace(config) != expected {
		t.Errorf("Config mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, config)
	}
}

func TestBuildConfigServiceTypes(t *testing.T) {
	tests := []struct {
		name        string
		serviceType ServiceType
		expected    string
	}{
		{
			name:        "CDN Auto service",
			serviceType: ServiceCDNAuto,
			expected:    `resource "fastly_service_cdn_auto" "test"`,
		},
		{
			name:        "Compute Auto service",
			serviceType: ServiceComputeAuto,
			expected:    `resource "fastly_service_compute_auto" "test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := BuildConfig(
				tt.serviceType,
				map[string]string{
					"SERVICE_NAME": "test",
				},
			)

			if !strings.Contains(config, tt.expected) {
				t.Errorf("Config should contain %q", tt.expected)
			}
		})
	}
}
