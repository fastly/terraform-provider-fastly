package loggingsumologic

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
		Format:            types.StringValue(constants.LoggingSumologicDefaultFormat),
		FormatVersion:     types.Int64Value(DefaultFormatVersion),
		Placement:         types.StringNull(),
		ResponseCondition: types.StringValue(DefaultResponseCondition),
	}
}

func defaultCommonModel() commonModel {
	return commonModel{
		Name:             types.StringValue(""),
		URL:              types.StringValue(""),
		MessageType:      types.StringValue(DefaultMessageType),
		ProcessingRegion: types.StringValue(DefaultProcessingRegion),
	}
}

func fullNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-sumologic")
	m.URL = types.StringValue("https://foo.sumologic.com")
	m.MessageType = types.StringValue("loggly")
	m.ProcessingRegion = types.StringValue("eu")
	m.Format = types.StringValue("%h %l %u")
	m.FormatVersion = types.Int64Value(1)
	m.Placement = types.StringValue("none")
	m.ResponseCondition = types.StringValue("response-condition-1")
	return m
}

func minimalNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-sumologic")
	m.URL = types.StringValue("https://foo.sumologic.com")
	return m
}

func fullComputeNestedModel() ComputeNestedModel {
	return ComputeNestedModel{commonModel: fullNestedModel().commonModel}
}

