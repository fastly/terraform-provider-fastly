package fastly

import (
	"testing"
)

func TestUserAgentContainsProviderVersion(t *testing.T) {
	c := Config{
		ApiKey:  "someapikey",
		BaseURL: "http://localhost",
	}
	_, diagnostics := c.Client()

	if diagnostics.HasError() {
		t.Errorf("Failed to create client: %s", diagToErr(diagnostics))
	}
}
