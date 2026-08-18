package logginghttps

import (
	"context"
	"fmt"
	"strings"
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
)

// Test helpers

func defaultNestedModel() NestedModel {
	return NestedModel{
		commonModel:       defaultCommonModel(),
		Format:            types.StringValue(constants.LoggingHTTPSDefaultFormat),
		FormatVersion:     types.Int64Value(DefaultFormatVersion),
		Placement:         types.StringNull(),
		ResponseCondition: types.StringValue(DefaultResponseCondition),
	}
}

func defaultCommonModel() commonModel {
	return commonModel{
		Name:              types.StringValue(""),
		URL:               types.StringValue(""),
		ContentType:       types.StringValue(DefaultContentType),
		CompressionCodec:  types.StringValue(DefaultCompressionCodec),
		GzipLevel:         types.Int64Value(DefaultGzipLevel),
		HeaderName:        types.StringValue(DefaultHeaderName),
		HeaderValue:       types.StringValue(DefaultHeaderValue),
		JSONFormat:        types.StringValue(DefaultJSONFormat),
		MessageType:       types.StringValue(DefaultMessageType),
		Method:            types.StringValue(DefaultMethod),
		Period:            types.Int64Value(DefaultPeriod),
		ProcessingRegion:  types.StringValue(DefaultProcessingRegion),
		RequestMaxBytes:   types.Int64Value(DefaultRequestMaxBytes),
		RequestMaxEntries: types.Int64Value(DefaultRequestMaxEntries),
		TLS:               NewTLSObject(types.StringValue(""), types.StringValue(""), types.StringValue(""), types.StringValue(DefaultTLSHostname)),
	}
}

func fullNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-https")
	m.URL = types.StringValue("https://https.example.com/logs")
	m.ContentType = types.StringValue("application/json")
	m.CompressionCodec = types.StringValue("")
	m.GzipLevel = types.Int64Value(6)
	m.HeaderName = types.StringValue("X-Custom")
	m.HeaderValue = types.StringValue("custom-value")
	m.JSONFormat = types.StringValue("1")
	m.MessageType = types.StringValue("loggly")
	m.Method = types.StringValue("PUT")
	m.Period = types.Int64Value(10)
	m.ProcessingRegion = types.StringValue("eu")
	m.RequestMaxBytes = types.Int64Value(1000000)
	m.RequestMaxEntries = types.Int64Value(1000)
	m.TLS = NewTLSObject(
		types.StringValue("ca-cert"),
		types.StringValue("client-cert"),
		types.StringValue("client-key"),
		types.StringValue("https.example.com"),
	)
	m.Format = types.StringValue("%h %l %u")
	m.FormatVersion = types.Int64Value(1)
	m.Placement = types.StringValue("none")
	m.ResponseCondition = types.StringValue("response-condition-1")
	return m
}

func minimalNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-https")
	m.URL = types.StringValue("https://https.example.com/logs")
	return m
}

func fullComputeNestedModel() ComputeNestedModel {
	return ComputeNestedModel{commonModel: fullNestedModel().commonModel}
}

