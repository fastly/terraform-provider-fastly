package loggingsumologic

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		name  string
		value types.String
		valid bool
	}{
		{"null is skipped", types.StringNull(), true},
		{"unknown is skipped", types.StringUnknown(), true},
		{"valid https URL", types.StringValue("https://collectors.sumologic.com/receiver/v1/http/token"), true},
		{"valid http URL", types.StringValue("http://collectors.sumologic.com/receiver/v1/http/token"), true},
		{"empty string", types.StringValue(""), false},
		{"missing scheme", types.StringValue("collectors.sumologic.com/receiver/v1/http/token"), false},
		{"non-http(s) scheme", types.StringValue("ftp://collectors.sumologic.com/receiver/v1/http/token"), false},
		{"missing host", types.StringValue("https://"), false},
		{"not a URL at all", types.StringValue("::not a url::"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			isURL{}.ValidateString(context.Background(), validator.StringRequest{ConfigValue: tt.value}, resp)
			assert.Equal(t, tt.valid, !resp.Diagnostics.HasError())
		})
	}
}
