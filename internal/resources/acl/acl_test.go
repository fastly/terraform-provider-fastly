package acl

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func ptrTo[T any](v T) *T {
	return &v
}

func defaultNestedModel() NestedModel {
	return NestedModel{
		Name:         types.StringValue(""),
		ACLID:        types.StringValue(""),
		ForceDestroy: types.BoolValue(false),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:         types.StringValue("test_acl"),
		ACLID:        types.StringValue("acl_abc123"),
		ForceDestroy: types.BoolValue(true),
	}
}

func minimalNestedModel() NestedModel {
	m := defaultNestedModel()
	m.Name = types.StringValue("test_acl")
	return m
}

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name     string
		acl      *fastly.ACL
		expected NestedModel
	}{
		{
			name:     "nil ACL returns empty model",
			acl:      nil,
			expected: NestedModel{},
		},
		{
			name: "ACL with all fields populated",
			acl: &fastly.ACL{
				Name:  ptrTo("test_acl"),
				ACLID: ptrTo("acl_abc123"),
			},
			expected: NestedModel{
				Name:         types.StringValue("test_acl"),
				ACLID:        types.StringValue("acl_abc123"),
				ForceDestroy: types.BoolValue(DefaultForceDestroy),
			},
		},
		{
			name: "ACL with minimal fields",
			acl: &fastly.ACL{
				Name: ptrTo("minimal_acl"),
			},
			expected: NestedModel{
				Name:         types.StringValue("minimal_acl"),
				ACLID:        types.StringValue(""),
				ForceDestroy: types.BoolValue(DefaultForceDestroy),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenToNestedModel(tt.acl)
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
	assert.Equal(t, nested.ACLID, m.ACLID)
	assert.Equal(t, nested.ForceDestroy, m.ForceDestroy)

	assert.Equal(t, types.StringValue("test-id"), m.ID)
	assert.Equal(t, types.StringValue("test-service"), m.Service)
	assert.Equal(t, types.Int64Value(1), m.Version)

	extracted := m.NestedModel
	assert.Equal(t, nested, extracted)
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		acl      *fastly.ACL
		validate func(t *testing.T, m *Model)
	}{
		{
			name: "nil ACL logs warning",
			acl:  nil,
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.String{}, m.ID)
				assert.Equal(t, types.String{}, m.Service)
				assert.Equal(t, types.Int64{}, m.Version)
			},
		},
		{
			name: "ACL with service metadata",
			acl: &fastly.ACL{
				ServiceID:      ptrTo("service_123"),
				ServiceVersion: ptrTo(5),
				Name:           ptrTo("test_acl"),
				ACLID:          ptrTo("acl_abc123"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service_123-5-test_acl"), m.ID)
				assert.Equal(t, types.StringValue("service_123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("test_acl"), m.Name)
				assert.Equal(t, types.StringValue("acl_abc123"), m.ACLID)
			},
		},
		{
			name: "ACL with all fields",
			acl: &fastly.ACL{
				ServiceID:      ptrTo("service_456"),
				ServiceVersion: ptrTo(10),
				Name:           ptrTo("full_acl"),
				ACLID:          ptrTo("acl_xyz789"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service_456-10-full_acl"), m.ID)
				assert.Equal(t, types.StringValue("service_456"), m.Service)
				assert.Equal(t, types.Int64Value(10), m.Version)
				assert.Equal(t, types.StringValue("full_acl"), m.Name)
				assert.Equal(t, types.StringValue("acl_xyz789"), m.ACLID)
				assert.Equal(t, types.BoolValue(DefaultForceDestroy), m.ForceDestroy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Model{}
			flatten(ctx, tt.acl, m)
			tt.validate(t, m)
		})
	}
}

func TestBuildCreateInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		model     NestedModel
		validate  func(t *testing.T, input *fastly.CreateACLInput)
	}{
		{
			name:      "minimal ACL",
			serviceID: "service_123",
			version:   5,
			model:     minimalNestedModel(),
			validate: func(t *testing.T, input *fastly.CreateACLInput) {
				assert.Equal(t, "service_123", input.ServiceID)
				assert.Equal(t, 5, input.ServiceVersion)
				assert.Equal(t, "test_acl", *input.Name)
			},
		},
		{
			name:      "ACL with full name",
			serviceID: "service_456",
			version:   10,
			model: func() NestedModel {
				m := fullNestedModel()
				m.Name = types.StringValue("full_acl")
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateACLInput) {
				assert.Equal(t, "service_456", input.ServiceID)
				assert.Equal(t, 10, input.ServiceVersion)
				assert.Equal(t, "full_acl", *input.Name)
			},
		},
		{
			name:      "ACL with special characters in name",
			serviceID: "service_789",
			version:   1,
			model: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("my_special_acl_123")
				return m
			}(),
			validate: func(t *testing.T, input *fastly.CreateACLInput) {
				assert.Equal(t, "service_789", input.ServiceID)
				assert.Equal(t, 1, input.ServiceVersion)
				assert.Equal(t, "my_special_acl_123", *input.Name)
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

func TestBuildDeleteInput(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		version   int
		aclName   string
		expected  *fastly.DeleteACLInput
	}{
		{
			name:      "delete minimal ACL",
			serviceID: "service_123",
			version:   5,
			aclName:   "test_acl",
			expected: &fastly.DeleteACLInput{
				ServiceID:      "service_123",
				ServiceVersion: 5,
				Name:           "test_acl",
			},
		},
		{
			name:      "delete ACL with long name",
			serviceID: "service_abc",
			version:   100,
			aclName:   "very_long_acl_name_with_many_underscores",
			expected: &fastly.DeleteACLInput{
				ServiceID:      "service_abc",
				ServiceVersion: 100,
				Name:           "very_long_acl_name_with_many_underscores",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteInput(tt.serviceID, tt.version, tt.aclName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

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
				m.Name = types.StringValue("acl_1")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("acl_2")
				return m
			}(),
			expected: false,
		},
		{
			name: "different acl_id",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.ACLID = types.StringValue("acl_abc123")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.ACLID = types.StringValue("acl_xyz789")
				return m
			}(),
			expected: false,
		},
		{
			name: "different force_destroy",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.ForceDestroy = types.BoolValue(true)
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.ForceDestroy = types.BoolValue(false)
				return m
			}(),
			expected: true,
		},
		{
			name: "empty string vs empty string ACLID",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.ACLID = types.StringValue("")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.ACLID = types.StringValue("")
				return m
			}(),
			expected: true,
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
			a:        []NestedModel{},
			b:        []NestedModel{},
			expected: true,
		},
		{
			name: "identical single element",
			a: []NestedModel{
				minimalNestedModel(),
			},
			b: []NestedModel{
				minimalNestedModel(),
			},
			expected: true,
		},
		{
			name: "identical multiple elements",
			a: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_b"), ACLID: types.StringValue("id_b")},
			},
			b: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_b"), ACLID: types.StringValue("id_b")},
			},
			expected: true,
		},
		{
			name: "different order but same content",
			a: []NestedModel{
				{Name: types.StringValue("acl_b"), ACLID: types.StringValue("id_b")},
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_b"), ACLID: types.StringValue("id_b")},
			},
			expected: true,
		},
		{
			name: "different lengths",
			a: []NestedModel{
				{Name: types.StringValue("acl_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("acl_a")},
				{Name: types.StringValue("acl_b")},
			},
			expected: false,
		},
		{
			name: "different content",
			a: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_different")},
			},
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

func TestMatchOrder(t *testing.T) {
	tests := []struct {
		name     string
		items    []NestedModel
		order    []NestedModel
		expected []NestedModel
	}{
		{
			name:     "empty items",
			items:    []NestedModel{},
			order:    []NestedModel{{Name: types.StringValue("acl_a")}},
			expected: []NestedModel{},
		},
		{
			name:     "empty order",
			items:    []NestedModel{{Name: types.StringValue("acl_a")}},
			order:    []NestedModel{},
			expected: []NestedModel{{Name: types.StringValue("acl_a")}},
		},
		{
			name: "items match order exactly",
			items: []NestedModel{
				{Name: types.StringValue("acl_b"), ACLID: types.StringValue("id_b")},
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_c"), ACLID: types.StringValue("id_c")},
			},
			order: []NestedModel{
				{Name: types.StringValue("acl_a")},
				{Name: types.StringValue("acl_b")},
				{Name: types.StringValue("acl_c")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_b"), ACLID: types.StringValue("id_b")},
				{Name: types.StringValue("acl_c"), ACLID: types.StringValue("id_c")},
			},
		},
		{
			name: "items not in order are appended",
			items: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_b"), ACLID: types.StringValue("id_b")},
				{Name: types.StringValue("acl_d"), ACLID: types.StringValue("id_d")},
			},
			order: []NestedModel{
				{Name: types.StringValue("acl_a")},
				{Name: types.StringValue("acl_c")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_b"), ACLID: types.StringValue("id_b")},
				{Name: types.StringValue("acl_d"), ACLID: types.StringValue("id_d")},
			},
		},
		{
			name: "order specifies items not in items list",
			items: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_c"), ACLID: types.StringValue("id_c")},
			},
			order: []NestedModel{
				{Name: types.StringValue("acl_b")},
				{Name: types.StringValue("acl_a")},
				{Name: types.StringValue("acl_d")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("acl_a"), ACLID: types.StringValue("id_a")},
				{Name: types.StringValue("acl_c"), ACLID: types.StringValue("id_c")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchOrder(tt.items, tt.order)
			assert.Equal(t, tt.expected, result)
		})
	}
}