// Tests for flatten.go

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name     string
		api      *fastly.HTTPS
		expected NestedModel
	}{
		{
			name:     "nil returns empty model",
			api:      nil,
			expected: NestedModel{},
		},
		{
			name: "only required fields uses defaults",
			api: &fastly.HTTPS{
				Name: new("test-https"),
				URL:  new("https://https.example.com/logs"),
			},
			expected: minimalNestedModel(),
		},
		{
			name: "all fields populated",
			api: &fastly.HTTPS{
				Name:              new("test-https"),
				URL:               new("https://https.example.com/logs"),
				ContentType:       new("application/json"),
				CompressionCodec:  new(""),
				GzipLevel:         new(6),
				HeaderName:        new("X-Custom"),
				HeaderValue:       new("custom-value"),
				JSONFormat:        new("1"),
				MessageType:       new("loggly"),
				Method:            new("PUT"),
				Period:            new(10),
				ProcessingRegion:  new("eu"),
				RequestMaxBytes:   new(1000000),
				RequestMaxEntries: new(1000),
				TLSCACert:         new("ca-cert"),
				TLSClientCert:     new("client-cert"),
				TLSClientKey:      new("client-key"),
				TLSHostname:       new("https.example.com"),
				Format:            new("%h %l %u"),
				FormatVersion:     new(1),
				Placement:         new("none"),
				ResponseCondition: new("response-condition-1"),
			},
			expected: fullNestedModel(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenToNestedModel(tt.api)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenToComputeNestedModel(t *testing.T) {
	api := &fastly.HTTPS{
		Name:              new("test-https"),
		URL:               new("https://https.example.com/logs"),
		ContentType:       new("application/json"),
		GzipLevel:         new(6),
		HeaderName:        new("X-Custom"),
		HeaderValue:       new("custom-value"),
		JSONFormat:        new("1"),
		MessageType:       new("loggly"),
		Method:            new("PUT"),
		Period:            new(10),
		ProcessingRegion:  new("eu"),
		RequestMaxBytes:   new(1000000),
		RequestMaxEntries: new(1000),
		TLSCACert:         new("ca-cert"),
		TLSClientCert:     new("client-cert"),
		TLSClientKey:      new("client-key"),
		TLSHostname:       new("https.example.com"),
		// VCL-only fields must be ignored by the Compute flatten.
		Format:            new("%h %l %u"),
		FormatVersion:     new(1),
		Placement:         new("none"),
		ResponseCondition: new("response-condition-1"),
	}

	result := FlattenToComputeNestedModel(api)
	assert.Equal(t, fullComputeNestedModel(), result)
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		api      *fastly.HTTPS
		validate func(t *testing.T, m *Model)
	}{
		{
			name: "nil leaves model untouched",
			api:  nil,
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.String{}, m.ID)
				assert.Equal(t, types.String{}, m.Service)
				assert.Equal(t, types.Int64{}, m.Version)
			},
		},
		{
			name: "service metadata builds composite ID",
			api: &fastly.HTTPS{
				ServiceID:      new("service-123"),
				ServiceVersion: new(5),
				Name:           new("test-https"),
				URL:            new("https://https.example.com/logs"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service-123-5-test-https"), m.ID)
				assert.Equal(t, types.StringValue("service-123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("test-https"), m.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Model{}
			flatten(ctx, tt.api, m)
			tt.validate(t, m)
		})
	}
}

func TestPreserveGzipSentinel(t *testing.T) {
	tests := []struct {
		name     string
		remote   NestedModel
		desired  NestedModel
		expected types.Int64
	}{
		{
			name: "desired unset restores sentinel over API auto-managed value",
			remote: func() NestedModel {
				m := minimalNestedModel()
				m.GzipLevel = types.Int64Value(3)
				return m
			}(),
			desired:  minimalNestedModel(),
			expected: types.Int64Value(DefaultGzipLevel),
		},
		{
			name: "desired set keeps the API value",
			remote: func() NestedModel {
				m := minimalNestedModel()
				m.GzipLevel = types.Int64Value(6)
				return m
			}(),
			desired: func() NestedModel {
				m := minimalNestedModel()
				m.GzipLevel = types.Int64Value(6)
				return m
			}(),
			expected: types.Int64Value(6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.remote
			preserveGzipSentinel(&m, tt.desired)
			assert.Equal(t, tt.expected, m.GzipLevel)
		})
	}
}

func TestInferGzipSentinelOnImport(t *testing.T) {
	tests := []struct {
		name     string
		m        commonModel
		expected types.Int64
	}{
		{
			name: "no codec, gzip_level 0 treated as unconfigured",
			m: commonModel{
				CompressionCodec: types.StringValue(""),
				GzipLevel:        types.Int64Value(0),
			},
			expected: types.Int64Value(DefaultGzipLevel),
		},
		{
			name: "codec set, gzip_level 0 is left alone",
			m: commonModel{
				CompressionCodec: types.StringValue("zstd"),
				GzipLevel:        types.Int64Value(0),
			},
			expected: types.Int64Value(0),
		},
		{
			name: "non-zero gzip_level is left alone",
			m: commonModel{
				CompressionCodec: types.StringValue(""),
				GzipLevel:        types.Int64Value(5),
			},
			expected: types.Int64Value(5),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			inferGzipSentinelOnImport(&m)
			assert.Equal(t, tt.expected, m.GzipLevel)
		})
	}
}

func TestPreserveGzipSentinelList(t *testing.T) {
	read := []NestedModel{
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("a")
			m.GzipLevel = types.Int64Value(3)
			return m
		}(),
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("b")
			m.GzipLevel = types.Int64Value(6)
			return m
		}(),
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("c")
			m.GzipLevel = types.Int64Value(0)
			return m
		}(),
	}
	desired := []NestedModel{
		func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
		func() NestedModel {
			m := minimalNestedModel()
			m.Name = types.StringValue("b")
			m.GzipLevel = types.Int64Value(6)
			return m
		}(),
		// "c" has no entry in desired, simulating a freshly imported or
		// undiscovered-in-config endpoint.
	}

	preserveGzipSentinelList(read, desired)

	assert.Equal(t, types.Int64Value(DefaultGzipLevel), read[2].GzipLevel, "unmatched entry falls back to the import heuristic")
	assert.Equal(t, types.Int64Value(DefaultGzipLevel), read[0].GzipLevel, "unmatched-by-desired sentinel should be restored")
	assert.Equal(t, types.Int64Value(6), read[1].GzipLevel, "explicitly configured value should be preserved")
}

