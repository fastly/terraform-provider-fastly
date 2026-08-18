package header

import (
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:              types.StringValue("header"),
		Action:            types.StringValue("delete"),
		Type:              types.StringValue("cache"),
		Destination:       types.StringValue("http.aws-id"),
		CacheCondition:    types.StringValue(""),
		IgnoreIfSet:       types.BoolValue(false),
		Priority:          types.Int64Value(100),
		Regex:             types.StringValue(""),
		RequestCondition:  types.StringValue(""),
		ResponseCondition: types.StringValue(""),
		Source:            types.StringValue(""),
		Substitution:      types.StringValue(""),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:              types.StringValue("header"),
		Action:            types.StringValue("regex"),
		Type:              types.StringValue("request"),
		Destination:       types.StringValue("http.X-Custom"),
		CacheCondition:    types.StringValue("cache-condition"),
		IgnoreIfSet:       types.BoolValue(true),
		Priority:          types.Int64Value(10),
		Regex:             types.StringValue("^foo"),
		RequestCondition:  types.StringValue("request-condition"),
		ResponseCondition: types.StringValue("response-condition"),
		Source:            types.StringValue("http.server-name"),
		Substitution:      types.StringValue("bar"),
	}
}

func fullAPIHeader() *fastly.Header {
	action := fastly.HeaderActionRegex
	headerType := fastly.HeaderTypeRequest
	return &fastly.Header{
		Name:              new("header"),
		Action:            &action,
		Type:              &headerType,
		Destination:       new("http.X-Custom"),
		CacheCondition:    new("cache-condition"),
		IgnoreIfSet:       new(true),
		Priority:          new(10),
		Regex:             new("^foo"),
		RequestCondition:  new("request-condition"),
		ResponseCondition: new("response-condition"),
		Source:            new("http.server-name"),
		Substitution:      new("bar"),
	}
}

func TestToModel(t *testing.T) {
	result := FlattenToNestedModel(fullAPIHeader())

	assert.True(t, fullNestedModel().ModelsEqual(result))
}

func TestOpsEqual(t *testing.T) {
	assert.True(t, ops{}.Equal(fullNestedModel(), fullAPIHeader()))
}

func TestOpsEqual_mismatch(t *testing.T) {
	remote := fullAPIHeader()
	remote.Priority = new(999)

	assert.False(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestBuildCreateInput(t *testing.T) {
	input := BuildCreateInput("service-id", 3, fullNestedModel())

	assert.Equal(t, "service-id", input.ServiceID)
	assert.Equal(t, 3, input.ServiceVersion)
	assert.Equal(t, "header", *input.Name)
	assert.Equal(t, fastly.HeaderActionRegex, *input.Action)
	assert.Equal(t, fastly.HeaderTypeRequest, *input.Type)
	assert.Equal(t, "http.X-Custom", *input.Destination)
	assert.Equal(t, "cache-condition", *input.CacheCondition)
	assert.True(t, bool(*input.IgnoreIfSet))
	assert.Equal(t, 10, *input.Priority)
	assert.Equal(t, "^foo", *input.Regex)
	assert.Equal(t, "request-condition", *input.RequestCondition)
	assert.Equal(t, "response-condition", *input.ResponseCondition)
	assert.Equal(t, "http.server-name", *input.Source)
	assert.Equal(t, "bar", *input.Substitution)
}

func TestBuildUpdateInput(t *testing.T) {
	input := BuildUpdateInput("service-id", 3, fullNestedModel())

	assert.Equal(t, "service-id", input.ServiceID)
	assert.Equal(t, 3, input.ServiceVersion)
	assert.Equal(t, "header", input.Name)
	assert.Equal(t, fastly.HeaderActionRegex, *input.Action)
	assert.Equal(t, fastly.HeaderTypeRequest, *input.Type)
	assert.Equal(t, "http.X-Custom", *input.Destination)
	assert.Equal(t, "cache-condition", *input.CacheCondition)
	assert.True(t, bool(*input.IgnoreIfSet))
	assert.Equal(t, 10, *input.Priority)
	assert.Equal(t, "^foo", *input.Regex)
	assert.Equal(t, "request-condition", *input.RequestCondition)
	assert.Equal(t, "response-condition", *input.ResponseCondition)
	assert.Equal(t, "http.server-name", *input.Source)
	assert.Equal(t, "bar", *input.Substitution)
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
