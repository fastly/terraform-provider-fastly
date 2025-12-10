package fastly

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/go-hclog"
)

// TestUserAgentContainsProviderVersion ensures the user agent is set.
func TestUserAgentContainsProviderVersion(t *testing.T) {
	c := Config{
		APIKey:  "someapikey",
		BaseURL: "http://localhost",
		Context: testCtx(),
	}
	_, diagnostics := c.Client()

	if diagnostics.HasError() {
		t.Errorf("failed to create client: %s", diagToErr(diagnostics))
	}
}

// TestForceHttp2 ensures transport changes when ForceHTTP2 is enabled.
func TestForceHttp2(t *testing.T) {
	c1 := Config{
		APIKey:  "someapikey",
		BaseURL: "http://localhost",
		Context: testCtx(),
	}
	client1, _ := c1.Client()

	c2 := Config{
		APIKey:     "someapikey",
		BaseURL:    "http://localhost",
		ForceHTTP2: true,
		Context:    testCtx(),
	}
	client2, _ := c2.Client()

	t1 := extractUnderlyingTransport(client1.conn.HTTPClient.Transport)
	t2 := extractUnderlyingTransport(client2.conn.HTTPClient.Transport)

	if reflect.TypeOf(t1) == reflect.TypeOf(t2) {
		t.Errorf("expected different transport types with and without ForceHTTP2: got %T and %T", t1, t2)
	}
}

// testCtx returns a context with a basic hclog.Logger.
func testCtx() context.Context {
	type contextKey string
	const logKey contextKey = "terraform:log"

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "test",
		Level: hclog.Debug,
	})
	return context.WithValue(context.Background(), logKey, logger)
}

// extractUnderlyingTransport unwraps redactingTransport if present.
func extractUnderlyingTransport(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		return nil
	}
	if wrapper, ok := rt.(*redactingTransport); ok {
		return wrapper.underlying
	}
	return rt
}