// Tests for expand.go

func TestBuildCreateInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		model     NestedModel
		validate  func(t *testing.T, input *fastly.CreateHTTPSInput)
	}{
		{
			name:      "minimal model",
			serviceID: "service-123",
			version:   5,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateHTTPSInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test-https", *input.Name)
				assert.Equal(t, "https://https.example.com/logs", *input.URL)
				assert.Equal(t, constants.LoggingHTTPSDefaultFormat, *input.Format)
				assert.Nil(t, input.GzipLevel, "unset gzip_level must be omitted, not sent as -1")
				assert.Nil(t, input.TLSCACert, "empty tls.ca_cert must be omitted, not sent as \"\"")
				assert.Nil(t, input.RequestMaxBytes, "zero request_max_bytes must be omitted on create")
				assert.Nil(t, input.Placement, "unset placement must not be sent as \"none\" — the API treats them differently")
			},
		},
		{
			name:      "fully populated model",
			serviceID: "service-456",
			version:   10,
			model:     fullNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateHTTPSInput) {
				assert.Equal(t, "test-https", *input.Name)
				assert.Equal(t, "https://https.example.com/logs", *input.URL)
				assert.Equal(t, "application/json", *input.ContentType)
				assert.Equal(t, 6, *input.GzipLevel)
				assert.Nil(t, input.CompressionCodec, "empty compression_codec must be omitted on create")
				assert.Equal(t, "X-Custom", *input.HeaderName)
				assert.Equal(t, "custom-value", *input.HeaderValue)
				assert.Equal(t, "1", *input.JSONFormat)
				assert.Equal(t, "loggly", *input.MessageType)
				assert.Equal(t, "PUT", *input.Method)
				assert.Equal(t, 10, *input.Period)
				assert.Equal(t, "eu", *input.ProcessingRegion)
				assert.Equal(t, 1000000, *input.RequestMaxBytes)
				assert.Equal(t, 1000, *input.RequestMaxEntries)
				assert.Equal(t, "ca-cert", *input.TLSCACert)
				assert.Equal(t, "client-cert", *input.TLSClientCert)
				assert.Equal(t, "client-key", *input.TLSClientKey)
				assert.Equal(t, "https.example.com", *input.TLSHostname)
				assert.Equal(t, "%h %l %u", *input.Format)
				assert.Equal(t, 1, *input.FormatVersion)
				assert.Equal(t, "none", *input.Placement)
				assert.Equal(t, "response-condition-1", *input.ResponseCondition)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := BuildCreateInput(tt.serviceID, tt.version, tt.model)
			tt.validate(t, input)
		})
	}
}

