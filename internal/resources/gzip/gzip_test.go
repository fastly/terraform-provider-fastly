package gzip

import (
	"testing"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func stringList(values ...string) types.List {
	elems := make([]attr.Value, len(values))
	for i, v := range values {
		elems[i] = types.StringValue(v)
	}
	return types.ListValueMust(types.StringType, elems)
}

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:           types.StringValue("gzip-config"),
		CacheCondition: types.StringNull(),
		ContentTypes:   types.ListNull(types.StringType),
		Extensions:     types.ListNull(types.StringType),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:           types.StringValue("gzip-config"),
		CacheCondition: types.StringValue("cache-condition"),
		ContentTypes:   stringList("text/html", "text/css"),
		Extensions:     stringList("css", "js"),
	}
}

func TestJoinStringList(t *testing.T) {
	tests := []struct {
		name     string
		list     types.List
		expected string
	}{
		{name: "null", list: types.ListNull(types.StringType), expected: ""},
		{name: "unknown", list: types.ListUnknown(types.StringType), expected: ""},
		{name: "empty", list: stringList(), expected: ""},
		{name: "single element", list: stringList("css"), expected: "css"},
		{name: "multiple elements", list: stringList("css", "js"), expected: "css js"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStringList(tt.list)
			if assert.NotNil(t, result) {
				assert.Equal(t, tt.expected, *result)
			}
		})
	}
}

func TestStringListValue(t *testing.T) {
	tests := []struct {
		name     string
		raw      *string
		expected types.List
	}{
		{name: "nil", raw: nil, expected: types.ListNull(types.StringType)},
		{name: "empty string", raw: new(""), expected: types.ListNull(types.StringType)},
		{name: "single element", raw: new("css"), expected: stringList("css")},
		{name: "multiple elements", raw: new("css js"), expected: stringList("css", "js")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringListValue(tt.raw)
			assert.True(t, tt.expected.Equal(result))
		})
	}
}

func TestToModel(t *testing.T) {
	api := &fastly.Gzip{
		Name:           new("gzip-config"),
		CacheCondition: new("cache-condition"),
		ContentTypes:   new("text/html text/css"),
		Extensions:     new("css js"),
	}

	result := ops{}.ToModel(api)

	assert.Equal(t, "gzip-config", result.Name.ValueString())
	assert.Equal(t, "cache-condition", result.CacheCondition.ValueString())
	assert.True(t, stringList("text/html", "text/css").Equal(result.ContentTypes))
	assert.True(t, stringList("css", "js").Equal(result.Extensions))
}

func TestToModel_emptyOptionalFields(t *testing.T) {
	api := &fastly.Gzip{
		Name:           new("gzip-config"),
		CacheCondition: new(""),
		ContentTypes:   new(""),
		Extensions:     new(""),
	}

	result := ops{}.ToModel(api)

	assert.Equal(t, "gzip-config", result.Name.ValueString())
	assert.True(t, result.CacheCondition.IsNull())
	assert.True(t, result.ContentTypes.IsNull())
	assert.True(t, result.Extensions.IsNull())
}

func TestOpsEqual_explicitEmptyListTreatedAsUnset(t *testing.T) {
	// Regression test: the Fastly API substitutes a default list when
	// content_types/extensions are omitted or empty, so an explicit empty
	// list in config must be compared as unset against the remote's
	// defaulted value rather than producing a permanent diff.
	desired := NestedModel{
		Name:           types.StringValue("gzip-config"),
		CacheCondition: types.StringNull(),
		ContentTypes:   stringList(),
		Extensions:     stringList(),
	}
	remote := &fastly.Gzip{
		Name:           new("gzip-config"),
		CacheCondition: new(""),
		ContentTypes:   new("text/html text/css text/javascript application/json"),
		Extensions:     new("css js"),
	}

	assert.True(t, ops{}.Equal(desired, remote))
}

