package loggingnewrelicotlp

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
)

// Test helpers

func defaultNestedModel() NestedModel {
	return NestedModel{
		commonModel:       defaultCommonModel(),
		Format:            types.StringValue(constants.LoggingNewRelicOTLPDefaultFormat),
		FormatVersion:     types.Int64Value(DefaultFormatVersion),
		Placement:         types.StringValue(DefaultPlacement),
		ResponseCondition: types.StringValue(DefaultResponseCondition),
	}
}

func defaultCommonModel() commonModel {
	return commonModel{
		Name:             types.StringValue(""),
		Token:            types.StringValue(""),
		Region:           types.StringValue(DefaultRegion),
		URL:              types.StringValue(DefaultURL),
		ProcessingRegion: types.StringValue(DefaultProcessingRegion),
	}
}

func fullNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-newrelic")
	m.Token = types.StringValue("insert-api-key")
	m.Region = types.StringValue("EU")
	m.URL = types.StringValue("https://otlp.eu01.nr-data.net")
	m.ProcessingRegion = types.StringValue("eu")
	m.Format = types.StringValue("%h %l %u")
	m.FormatVersion = types.Int64Value(1)
	m.Placement = types.StringValue("waf_debug")
	m.ResponseCondition = types.StringValue("response-condition-1")
	return m
}

func minimalNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-newrelic")
	m.Token = types.StringValue("insert-api-key")
	return m
}

func fullComputeNestedModel() ComputeNestedModel {
	return ComputeNestedModel{commonModel: fullNestedModel().commonModel}
}