func TestBuildComputeCreateInput(t *testing.T) {
	input := BuildComputeCreateInput("service-456", 10, fullComputeNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-https", *input.Name)
	assert.Equal(t, "https://https.example.com/logs", *input.URL)
	assert.Equal(t, "eu", *input.ProcessingRegion)
	assert.Equal(t, 1000000, *input.RequestMaxBytes)
	assert.Nil(t, input.Format, "VCL-only fields must never be set for Compute")
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestBuildUpdateInput(t *testing.T) {
	input := BuildUpdateInput("service-456", 10, fullNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-https", input.Name)
	assert.Equal(t, "test-https", *input.NewName)
	assert.Equal(t, "https://https.example.com/logs", *input.URL)
	assert.Equal(t, "application/json", *input.ContentType)
	assert.Equal(t, 6, *input.GzipLevel)
	assert.Equal(t, "X-Custom", *input.HeaderName)
	assert.Equal(t, "custom-value", *input.HeaderValue)
	assert.Equal(t, "ca-cert", *input.TLSCACert)
	assert.Equal(t, "client-key", *input.TLSClientKey)
	assert.Equal(t, "eu", *input.ProcessingRegion)
	assert.Equal(t, 1000000, *input.RequestMaxBytes)
	assert.Equal(t, 1000, *input.RequestMaxEntries)
	assert.Equal(t, "%h %l %u", *input.Format)
	assert.Equal(t, 1, *input.FormatVersion)
	assert.Equal(t, fastly.NewNullable("none"), input.Placement)
	assert.Equal(t, "response-condition-1", *input.ResponseCondition)
}

// TestBuildUpdateInputClearsClearableFields verifies that fields defaulting to
// an empty/zero value are always sent as a concrete value on update — even
// when empty/zero — so clearing them actually reaches the API rather than
// being omitted (which would leave a previously-set value in place). placement
// is cleared the same way, but as an explicit JSON null rather than an empty
// string — see BuildUpdateInput.
func TestBuildUpdateInputClearsClearableFields(t *testing.T) {
	input := BuildUpdateInput("service-1", 1, minimalNestedModel())

	assert.NotNil(t, input.ContentType, "content_type must be sent even when empty")
	assert.Equal(t, "", *input.ContentType)
	assert.NotNil(t, input.HeaderName, "header_name must be sent even when empty")
	assert.Equal(t, "", *input.HeaderName)
	assert.NotNil(t, input.HeaderValue, "header_value must be sent even when empty")
	assert.Equal(t, "", *input.HeaderValue)
	assert.NotNil(t, input.CompressionCodec, "compression_codec must be sent even when empty")
	assert.Equal(t, "", *input.CompressionCodec)
	assert.NotNil(t, input.TLSCACert, "tls.ca_cert must be sent even when empty")
	assert.Equal(t, "", *input.TLSCACert)
	assert.NotNil(t, input.RequestMaxBytes, "request_max_bytes must be sent even when zero")
	assert.Equal(t, 0, *input.RequestMaxBytes)
	assert.NotNil(t, input.ResponseCondition, "response_condition must be sent even when empty")
	assert.Equal(t, "", *input.ResponseCondition)
	assert.Nil(t, input.GzipLevel, "unset gzip_level must still be omitted on update")
	assert.NotNil(t, input.Placement, "unset placement must be sent as an explicit null, not omitted (omitting leaves a previously-set \"none\" in place)")
	assert.Equal(t, fastly.NullValue[string](), input.Placement)
}

func TestBuildComputeUpdateInput(t *testing.T) {
	input := BuildComputeUpdateInput("service-456", 10, fullComputeNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-https", input.Name)
	assert.Equal(t, "test-https", *input.NewName)
	assert.Equal(t, "eu", *input.ProcessingRegion)
	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestClearVCLOnlyCreateFields(t *testing.T) {
	input := &fastly.CreateHTTPSInput{
		Format:            new("some-format"),
		FormatVersion:     new(2),
		Placement:         new("none"),
		ResponseCondition: new("cond"),
	}

	ClearVCLOnlyCreateFields(input)

	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestClearVCLOnlyUpdateFields(t *testing.T) {
	input := &fastly.UpdateHTTPSInput{
		Format:            new("some-format"),
		FormatVersion:     new(2),
		Placement:         fastly.NewNullable("none"),
		ResponseCondition: new("cond"),
	}

	ClearVCLOnlyUpdateFields(input)

	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

// TestResetVCLOnlyToDefaults covers the Compute read-back path. On a Compute
// service the VCL-only fields are never sent, so the API reports its own
// server-side values — a different default format, and placement forced to
// "none". Adopting those breaks consistency-after-apply, so they must be reset
// to exactly the values a plan produces.
func TestResetVCLOnlyToDefaults(t *testing.T) {
	m := FlattenToNestedModel(&fastly.HTTPS{
		Name:              new("test-https"),
		URL:               new("https://https.example.com/logs"),
		ProcessingRegion:  new("none"),
		Format:            new("{\n  \"time\": 1\n}\n"),
		FormatVersion:     new(1),
		Placement:         new("none"),
		ResponseCondition: new("some-condition"),
	})

	ResetVCLOnlyToDefaults(&m)

	assert.Equal(t, constants.LoggingHTTPSDefaultFormat, m.Format.ValueString())
	assert.Equal(t, int64(DefaultFormatVersion), m.FormatVersion.ValueInt64())
	assert.True(t, m.Placement.IsNull(), "placement must go back to unset, not the API's forced \"none\"")
	assert.Equal(t, DefaultResponseCondition, m.ResponseCondition.ValueString())

	// Non-VCL-only fields must survive untouched.
	assert.Equal(t, "test-https", m.Name.ValueString())
	assert.Equal(t, "none", m.ProcessingRegion.ValueString())
}

// TestResetVCLOnlyToDefaultsMatchesPlannedDefaults ties the reset to the schema
// itself: the values it writes must equal the schema's declared defaults, or
// Create/Update would still disagree with the plan.
func TestResetVCLOnlyToDefaultsMatchesPlannedDefaults(t *testing.T) {
	var m NestedModel
	ResetVCLOnlyToDefaults(&m)

	attrs := CommonAttributes()

	format := attrs["format"].(schema.StringAttribute)
	var fResp defaults.StringResponse
	format.Default.DefaultString(context.Background(), defaults.StringRequest{}, &fResp)
	assert.Equal(t, fResp.PlanValue, m.Format, "format must match its schema default")

	formatVersion := attrs["format_version"].(schema.Int64Attribute)
	var fvResp defaults.Int64Response
	formatVersion.Default.DefaultInt64(context.Background(), defaults.Int64Request{}, &fvResp)
	assert.Equal(t, fvResp.PlanValue, m.FormatVersion, "format_version must match its schema default")

	responseCondition := attrs["response_condition"].(schema.StringAttribute)
	var rcResp defaults.StringResponse
	responseCondition.Default.DefaultString(context.Background(), defaults.StringRequest{}, &rcResp)
	assert.Equal(t, rcResp.PlanValue, m.ResponseCondition, "response_condition must match its schema default")

	// placement is Optional-only with no Default, so an absent config value plans
	// as null — the reset has to produce null, not "".
	assert.Nil(t, attrs["placement"].(schema.StringAttribute).Default)
	assert.True(t, m.Placement.IsNull())
}

// Tests for schema.go

func TestModelsEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        NestedModel
		b        NestedModel
		expected bool
	}{
		{
			name:     "identical models",
			a:        fullNestedModel(),
			b:        fullNestedModel(),
			expected: true,
		},
		{
			name:     "default models",
			a:        defaultNestedModel(),
			b:        defaultNestedModel(),
			expected: true,
		},
		{
			name: "different tls client_key",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.TLS = NewTLSObject(types.StringValue(""), types.StringValue(""), types.StringValue("key-1"), types.StringValue(""))
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.TLS = NewTLSObject(types.StringValue(""), types.StringValue(""), types.StringValue("key-2"), types.StringValue(""))
				return m
			}(),
			expected: false,
		},
		{
			name: "different format only affects NestedModel equality",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Format = types.StringValue("format-a")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Format = types.StringValue("format-b")
				return m
			}(),
			expected: false,
		},
		{
			name: "unset placement differs from explicit none",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Placement = types.StringNull()
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Placement = types.StringValue("none")
				return m
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.a.ModelsEqual(tt.b))
		})
	}
}

func TestComputeModelsEqual(t *testing.T) {
	a := fullComputeNestedModel()
	b := fullComputeNestedModel()
	assert.True(t, a.ModelsEqual(b))

	b.ProcessingRegion = types.StringValue("us")
	assert.False(t, a.ModelsEqual(b))
}

// TestComputeModelsEqualIgnoresVCLOnlyFields verifies that a Compute endpoint
// whose remote state carries VCL-only fields still compares equal to the desired
// Compute model — otherwise ComputeReconcile would issue a pointless update on
// every apply.
func TestComputeModelsEqualIgnoresVCLOnlyFields(t *testing.T) {
	desired := fullComputeNestedModel()

	remote := &fastly.HTTPS{
		Name:              new("test-https"),
		URL:               new("https://https.example.com/logs"),
		ContentType:       new("application/json"),
		GzipLevel:         new(6),
		HeaderName:        new("X-Custom"),
		HeaderValue:       new("custom-value"),
		JSONFormat:        new("1"),
		MessageType:       new("loggly"),
		Method:            new("PUT"),
		Period:            new(10),
		ProcessingRegion:  new("eu"),
		RequestMaxBytes:   new(1000000),
		RequestMaxEntries: new(1000),
		TLSCACert:         new("ca-cert"),
		TLSClientCert:     new("client-cert"),
		TLSClientKey:      new("client-key"),
		TLSHostname:       new("https.example.com"),
		Format:            new("something-else-entirely"),
		FormatVersion:     new(1),
		Placement:         new("none"),
		ResponseCondition: new("some-condition"),
	}

	assert.True(t, desired.ModelsEqual(FlattenToComputeNestedModel(remote)))
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []NestedModel
		b        []NestedModel
		expected bool
	}{
		{
			name:     "both empty",
			a:        []NestedModel{},
			b:        []NestedModel{},
			expected: true,
		},
		{
			name: "different order but same content matches by name",
			a: []NestedModel{
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("b"); return m }(),
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
			},
			b: []NestedModel{
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("b"); return m }(),
			},
			expected: true,
		},
		{
			name: "different content",
			a: []NestedModel{
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
			},
			b: []NestedModel{
				func() NestedModel {
					m := minimalNestedModel()
					m.Name = types.StringValue("a")
					m.ProcessingRegion = types.StringValue("eu")
					return m
				}(),
			},
			expected: false,
		},
		{
			name: "different length",
			a: []NestedModel{
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
			},
			b: []NestedModel{
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }(),
				func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("b"); return m }(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Equal(tt.a, tt.b))
		})
	}
}

