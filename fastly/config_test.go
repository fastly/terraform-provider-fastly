package fastly

import (
	"reflect"
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

func TestForceHttp2(t *testing.T) {
	c1 := Config{
		ApiKey:  "someapikey",
		BaseURL: "http://localhost",
	}
	client1, _ := c1.Client()

	c2 := Config{
		ApiKey:     "someapikey",
		BaseURL:    "http://localhost",
		ForceHttp2: true,
	}
	client2, _ := c2.Client()

	tv1 := reflect.ValueOf(client1.conn.HTTPClient.Transport).Elem()
	// <http.Transport Value>
	ts1 := tv1.FieldByName("transport").Elem().Elem().String()

	tv2 := reflect.ValueOf(client2.conn.HTTPClient.Transport).Elem()
	// <http2.Transport Value>
	ts2 := tv2.FieldByName("transport").Elem().Elem().String()

	if ts1 == ts2 {
		t.Errorf("Failed to create client with force_http2: %#v, %#v", ts1, ts2)
	}
}