// Tests for flatten.go

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name     string
		api      *fastly.Sumologic
		expected NestedModel
	}{
		{
			name:     "nil returns empty model",
			api:      nil,
			expected: NestedModel{},
		},
		{
			name: "only required fields uses defaults",
			api: &fastly.Sumologic{
				Name: new("test-sumologic"),
				URL:  new("https://foo.sumologic.com"),
			},
			expected: minimalNestedModel(),
		},
		{
			name: "all fields populated",
			api: &fastly.Sumologic{
				Name:              new("test-sumologic"),
				URL:               new("https://foo.sumologic.com"),
				MessageType:       new("loggly"),
				ProcessingRegion:  new("eu"),
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
	api := &fastly.Sumologic{
		Name:             new("test-sumologic"),
		URL:              new("https://foo.sumologic.com"),
		MessageType:      new("loggly"),
		ProcessingRegion: new("eu"),
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
		api      *fastly.Sumologic
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
			api: &fastly.Sumologic{
				ServiceID:      new("service-123"),
				ServiceVersion: new(5),
				Name:           new("test-sumologic"),
				URL:            new("https://foo.sumologic.com"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service-123-5-test-sumologic"), m.ID)
				assert.Equal(t, types.StringValue("service-123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("test-sumologic"), m.Name)
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

// Tests for expand.go

func TestBuildCreateInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		model     NestedModel
		validate  func(t *testing.T, input *fastly.CreateSumologicInput)
	}{
		{
			name:      "minimal model",
			serviceID: "service-123",
			version:   5,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateSumologicInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test-sumologic", *input.Name)
				assert.Equal(t, "https://foo.sumologic.com", *input.URL)
				assert.Equal(t, DefaultMessageType, *input.MessageType)
				assert.Equal(t, "none", *input.ProcessingRegion)
				assert.Equal(t, constants.LoggingSumologicDefaultFormat, *input.Format)
				assert.Nil(t, input.Placement, "unset placement must not be sent as \"none\" — the API treats them differently")
			},
		},
		{
			name:      "fully populated model",
			serviceID: "service-456",
			version:   10,
			model:     fullNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateSumologicInput) {
				assert.Equal(t, "test-sumologic", *input.Name)
				assert.Equal(t, "https://foo.sumologic.com", *input.URL)
				assert.Equal(t, "loggly", *input.MessageType)
				assert.Equal(t, "eu", *input.ProcessingRegion)
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
	assert.Equal(t, "test-sumologic", *input.Name)
	assert.Equal(t, "https://foo.sumologic.com", *input.URL)
	assert.Equal(t, "loggly", *input.MessageType)
	assert.Equal(t, "eu", *input.ProcessingRegion)
	assert.Nil(t, input.Format, "VCL-only fields must never be set for Compute")
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestBuildUpdateInput(t *testing.T) {
	input := BuildUpdateInput("service-456", 10, fullNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-sumologic", input.Name)
	assert.Equal(t, "test-sumologic", *input.NewName)
	assert.Equal(t, "https://foo.sumologic.com", *input.URL)
	assert.Equal(t, "loggly", *input.MessageType)
	assert.Equal(t, "eu", *input.ProcessingRegion)
	assert.Equal(t, "%h %l %u", *input.Format)
	assert.Equal(t, 1, *input.FormatVersion)
	assert.Equal(t, fastly.NewNullable("none"), input.Placement)
	assert.Equal(t, "response-condition-1", *input.ResponseCondition)
}

// TestBuildUpdateInputClearsClearableFields verifies that response_condition is
// always sent as a concrete value on update — even when empty — so clearing it
// actually reaches the API rather than being omitted (which would leave a
// previously-set value in place). placement is cleared the same way, but as an
// explicit JSON null rather than an empty string — see BuildUpdateInput.
func TestBuildUpdateInputClearsClearableFields(t *testing.T) {
	input := BuildUpdateInput("service-1", 1, minimalNestedModel())

	assert.NotNil(t, input.ResponseCondition, "response_condition must be sent even when empty")
	assert.Equal(t, "", *input.ResponseCondition)
	assert.NotNil(t, input.Placement, "unset placement must be sent as an explicit null, not omitted (omitting leaves a previously-set \"none\" in place)")
	assert.Equal(t, fastly.NullValue[string](), input.Placement)
}

func TestBuildComputeUpdateInput(t *testing.T) {
	input := BuildComputeUpdateInput("service-456", 10, fullComputeNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-sumologic", input.Name)
	assert.Equal(t, "test-sumologic", *input.NewName)
	assert.Equal(t, "https://foo.sumologic.com", *input.URL)
	assert.Equal(t, "loggly", *input.MessageType)
	assert.Equal(t, "eu", *input.ProcessingRegion)
	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestClearVCLOnlyCreateFields(t *testing.T) {
	input := &fastly.CreateSumologicInput{
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
	input := &fastly.UpdateSumologicInput{
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
	// What the API actually reports back for a Compute service.
	m := FlattenToNestedModel(&fastly.Sumologic{
		Name:              new("test-sumologic"),
		URL:               new("https://foo.sumologic.com"),
		MessageType:       new("classic"),
		ProcessingRegion:  new("none"),
		Format:            new("{\n  \"source\": \"fastly\"\n}\n"),
		FormatVersion:     new(1),
		Placement:         new("none"),
		ResponseCondition: new("some-condition"),
	})

	ResetVCLOnlyToDefaults(&m)

	assert.Equal(t, constants.LoggingSumologicDefaultFormat, m.Format.ValueString())
	assert.Equal(t, int64(DefaultFormatVersion), m.FormatVersion.ValueInt64())
	assert.True(t, m.Placement.IsNull(), "placement must go back to unset, not the API's forced \"none\"")
	assert.Equal(t, DefaultResponseCondition, m.ResponseCondition.ValueString())

	// Non-VCL-only fields must survive untouched.
	assert.Equal(t, "test-sumologic", m.Name.ValueString())
	assert.Equal(t, "https://foo.sumologic.com", m.URL.ValueString())
	assert.Equal(t, "classic", m.MessageType.ValueString())
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
			name: "different url",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.URL = types.StringValue("https://one.sumologic.com")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.URL = types.StringValue("https://two.sumologic.com")
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

	b.MessageType = types.StringValue("classic")
	assert.False(t, a.ModelsEqual(b))
}

// TestComputeModelsEqualIgnoresVCLOnlyFields verifies that a Compute endpoint
// whose remote state carries VCL-only fields still compares equal to the desired
// Compute model — otherwise ComputeReconcile would issue a pointless update on
// every apply.
func TestComputeModelsEqualIgnoresVCLOnlyFields(t *testing.T) {
	desired := fullComputeNestedModel()

	remote := &fastly.Sumologic{
		Name:              new("test-sumologic"),
		URL:               new("https://foo.sumologic.com"),
		MessageType:       new("loggly"),
		ProcessingRegion:  new("eu"),
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
					m.MessageType = types.StringValue("loggly")
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

	b[0].MessageType = types.StringValue("classic")
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
// `logging_sumologic { format = ... }` inside fastly_service_compute_auto fail
// at plan time with Terraform's own "Unsupported argument" error.
func TestComputeAttributesOmitsVCLOnly(t *testing.T) {
	compute := ComputeAttributes()
	common := CommonAttributes()

	for _, name := range []string{"format", "format_version", "placement", "response_condition"} {
		assert.NotContains(t, compute, name)
		assert.Contains(t, common, name)
	}

	for _, name := range []string{"name", "url", "message_type", "processing_region"} {
		assert.Contains(t, compute, name)
		assert.Contains(t, common, name)
	}
}

// TestSchemaValidators pins the accepted values for the remaining validators, so
// a change to an enum member or a bound is a test failure rather than a surprise
// at plan time: message_type classic/loggly/logplex/blank, processing_region
// none/us/eu, format_version 1-2, placement "none" only, format capped at
// 12288, url must be an absolute http(s) URL with a host.
func TestSchemaValidators(t *testing.T) {
	attrs := CommonAttributes()

	stringCases := []struct {
		name  string
		attr  string
		value string
		valid bool
	}{
		{"message_type classic", "message_type", "classic", true},
		{"message_type loggly", "message_type", "loggly", true},
		{"message_type logplex", "message_type", "logplex", true},
		{"message_type blank", "message_type", "blank", true},
		{"message_type rejects unknown", "message_type", "something-else", false},
		{"processing_region none", "processing_region", "none", true},
		{"processing_region us", "processing_region", "us", true},
		{"processing_region eu", "processing_region", "eu", true},
		{"processing_region rejects wrong case", "processing_region", "US", false},
		{"processing_region rejects empty", "processing_region", "", false},
		{"placement none", "placement", "none", true},
		{"placement rejects other VCL subroutines", "placement", "waf_debug", false},
		{"placement rejects empty", "placement", "", false},
		{"format at max length", "format", strings.Repeat("x", maximumFormatLength), true},
		{"format over max length", "format", strings.Repeat("x", maximumFormatLength+1), false},
		{"url accepts https", "url", "https://collectors.sumologic.com/receiver/v1/http/token", true},
		{"url accepts http", "url", "http://collectors.sumologic.com/receiver/v1/http/token", true},
		{"url rejects missing scheme", "url", "collectors.sumologic.com/receiver/v1/http/token", false},
		{"url rejects non-http(s) scheme", "url", "ftp://collectors.sumologic.com/receiver/v1/http/token", false},
		{"url rejects missing host", "url", "https://", false},
		{"url rejects empty", "url", "", false},
	}
	for _, tt := range stringCases {
		t.Run(tt.name, func(t *testing.T) {
			a := attrs[tt.attr].(schema.StringAttribute)
			require.Len(t, a.Validators, 1)
			resp := &validator.StringResponse{}
			a.Validators[0].ValidateString(context.Background(),
				validator.StringRequest{ConfigValue: types.StringValue(tt.value)}, resp)
			assert.Equal(t, tt.valid, !resp.Diagnostics.HasError())
		})
	}

	int64Cases := []struct {
		value int64
		valid bool
	}{{0, false}, {1, true}, {2, true}, {3, false}}
	for _, tt := range int64Cases {
		t.Run(fmt.Sprintf("format_version %d", tt.value), func(t *testing.T) {
			a := attrs["format_version"].(schema.Int64Attribute)
			require.Len(t, a.Validators, 1)
			resp := &validator.Int64Response{}
			a.Validators[0].ValidateInt64(context.Background(),
				validator.Int64Request{ConfigValue: types.Int64Value(tt.value)}, resp)
			assert.Equal(t, tt.valid, !resp.Diagnostics.HasError())
		})
	}

	// name and response_condition accept any string; assert that rather than
	// leaving it implicit.
	assert.Empty(t, attrs["name"].(schema.StringAttribute).Validators)
	assert.Empty(t, attrs["response_condition"].(schema.StringAttribute).Validators)
}