func TestComputeEqual(t *testing.T) {
	a := []ComputeNestedModel{fullComputeNestedModel()}
	b := []ComputeNestedModel{fullComputeNestedModel()}
	assert.True(t, ComputeEqual(a, b))

	b[0].ProcessingRegion = types.StringValue("us")
	assert.False(t, ComputeEqual(a, b))
}

func TestMatchOrder(t *testing.T) {
	itemA := func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("a"); return m }()
	itemB := func() NestedModel { m := minimalNestedModel(); m.Name = types.StringValue("b"); return m }()
	items := []NestedModel{itemB, itemA}

	orderA := minimalNestedModel()
	orderA.Name = types.StringValue("a")
	orderB := minimalNestedModel()
	orderB.Name = types.StringValue("b")
	order := []NestedModel{orderA, orderB}

	result := MatchOrder(items, order)

	assert.Len(t, result, 2)
	assert.Equal(t, "a", result[0].Name.ValueString())
	assert.Equal(t, "b", result[1].Name.ValueString())
}

func TestComputeMatchOrder(t *testing.T) {
	mk := func(name string) ComputeNestedModel {
		m := fullComputeNestedModel()
		m.Name = types.StringValue(name)
		return m
	}
	items := []ComputeNestedModel{mk("b"), mk("a")}
	order := []ComputeNestedModel{mk("a"), mk("b")}

	result := ComputeMatchOrder(items, order)

	assert.Len(t, result, 2)
	assert.Equal(t, "a", result[0].Name.ValueString())
	assert.Equal(t, "b", result[1].Name.ValueString())
}

