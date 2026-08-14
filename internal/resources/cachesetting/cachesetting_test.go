package cachesetting

import (
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:           types.StringValue("cache-setting"),
		Action:         types.StringNull(),
		CacheCondition: types.StringNull(),
		StaleTTL:       types.Int64Value(0),
		TTL:            types.Int64Value(0),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:           types.StringValue("cache-setting"),
		Action:         types.StringValue("cache"),
		CacheCondition: types.StringValue("cache-condition"),
		StaleTTL:       types.Int64Value(120),
		TTL:            types.Int64Value(3600),
	}
}

func TestToModel(t *testing.T) {
	action := fastly.CacheSettingActionCache
	api := &fastly.CacheSetting{
		Name:           new("cache-setting"),
		Action:         &action,
		CacheCondition: new("cache-condition"),
		StaleTTL:       new(120),
		TTL:            new(3600),
	}

	result := ops{}.ToModel(api)

	assert.True(t, fullNestedModel().ModelsEqual(result))
}

func TestToModel_emptyOptionalFields(t *testing.T) {
	emptyAction := fastly.CacheSettingAction("")
	api := &fastly.CacheSetting{
		Name:           new("cache-setting"),
		Action:         &emptyAction,
		CacheCondition: new(""),
		StaleTTL:       new(0),
		TTL:            new(0),
	}

	result := ops{}.ToModel(api)

	assert.Equal(t, "cache-setting", result.Name.ValueString())
	assert.True(t, result.Action.IsNull())
	assert.True(t, result.CacheCondition.IsNull())
	assert.Equal(t, int64(0), result.StaleTTL.ValueInt64())
	assert.Equal(t, int64(0), result.TTL.ValueInt64())
}

func TestToModel_nilOptionalFields(t *testing.T) {
	api := &fastly.CacheSetting{
		Name:     new("cache-setting"),
		StaleTTL: new(0),
		TTL:      new(0),
	}

	result := ops{}.ToModel(api)

	assert.True(t, result.Action.IsNull())
	assert.True(t, result.CacheCondition.IsNull())
}

func TestOpsEqual(t *testing.T) {
	action := fastly.CacheSettingActionCache
	remote := &fastly.CacheSetting{
		Name:           new("cache-setting"),
		Action:         &action,
		CacheCondition: new("cache-condition"),
		StaleTTL:       new(120),
		TTL:            new(3600),
	}

	assert.True(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestOpsEqual_mismatch(t *testing.T) {
	remote := &fastly.CacheSetting{
		Name:     new("cache-setting"),
		StaleTTL: new(0),
		TTL:      new(0),
	}

	assert.False(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestActionPointer(t *testing.T) {
	tests := []struct {
		name     string
		value    types.String
		expected *fastly.CacheSettingAction
	}{
		{name: "null", value: types.StringNull(), expected: nil},
		{name: "unknown", value: types.StringUnknown(), expected: nil},
		{name: "empty", value: types.StringValue(""), expected: nil},
		{name: "cache", value: types.StringValue("cache"), expected: new(fastly.CacheSettingActionCache)},
		{name: "uppercase", value: types.StringValue("PASS"), expected: new(fastly.CacheSettingActionPass)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := actionPointer(tt.value)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else if assert.NotNil(t, result) {
				assert.Equal(t, *tt.expected, *result)
			}
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
			name:     "same content",
			a:        []NestedModel{fullNestedModel()},
			b:        []NestedModel{fullNestedModel()},
			expected: true,
		},
		{
			name: "different lengths",
			a:    []NestedModel{minimalNestedModel()},
			b: []NestedModel{
				minimalNestedModel(),
				fullNestedModel(),
			},
			expected: false,
		},
		{
			name:     "different content",
			a:        []NestedModel{minimalNestedModel()},
			b:        []NestedModel{fullNestedModel()},
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
	a := NestedModel{Name: types.StringValue("a")}
	b := NestedModel{Name: types.StringValue("b")}

	result := MatchOrder([]NestedModel{b, a}, []NestedModel{a, b})

	assert.Equal(t, []NestedModel{a, b}, result)
}