// Tests for flatten.go

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name     string
		api      *fastly.NewRelicOTLP
		expected NestedModel
	}{
		{
			name:     "nil returns empty model",
			api:      nil,
			expected: NestedModel{},
		},
		{
			name: "only required fields uses defaults",
			api: &fastly.NewRelicOTLP{
				Name:  new("test-newrelic"),
				Token: new("insert-api-key"),
			},
			expected: minimalNestedModel(),
		},
		{
			name: "all fields populated",
			api: &fastly.NewRelicOTLP{
				Name:              new("test-newrelic"),
				Token:             new("insert-api-key"),
				Region:            new("EU"),
				URL:               new("https://otlp.eu01.nr-data.net"),
				ProcessingRegion:  new("eu"),
				Format:            new("%h %l %u"),
				FormatVersion:     new(1),
				Placement:         new("waf_debug"),
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
	api := &fastly.NewRelicOTLP{
		Name:             new("test-newrelic"),
		Token:            new("insert-api-key"),
		Region:           new("EU"),
		URL:              new("https://otlp.eu01.nr-data.net"),
		ProcessingRegion: new("eu"),
		// VCL-only fields must be ignored by the Compute flatten.
		Format:            new("%h %l %u"),
		FormatVersion:     new(1),
		Placement:         new("waf_debug"),
		ResponseCondition: new("response-condition-1"),
	}

	result := FlattenToComputeNestedModel(api)
	assert.Equal(t, fullComputeNestedModel(), result)
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		api      *fastly.NewRelicOTLP
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
			api: &fastly.NewRelicOTLP{
				ServiceID:      new("service-123"),
				ServiceVersion: new(5),
				Name:           new("test-newrelic"),
				Token:          new("insert-api-key"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service-123-5-test-newrelic"), m.ID)
				assert.Equal(t, types.StringValue("service-123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("test-newrelic"), m.Name)
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
		validate  func(t *testing.T, input *fastly.CreateNewRelicOTLPInput)
	}{
		{
			name:      "minimal model",
			serviceID: "service-123",
			version:   5,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateNewRelicOTLPInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test-newrelic", *input.Name)
				assert.Equal(t, "insert-api-key", *input.Token)
				assert.Equal(t, "US", *input.Region)
				assert.Equal(t, "none", *input.ProcessingRegion)
			},
		},
		{
			name:      "fully populated model",
			serviceID: "service-456",
			version:   10,
			model:     fullNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateNewRelicOTLPInput) {
				assert.Equal(t, "test-newrelic", *input.Name)
				assert.Equal(t, "insert-api-key", *input.Token)
				assert.Equal(t, "EU", *input.Region)
				assert.Equal(t, "https://otlp.eu01.nr-data.net", *input.URL)
				assert.Equal(t, "eu", *input.ProcessingRegion)
				assert.Equal(t, "%h %l %u", *input.Format)
				assert.Equal(t, 1, *input.FormatVersion)
				assert.Equal(t, "waf_debug", *input.Placement)
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
	assert.Equal(t, "test-newrelic", *input.Name)
	assert.Equal(t, "insert-api-key", *input.Token)
	assert.Equal(t, "EU", *input.Region)
	assert.Nil(t, input.Format, "VCL-only fields must never be set for Compute")
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestBuildUpdateInput(t *testing.T) {
	input := BuildUpdateInput("service-456", 10, fullNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-newrelic", input.Name)
	assert.Equal(t, "test-newrelic", *input.NewName)
	assert.Equal(t, "insert-api-key", *input.Token)
	assert.Equal(t, "EU", *input.Region)
	assert.Equal(t, "%h %l %u", *input.Format)
	assert.Equal(t, 1, *input.FormatVersion)
	assert.Equal(t, "waf_debug", *input.Placement)
	assert.Equal(t, "response-condition-1", *input.ResponseCondition)
}

// TestBuildUpdateInputClearsClearableFields verifies that url and
// response_condition are always sent as a concrete value on update — even when
// empty — so clearing them actually reaches the API rather than being omitted
// (which would leave a previously-set value in place).
func TestBuildUpdateInputClearsClearableFields(t *testing.T) {
	input := BuildUpdateInput("service-1", 1, minimalNestedModel())

	assert.NotNil(t, input.URL, "url must be sent even when empty")
	assert.Equal(t, "", *input.URL)
	assert.NotNil(t, input.ResponseCondition, "response_condition must be sent even when empty")
	assert.Equal(t, "", *input.ResponseCondition)
}

func TestBuildComputeUpdateInputClearsURL(t *testing.T) {
	input := BuildComputeUpdateInput("service-1", 1, ComputeNestedModel{commonModel: minimalNestedModel().commonModel})

	assert.NotNil(t, input.URL, "url must be sent even when empty")
	assert.Equal(t, "", *input.URL)
}

func TestBuildComputeUpdateInput(t *testing.T) {
	input := BuildComputeUpdateInput("service-456", 10, fullComputeNestedModel())

	assert.Equal(t, "service-456", input.ServiceID)
	assert.Equal(t, 10, input.ServiceVersion)
	assert.Equal(t, "test-newrelic", input.Name)
	assert.Equal(t, "test-newrelic", *input.NewName)
	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestClearVCLOnlyCreateFields(t *testing.T) {
	input := &fastly.CreateNewRelicOTLPInput{
		Format:            new("some-format"),
		FormatVersion:     new(2),
		Placement:         new("waf_debug"),
		ResponseCondition: new("cond"),
	}

	ClearVCLOnlyCreateFields(input)

	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
}

func TestClearVCLOnlyUpdateFields(t *testing.T) {
	input := &fastly.UpdateNewRelicOTLPInput{
		Format:            new("some-format"),
		FormatVersion:     new(2),
		Placement:         new("waf_debug"),
		ResponseCondition: new("cond"),
	}

	ClearVCLOnlyUpdateFields(input)

	assert.Nil(t, input.Format)
	assert.Nil(t, input.FormatVersion)
	assert.Nil(t, input.Placement)
	assert.Nil(t, input.ResponseCondition)
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
			name: "different token",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Token = types.StringValue("token-1")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Token = types.StringValue("token-2")
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

	b.Region = types.StringValue("US")
	assert.False(t, a.ModelsEqual(b))
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
					m.Region = types.StringValue("EU")
					return m
				}(),
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

	b[0].Region = types.StringValue("US")
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