// TestComputeAttributesOmitsVCLOnly locks in that the Compute nested block
// schema does not expose the VCL-only attributes, which is what makes
// `logging_https { format = ... }` inside fastly_service_compute_auto fail at
// plan time with Terraform's own "Unsupported argument" error.
func TestComputeAttributesOmitsVCLOnly(t *testing.T) {
	compute := ComputeAttributes()
	common := CommonAttributes()

	for _, name := range []string{"format", "format_version", "placement", "response_condition"} {
		assert.NotContains(t, compute, name)
		assert.Contains(t, common, name)
	}

	for _, name := range []string{"name", "url", "content_type", "compression_codec", "gzip_level", "header_name", "header_value", "json_format", "message_type", "method", "period", "processing_region", "request_max_bytes", "request_max_entries", "tls"} {
		assert.Contains(t, compute, name)
		assert.Contains(t, common, name)
	}

	// The tls.* fields are nested, never top-level attributes.
	for _, name := range []string{"ca_cert", "client_cert", "client_key", "hostname"} {
		assert.NotContains(t, compute, name)
		assert.NotContains(t, common, name)
	}
}

// TestTLSAttribute locks in the tls object shape: Optional+Computed, defaulted
// to empty strings (no environment variables exist for these in the live
// provider), with client_key marked Sensitive since it is credential material
// used to authenticate this endpoint via mutual TLS.
func TestTLSAttribute(t *testing.T) {
	tls, ok := ComputeAttributes()["tls"].(schema.SingleNestedAttribute)
	require.True(t, ok, "tls must be a SingleNestedAttribute")

	assert.True(t, tls.Optional)
	assert.True(t, tls.Computed)
	assert.NotNil(t, tls.Default)

	clientKey, ok := tls.Attributes["client_key"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, clientKey.Sensitive, "the TLS client private key must never be rendered in plan output")

	for _, name := range []string{"ca_cert", "client_cert", "hostname"} {
		attr, ok := tls.Attributes[name].(schema.StringAttribute)
		require.True(t, ok, "tls.%s must be a StringAttribute", name)
		assert.False(t, attr.Sensitive, "tls.%s is not credential material", name)
	}

	assert.Len(t, tls.Attributes, 4)
}

