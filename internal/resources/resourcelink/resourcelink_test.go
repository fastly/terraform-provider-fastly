package resourcelink

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:       types.StringValue("my_acl"),
		ResourceID: types.StringValue("acl_abc123"),
		LinkID:     types.StringValue("link_xyz789"),
	}
}

func TestFlattenToNestedModel(t *testing.T) {
	tests := []struct {
		name     string
		resource *fastly.Resource
		expected NestedModel
	}{
		{
			name:     "nil resource returns empty model",
			resource: nil,
			expected: NestedModel{},
		},
		{
			name: "resource with all fields populated",
			resource: &fastly.Resource{
				Name:       new("my_acl"),
				ResourceID: new("acl_abc123"),
				LinkID:     new("link_xyz789"),
			},
			expected: NestedModel{
				Name:       types.StringValue("my_acl"),
				ResourceID: types.StringValue("acl_abc123"),
				LinkID:     types.StringValue("link_xyz789"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenToNestedModel(tt.resource)
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
	assert.Equal(t, nested.ResourceID, m.ResourceID)
	assert.Equal(t, nested.LinkID, m.LinkID)

	assert.Equal(t, types.StringValue("test-id"), m.ID)
	assert.Equal(t, types.StringValue("test-service"), m.Service)
	assert.Equal(t, types.Int64Value(1), m.Version)
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		resource *fastly.Resource
		validate func(t *testing.T, m *Model)
	}{
		{
			name:     "nil resource logs warning",
			resource: nil,
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.String{}, m.ID)
				assert.Equal(t, types.String{}, m.Service)
				assert.Equal(t, types.Int64{}, m.Version)
			},
		},
		{
			name: "resource with service metadata",
			resource: &fastly.Resource{
				ServiceID:      new("service_123"),
				ServiceVersion: new(5),
				Name:           new("my_acl"),
				ResourceID:     new("acl_abc123"),
				LinkID:         new("link_xyz789"),
			},
			validate: func(t *testing.T, m *Model) {
				assert.Equal(t, types.StringValue("service_123-5-link_xyz789"), m.ID)
				assert.Equal(t, types.StringValue("service_123"), m.Service)
				assert.Equal(t, types.Int64Value(5), m.Version)
				assert.Equal(t, types.StringValue("my_acl"), m.Name)
				assert.Equal(t, types.StringValue("acl_abc123"), m.ResourceID)
				assert.Equal(t, types.StringValue("link_xyz789"), m.LinkID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Model{}
			flatten(ctx, tt.resource, m)
			tt.validate(t, m)
		})
	}
}

func TestBuildCreateInput(t *testing.T) {
	input := BuildCreateInput("service_123", 5, fullNestedModel())

	assert.Equal(t, "service_123", input.ServiceID)
	assert.Equal(t, 5, input.ServiceVersion)
	assert.Equal(t, "my_acl", *input.Name)
	assert.Equal(t, "acl_abc123", *input.ResourceID)
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
			name: "different name, same resource_id",
			a: func() NestedModel {
				m := fullNestedModel()
				m.Name = types.StringValue("alias_1")
				return m
			}(),
			b: func() NestedModel {
				m := fullNestedModel()
				m.Name = types.StringValue("alias_2")
				return m
			}(),
			expected: false,
		},
		{
			name: "different resource_id",
			a: func() NestedModel {
				m := fullNestedModel()
				m.ResourceID = types.StringValue("acl_abc123")
				return m
			}(),
			b: func() NestedModel {
				m := fullNestedModel()
				m.ResourceID = types.StringValue("acl_xyz789")
				return m
			}(),
			expected: false,
		},
		{
			name: "different link_id is ignored",
			a: func() NestedModel {
				m := fullNestedModel()
				m.LinkID = types.StringValue("link_1")
				return m
			}(),
			b: func() NestedModel {
				m := fullNestedModel()
				m.LinkID = types.StringValue("link_2")
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
			name: "identical multiple elements",
			a: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
				{Name: types.StringValue("alias_b"), ResourceID: types.StringValue("id_b")},
			},
			b: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
				{Name: types.StringValue("alias_b"), ResourceID: types.StringValue("id_b")},
			},
			expected: true,
		},
		{
			name: "different order but same content",
			a: []NestedModel{
				{Name: types.StringValue("alias_b"), ResourceID: types.StringValue("id_b")},
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
				{Name: types.StringValue("alias_b"), ResourceID: types.StringValue("id_b")},
			},
			expected: true,
		},
		{
			name: "different lengths",
			a: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
				{Name: types.StringValue("alias_b"), ResourceID: types.StringValue("id_b")},
			},
			expected: false,
		},
		{
			name: "same resource_id, different alias",
			a: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("renamed"), ResourceID: types.StringValue("id_a")},
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
			order:    []NestedModel{{ResourceID: types.StringValue("id_a")}},
			expected: []NestedModel{},
		},
		{
			name: "items match order exactly, keyed by resource_id",
			items: []NestedModel{
				{Name: types.StringValue("alias_b"), ResourceID: types.StringValue("id_b")},
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
			},
			order: []NestedModel{
				{ResourceID: types.StringValue("id_a")},
				{ResourceID: types.StringValue("id_b")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
				{Name: types.StringValue("alias_b"), ResourceID: types.StringValue("id_b")},
			},
		},
		{
			name: "items not in order are appended",
			items: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
				{Name: types.StringValue("alias_d"), ResourceID: types.StringValue("id_d")},
			},
			order: []NestedModel{
				{ResourceID: types.StringValue("id_a")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("alias_a"), ResourceID: types.StringValue("id_a")},
				{Name: types.StringValue("alias_d"), ResourceID: types.StringValue("id_d")},
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
