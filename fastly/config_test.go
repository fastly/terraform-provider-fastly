package fastly

import (
	"strings"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/terraform-providers/terraform-provider-fastly/version"
)

func TestUserAgentContainsProviderVersion(t *testing.T) {
	c := Config{
		ApiKey:  "someapikey",
		BaseURL: "http://localhost",
	}
	_, err := c.Client()

	if err != nil {
		t.Errorf("Failed to create client: %s", err)
	}
	configuredUserAgent := gofastly.UserAgent
	if !strings.Contains(configuredUserAgent, TerraformProviderProductUserAgent+"/"+version.ProviderVersion) {
		t.Errorf("User agent doesn't contain the terraform provider version")
	}
}
