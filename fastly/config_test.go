package fastly

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestUserAgentContainsProviderVersion(t *testing.T) {
	c := Config{
		APIKey:  "someapikey",
		BaseURL: "http://localhost",
	}
	_, diagnostics := c.Client()

	if diagnostics.HasError() {
		t.Errorf("failed to create client: %s", diagToErr(diagnostics))
	}
}

func TestForceHttp2(t *testing.T) {
	c1 := Config{
		APIKey:  "someapikey",
		BaseURL: "http://localhost",
	}
	client1, _ := c1.Client()

	c2 := Config{
		APIKey:     "someapikey",
		BaseURL:    "http://localhost",
		ForceHTTP2: true,
	}
	client2, _ := c2.Client()

	tv1 := reflect.ValueOf(client1.conn.HTTPClient.Transport).Elem()
	// http.Transport
	ts1 := reflect.Indirect(tv1.FieldByName("transport").Elem()).Type().String()

	tv2 := reflect.ValueOf(client2.conn.HTTPClient.Transport).Elem()
	// http2.Transport
	ts2 := reflect.Indirect(tv2.FieldByName("transport").Elem()).Type().String()

	if ts1 == ts2 {
		t.Errorf("failed to create client with force_http2: %#v, %#v", ts1, ts2)
	}
}

func TestDisplaySensitiveFields(t *testing.T) {
	c1 := Config{
		APIKey:                 "someapikey",
		BaseURL:                "http://localhost",
		DisplaySensitiveFields: true,
	}
	_, _ = c1.Client()
	if !DisplaySensitiveFields {
		t.Errorf("expected !DisplaySensitiveFields to equal true")
	}
}

func TestSecretKeySchemaDisplayFunc(t *testing.T) {
	//set global sensitive var to true to display sensitive values
	c1 := Config{
		APIKey:                 "someapikey",
		BaseURL:                "http://localhost",
		DisplaySensitiveFields: true,
	}
	_, _ = c1.Client()
	computeAttributes := ServiceMetadata{ServiceTypeCompute}
	v := NewServiceLoggingGooglePubSub(computeAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_googlepubsub"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Expect secret_key to not be sensitive
	if loggingResourceSchema["secret_key"].Sensitive {
		t.Fatalf("Expected secret_key not to be marked as a Sensitive value")
	}
}
