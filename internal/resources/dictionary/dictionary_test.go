package dictionary

import (
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:         types.StringValue("test_dictionary"),
		DictionaryID: types.StringValue(""),
		WriteOnly:    types.BoolValue(DefaultWriteOnly),
		ForceDestroy: types.BoolValue(DefaultForceDestroy),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:         types.StringValue("test_dictionary"),
		DictionaryID: types.StringValue("dict_abc123"),
		WriteOnly:    types.BoolValue(true),
		ForceDestroy: types.BoolValue(true),
	}
}

func TestToModel(t *testing.T) {
	api := &fastly.Dictionary{
		Name:         new("test_dictionary"),
		DictionaryID: new("dict_abc123"),
		WriteOnly:    new(true),
	}

	result := ops{}.ToModel(api)

	assert.Equal(t, types.StringValue("test_dictionary"), result.Name)
	assert.Equal(t, types.StringValue("dict_abc123"), result.DictionaryID)
	assert.Equal(t, types.BoolValue(true), result.WriteOnly)
	// ForceDestroy is configuration-only and is never returned by the API,
	// so ToModel always fills in the default.
	assert.Equal(t, types.BoolValue(DefaultForceDestroy), result.ForceDestroy)
}

func TestOpsEqual(t *testing.T) {
	remote := &fastly.Dictionary{
		Name:         new("test_dictionary"),
		DictionaryID: new("dict_abc123"),
		WriteOnly:    new(false),
	}

	tests := []struct {
		name     string
		desired  NestedModel
		expected bool
	}{
		{
			name:     "matching name and write_only",
			desired:  minimalNestedModel(),
			expected: true,
		},
		{
			name: "different write_only",
			desired: func() NestedModel {
				m := minimalNestedModel()
				m.WriteOnly = types.BoolValue(true)
				return m
			}(),
			expected: false,
		},
		{
			name: "different name",
			desired: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("other_dictionary")
				return m
			}(),
			expected: false,
		},
		{
			name: "force_destroy is ignored, since it is configuration-only",
			desired: func() NestedModel {
				m := minimalNestedModel()
				m.ForceDestroy = types.BoolValue(true)
				return m
			}(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ops{}.Equal(tt.desired, remote))
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
			name:     "identical full models",
			a:        fullNestedModel(),
			b:        fullNestedModel(),
			expected: true,
		},
		{
			name:     "identical minimal models",
			a:        minimalNestedModel(),
			b:        minimalNestedModel(),
			expected: true,
		},
		{
			name: "different name",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("dict_a")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("dict_b")
				return m
			}(),
			expected: false,
		},
		{
			name: "different write_only",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.WriteOnly = types.BoolValue(true)
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.WriteOnly = types.BoolValue(false)
				return m
			}(),
			expected: false,
		},
		{
			name: "different force_destroy does not affect equality",
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
			name: "different dictionary_id does not affect equality",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.DictionaryID = types.StringValue("dict_1")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.DictionaryID = types.StringValue("dict_2")
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
			name:     "identical single element",
			a:        []NestedModel{minimalNestedModel()},
			b:        []NestedModel{minimalNestedModel()},
			expected: true,
		},
		{
			name: "different order but same content",
			a: []NestedModel{
				{Name: types.StringValue("dict_b")},
				{Name: types.StringValue("dict_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("dict_a")},
				{Name: types.StringValue("dict_b")},
			},
			expected: true,
		},
		{
			name: "different lengths",
			a: []NestedModel{
				{Name: types.StringValue("dict_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("dict_a")},
				{Name: types.StringValue("dict_b")},
			},
			expected: false,
		},
		{
			name: "different content",
			a: []NestedModel{
				{Name: types.StringValue("dict_a"), WriteOnly: types.BoolValue(false)},
			},
			b: []NestedModel{
				{Name: types.StringValue("dict_a"), WriteOnly: types.BoolValue(true)},
			},
			expected: false,
		},
		{
			// Regression test: unlike ops.Equal (used for live API comparison, where
			// force_destroy is irrelevant since it's never sent to the API), the
			// top-level Equal must treat a force_destroy-only difference as a change.
			// Otherwise servicecdnauto/servicecomputeauto's Update takes their
			// no-op branch, which reorders nested items from stale prior state and
			// never applies the new force_destroy value.
			name: "different force_destroy is a change",
			a: []NestedModel{
				{Name: types.StringValue("dict_a"), WriteOnly: types.BoolValue(false), ForceDestroy: types.BoolValue(false)},
			},
			b: []NestedModel{
				{Name: types.StringValue("dict_a"), WriteOnly: types.BoolValue(false), ForceDestroy: types.BoolValue(true)},
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
			order:    []NestedModel{{Name: types.StringValue("dict_a")}},
			expected: []NestedModel{},
		},
		{
			name: "items match order exactly",
			items: []NestedModel{
				{Name: types.StringValue("dict_b"), DictionaryID: types.StringValue("id_b")},
				{Name: types.StringValue("dict_a"), DictionaryID: types.StringValue("id_a")},
			},
			order: []NestedModel{
				{Name: types.StringValue("dict_a")},
				{Name: types.StringValue("dict_b")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("dict_a"), DictionaryID: types.StringValue("id_a")},
				{Name: types.StringValue("dict_b"), DictionaryID: types.StringValue("id_b")},
			},
		},
		{
			name: "items not in order are appended",
			items: []NestedModel{
				{Name: types.StringValue("dict_a"), DictionaryID: types.StringValue("id_a")},
				{Name: types.StringValue("dict_c"), DictionaryID: types.StringValue("id_c")},
			},
			order: []NestedModel{
				{Name: types.StringValue("dict_a")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("dict_a"), DictionaryID: types.StringValue("id_a")},
				{Name: types.StringValue("dict_c"), DictionaryID: types.StringValue("id_c")},
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