func TestOpsWithPreviousEqual_neverConfiguredTreatedAsUnset(t *testing.T) {
	// A field that was never configured (no previous item at all) behaves the
	// same as the plain ops.Equal case: the remote's API-substituted default
	// should not be treated as a diff.
	desired := NestedModel{
		Name:           types.StringValue("gzip-config"),
		CacheCondition: types.StringNull(),
		ContentTypes:   types.ListNull(types.StringType),
		Extensions:     types.ListNull(types.StringType),
	}
	remote := &fastly.Gzip{
		Name:           new("gzip-config"),
		CacheCondition: new(""),
		ContentTypes:   new("text/html text/css text/javascript application/json"),
		Extensions:     new("css js"),
	}

	o := opsWithPrevious{previousByName: map[string]NestedModel{}}
	assert.True(t, o.Equal(desired, remote))
}

func TestOpsWithPreviousEqual_removedValueNotTreatedAsUnset(t *testing.T) {
	// Regression test: removing a previously configured content_types/extensions
	// value from config must not be treated as equal to the (still populated)
	// remote value, otherwise Update is skipped and the stale value persists
	// remotely while state hides it (see philippschulte's PR review).
	desired := NestedModel{
		Name:           types.StringValue("gzip-config"),
		CacheCondition: types.StringNull(),
		ContentTypes:   types.ListNull(types.StringType),
		Extensions:     stringList("css", "js"),
	}
	remote := &fastly.Gzip{
		Name:           new("gzip-config"),
		CacheCondition: new(""),
		ContentTypes:   new("text/html"),
		Extensions:     new("css js"),
	}

	previous := NestedModel{
		Name:         types.StringValue("gzip-config"),
		ContentTypes: stringList("text/html"),
		Extensions:   stringList("css", "js"),
	}
	o := opsWithPrevious{previousByName: map[string]NestedModel{"gzip-config": previous}}

	assert.False(t, o.Equal(desired, remote))
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
				m.Name = types.StringValue("gzip-a")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("gzip-b")
				return m
			}(),
			expected: false,
		},
		{
			name: "different cache_condition",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.CacheCondition = types.StringValue("condition-a")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.CacheCondition = types.StringValue("condition-b")
				return m
			}(),
			expected: false,
		},
		{
			name: "null vs empty string cache_condition",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.CacheCondition = types.StringNull()
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.CacheCondition = types.StringValue("")
				return m
			}(),
			expected: true,
		},
		{
			name: "null vs empty content_types",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.ContentTypes = types.ListNull(types.StringType)
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.ContentTypes = stringList()
				return m
			}(),
			expected: true,
		},
		{
			name: "different content_types",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.ContentTypes = stringList("text/html")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.ContentTypes = stringList("text/css")
				return m
			}(),
			expected: false,
		},
		{
			name: "different extensions",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Extensions = stringList("css")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Extensions = stringList("js")
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
				{Name: types.StringValue("gzip-b")},
				{Name: types.StringValue("gzip-a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-b")},
			},
			expected: true,
		},
		{
			name: "different lengths",
			a: []NestedModel{
				{Name: types.StringValue("gzip-a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-b")},
			},
			expected: false,
		},
		{
			name: "different content",
			a: []NestedModel{
				{Name: types.StringValue("gzip-a"), CacheCondition: types.StringValue("condition-a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("gzip-a"), CacheCondition: types.StringValue("condition-b")},
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
			order:    []NestedModel{{Name: types.StringValue("gzip-a")}},
			expected: []NestedModel{},
		},
		{
			name: "items match order exactly",
			items: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-b")},
			},
			order: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-b")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-b")},
			},
		},
		{
			name: "items reversed relative to order",
			items: []NestedModel{
				{Name: types.StringValue("gzip-b")},
				{Name: types.StringValue("gzip-a")},
			},
			order: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-b")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-b")},
			},
		},
		{
			name: "items not in order are appended",
			items: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-c")},
			},
			order: []NestedModel{
				{Name: types.StringValue("gzip-a")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("gzip-a")},
				{Name: types.StringValue("gzip-c")},
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