// TestTLSAccessors mirrors the tls object accessor handling in loggingsplunk:
// null/unknown object, absent attribute — an empty string is the safe answer
// rather than a panic.
func TestTLSAccessors(t *testing.T) {
	full := fullNestedModel()
	assert.Equal(t, "ca-cert", full.TLSCACert().ValueString())
	assert.Equal(t, "client-cert", full.TLSClientCert().ValueString())
	assert.Equal(t, "client-key", full.TLSClientKey().ValueString())
	assert.Equal(t, "https.example.com", full.TLSHostname().ValueString())

	tests := map[string]types.Object{
		"null object":    types.ObjectNull(tlsAttributeTypes),
		"unknown object": types.ObjectUnknown(tlsAttributeTypes),
	}
	for name, obj := range tests {
		t.Run(name, func(t *testing.T) {
			m := commonModel{TLS: obj}
			assert.Equal(t, "", m.TLSCACert().ValueString())
			assert.Equal(t, "", m.TLSClientCert().ValueString())
			assert.Equal(t, "", m.TLSClientKey().ValueString())
			assert.Equal(t, "", m.TLSHostname().ValueString())
		})
	}
}

// TestSchemaValidators pins the accepted values for the schema's validators, so
// a change to an enum member or a bound is a test failure rather than a
// surprise at plan time.
func TestSchemaValidators(t *testing.T) {
	attrs := CommonAttributes()

	stringCases := []struct {
		name  string
		attr  string
		value string
		valid bool
	}{
		{"compression_codec zstd", "compression_codec", "zstd", true},
		{"compression_codec snappy", "compression_codec", "snappy", true},
		{"compression_codec gzip", "compression_codec", "gzip", true},
		{"compression_codec rejects unknown", "compression_codec", "bzip2", false},
		{"json_format 0", "json_format", "0", true},
		{"json_format 1", "json_format", "1", true},
		{"json_format 2", "json_format", "2", true},
		{"json_format rejects other values", "json_format", "3", false},
		{"message_type classic", "message_type", "classic", true},
		{"message_type loggly", "message_type", "loggly", true},
		{"message_type logplex", "message_type", "logplex", true},
		{"message_type blank", "message_type", "blank", true},
		{"message_type rejects unknown", "message_type", "other", false},
		{"method POST", "method", "POST", true},
		{"method PUT", "method", "PUT", true},
		{"method rejects lowercase", "method", "post", false},
		{"processing_region none", "processing_region", "none", true},
		{"processing_region us", "processing_region", "us", true},
		{"processing_region eu", "processing_region", "eu", true},
		{"processing_region rejects wrong case", "processing_region", "US", false},
		{"placement none", "placement", "none", true},
		{"placement rejects other VCL subroutines", "placement", "waf_debug", false},
		{"placement rejects empty", "placement", "", false},
		{"format at max length", "format", strings.Repeat("x", maximumFormatLength), true},
		{"format over max length", "format", strings.Repeat("x", maximumFormatLength+1), false},
		{"url accepts https", "url", "https://example.com/logs", true},
		{"url rejects http", "url", "http://example.com/logs", false},
		{"url rejects scheme-less", "url", "example.com/logs", false},
	}
	for _, tt := range stringCases {
		t.Run(tt.name, func(t *testing.T) {
			a := attrs[tt.attr].(schema.StringAttribute)
			require.NotEmpty(t, a.Validators)
			resp := &validator.StringResponse{}
			for _, v := range a.Validators {
				v.ValidateString(context.Background(),
					validator.StringRequest{ConfigValue: types.StringValue(tt.value)}, resp)
			}
			assert.Equal(t, tt.valid, !resp.Diagnostics.HasError())
		})
	}

	formatVersionCases := []struct {
		value int64
		valid bool
	}{{0, false}, {1, true}, {2, true}, {3, false}}
	for _, tt := range formatVersionCases {
		t.Run(fmt.Sprintf("format_version %d", tt.value), func(t *testing.T) {
			a := attrs["format_version"].(schema.Int64Attribute)
			require.Len(t, a.Validators, 1)
			resp := &validator.Int64Response{}
			a.Validators[0].ValidateInt64(context.Background(),
				validator.Int64Request{ConfigValue: types.Int64Value(tt.value)}, resp)
			assert.Equal(t, tt.valid, !resp.Diagnostics.HasError())
		})
	}

	gzipLevelCases := []struct {
		value int64
		valid bool
	}{{-1, false}, {0, true}, {9, true}, {10, false}}
	for _, tt := range gzipLevelCases {
		t.Run(fmt.Sprintf("gzip_level %d", tt.value), func(t *testing.T) {
			a := attrs["gzip_level"].(schema.Int64Attribute)
			require.Len(t, a.Validators, 2)
			resp := &validator.Int64Response{}
			// Only the range validator (index 1); the conflict validator needs a
			// config object and is covered separately in TestGzipLevelCodecConflict.
			a.Validators[1].ValidateInt64(context.Background(),
				validator.Int64Request{ConfigValue: types.Int64Value(tt.value)}, resp)
			assert.Equal(t, tt.valid, !resp.Diagnostics.HasError())
		})
	}

	// name and response_condition accept any string; assert that rather than
	// leaving it implicit.
	assert.Empty(t, attrs["name"].(schema.StringAttribute).Validators)
	assert.Empty(t, attrs["response_condition"].(schema.StringAttribute).Validators)
}
