package condition

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// Test helpers

// defaultNestedModel returns a NestedModel with default values for all fields
func defaultNestedModel() NestedModel {
	return NestedModel{
		Name:      types.StringValue(""),
		Type:      types.StringValue(""),
		Statement: types.StringValue(""),
		Priority:  types.Int64Value(10),
	}
}

// fullNestedModel returns a NestedModel with all fields populated with non-default values
func fullNestedModel() NestedModel {
	return NestedModel{
		Name:      types.StringValue("test-condition"),
		Type:      types.StringValue("REQUEST"),
		Statement: types.StringValue(`req.url ~ "^/admin"`),
		Priority:  types.Int64Value(5),
	}
}

// minimalNestedModel returns a NestedModel with only required fields for BuildCreateInput
func minimalNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test-condition")
	m.Type = types.StringValue("REQUEST")
	m.Statement = types.StringValue(`req.url ~ "^/admin"`)
	return m
}

// Tests for flatten.go

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name      string
		condition *fastly.Condition
		expected  NestedModel
	}{
		{
			name:      "nil condition returns empty model",
			condition: nil,
			expected:  NestedModel{},
		},
		{
			name: "condition with all fields populated",
			condition: &fastly.Condition{
				Name:      new("test-condition"),
				Type:      new("REQUEST"),
				Statement: new(`req.url ~ "^/admin"`),
				Priority:  new(5),
			},
			expected: fullNestedModel(),
		},
		{
			name: "condition with nil priority uses default",
			condition: &fastly.Condition{
				Name:      new("default-condition"),
				Type:      new("CACHE"),
				Statement: new("req.url ~ \"^/\""),
			},
			expected: func() NestedModel {
				m := defaultNestedModel()
				m.Name = types.StringValue("default-condition")
				m.Type = types.StringValue("CACHE")
				m.Statement = types.StringValue(`req.url ~ "^/"`)
				return m
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenToNestedModel(tt.condition)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModelEmbedding(t *testing.T) {
	nested := fullNestedModel()
	m := Model{
		NestedModel: nested,
		ID:          types.StringValue("test-id"),
		Service:     types.StringValue("test-service"),
		Version:     types.Int64Value(1),
	}

	assert.Equal(t, nested.Name, m.Name)
	assert.Equal(t, nested.Type, m.Type)
	assert.Equal(t, nested.Statement, m.Statement)
	assert.Equal(t, nested.Priority, m.Priority)

	assert.Equal(t, types.StringValue("test-id"), m.ID)
	assert.Equal(t, types.StringValue("test-service"), m.Service)
	assert.Equal(t, types.Int64Value(1), m.Version)

	extracted := m.NestedModel
	assert.Equal(t, nested, extracted)
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name      string
		condition *fastly.Condition
		validate  func(t *testing.T, m *Model)
	}{
		{
			name:      "nil condition logs warning",
			condition: nil,
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.String{}, m.ID)
				assert.Equal(t, types.String{}, m.Service)
				assert.Equal(t, types.Int64{}, m.Version)
			},
		},
		{
			name: "condition with service metadata",
			condition: &fastly.Condition{
				ServiceID:      new("service-123"),
				ServiceVersion: new(5),
				Name:           new("test-condition"),
				Type:           new("REQUEST"),
				Statement:      new(`req.url ~ "^/admin"`),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service-123-5-test-condition"), m.ID)
				assert.Equal(t, types.StringValue("service-123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("test-condition"), m.Name)
				assert.Equal(t, types.StringValue("REQUEST"), m.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Model{}
			flatten(ctx, tt.condition, m)
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
		validate  func(t *testing.T, input *fastly.CreateConditionInput)
	}{
		{
			name:      "minimal condition",
			serviceID: "service-123",
			version:   5,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateConditionInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test-condition", *input.Name)
				assert.Equal(t, "REQUEST", *input.Type)
				assert.Equal(t, `req.url ~ "^/admin"`, *input.Statement)
				assert.Equal(t, 10, *input.Priority)
			},
		},
		{
			name:      "condition with custom priority",
			serviceID: "service-456",
			version:   10,
			model: func() NestedModel {
				m := fullNestedModel()
				m.Priority = types.Int64Value(1)
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateConditionInput) {
				assert.Equal(t, 1, *input.Priority)
			},
		},
		{
			name:      "statement with surrounding whitespace is trimmed",
			serviceID: "service-789",
			version:   1,
			model: func() NestedModel {
				m := minimalNestedModel()
				m.Statement = types.StringValue("\n  req.url ~ \"^/admin\"  \n")
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateConditionInput) {
				assert.Equal(t, `req.url ~ "^/admin"`, *input.Statement)
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

func TestBuildUpdateInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		model     NestedModel
		validate  func(t *testing.T, input *fastly.UpdateConditionInput)
	}{
		{
			name:      "condition update omits type",
			serviceID: "service-123",
			version:   5,
			model:     fullNestedModel(),
			validate: func(t *testing.T, input *fastly.UpdateConditionInput) {
				assert.Equal(t, "service-123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test-condition", input.Name)
				assert.Equal(t, `req.url ~ "^/admin"`, *input.Statement)
				assert.Equal(t, 5, *input.Priority)
				assert.Nil(t, input.Type)
			},
		},
		{
			name:      "statement with surrounding whitespace is trimmed",
			serviceID: "service-789",
			version:   1,
			model: func() NestedModel {
				m := minimalNestedModel()
				m.Statement = types.StringValue("\n  req.url ~ \"^/admin\"  \n")
				return m
			}(),
			validate: func(t *testing.T, input *fastly.UpdateConditionInput) {
				assert.Equal(t, `req.url ~ "^/admin"`, *input.Statement)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := BuildUpdateInput(tt.serviceID, tt.version, tt.model)
			tt.validate(t, input)
		})
	}
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
			name: "different name",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("condition-1")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("condition-2")
				return m
			}(),
			expected: false,
		},
		{
			name: "different type",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Type = types.StringValue("REQUEST")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Type = types.StringValue("CACHE")
				return m
			}(),
			expected: false,
		},
		{
			name: "different statement",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Statement = types.StringValue("req.url ~ \"^/a\"")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Statement = types.StringValue("req.url ~ \"^/b\"")
				return m
			}(),
			expected: false,
		},
		{
			name: "different priority",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Priority = types.Int64Value(10)
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Priority = types.Int64Value(20)
				return m
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.ModelsEqual(tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
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
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "identical single item",
			a:        []NestedModel{fullNestedModel()},
			b:        []NestedModel{fullNestedModel()},
			expected: true,
		},
		{
			name: "different order, same contents",
			a: []NestedModel{
				minimalNestedModel(),
				func() NestedModel {
					m := minimalNestedModel()
					m.Name = types.StringValue("second")
					return m
				}(),
			},
			b: []NestedModel{
				func() NestedModel {
					m := minimalNestedModel()
					m.Name = types.StringValue("second")
					return m
				}(),
				minimalNestedModel(),
			},
			expected: true,
		},
		{
			name:     "different lengths",
			a:        []NestedModel{minimalNestedModel()},
			b:        nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Equal(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
