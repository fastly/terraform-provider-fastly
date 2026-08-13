package healthcheck

import (
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:             types.StringValue("test_healthcheck"),
		Host:             types.StringValue("example.com"),
		Path:             types.StringValue("/"),
		CheckInterval:    types.Int64Value(DefaultCheckInterval),
		ExpectedResponse: types.Int64Value(DefaultExpectedResponse),
		Headers:          types.SetNull(types.StringType),
		HTTPVersion:      types.StringValue(DefaultHTTPVersion),
		Initial:          types.Int64Value(DefaultInitial),
		Method:           types.StringValue(DefaultMethod),
		Threshold:        types.Int64Value(DefaultThreshold),
		Timeout:          types.Int64Value(DefaultTimeout),
		Window:           types.Int64Value(DefaultWindow),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:             types.StringValue("test_healthcheck"),
		Host:             types.StringValue("example.com"),
		Path:             types.StringValue("/healthz"),
		CheckInterval:    types.Int64Value(1000),
		ExpectedResponse: types.Int64Value(201),
		Headers:          types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Api-Key: abc123")}),
		HTTPVersion:      types.StringValue("1.0"),
		Initial:          types.Int64Value(2),
		Method:           types.StringValue("GET"),
		Threshold:        types.Int64Value(2),
		Timeout:          types.Int64Value(3000),
		Window:           types.Int64Value(10),
	}
}

func TestToModel(t *testing.T) {
	tests := []struct {
		name     string
		api      *fastly.HealthCheck
		expected NestedModel
	}{
		{
			name: "fully populated",
			api: &fastly.HealthCheck{
				Name:             new("test_healthcheck"),
				Host:             new("example.com"),
				Path:             new("/healthz"),
				CheckInterval:    new(1000),
				ExpectedResponse: new(201),
				Headers:          []string{"b: 2", "a: 1"},
				HTTPVersion:      new("1.0"),
				Initial:          new(2),
				Method:           new("GET"),
				Threshold:        new(2),
				Timeout:          new(3000),
				Window:           new(10),
			},
			expected: NestedModel{
				Name:             types.StringValue("test_healthcheck"),
				Host:             types.StringValue("example.com"),
				Path:             types.StringValue("/healthz"),
				CheckInterval:    types.Int64Value(1000),
				ExpectedResponse: types.Int64Value(201),
				Headers:          types.SetValueMust(types.StringType, []attr.Value{types.StringValue("b: 2"), types.StringValue("a: 1")}),
				HTTPVersion:      types.StringValue("1.0"),
				Initial:          types.Int64Value(2),
				Method:           types.StringValue("GET"),
				Threshold:        types.Int64Value(2),
				Timeout:          types.Int64Value(3000),
				Window:           types.Int64Value(10),
			},
		},
		{
			// Optional fields the API omits (nil pointers, no headers) fall back to the
			// same defaults the schema applies, so a health check that was never given
			// explicit values doesn't show a permanent diff.
			name: "optional fields omitted by the API use defaults",
			api: &fastly.HealthCheck{
				Name: new("test_healthcheck"),
				Host: new("example.com"),
				Path: new("/"),
			},
			expected: minimalNestedModel(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ops{}.ToModel(tt.api)
			assert.True(t, tt.expected.ModelsEqual(result), "expected %+v, got %+v", tt.expected, result)
		})
	}
}

func TestOpsEqual(t *testing.T) {
	remote := &fastly.HealthCheck{
		Name:             new("test_healthcheck"),
		Host:             new("example.com"),
		Path:             new("/"),
		CheckInterval:    new(DefaultCheckInterval),
		ExpectedResponse: new(DefaultExpectedResponse),
		HTTPVersion:      new(DefaultHTTPVersion),
		Initial:          new(DefaultInitial),
		Method:           new(DefaultMethod),
		Threshold:        new(DefaultThreshold),
		Timeout:          new(DefaultTimeout),
		Window:           new(DefaultWindow),
	}

	tests := []struct {
		name     string
		desired  NestedModel
		expected bool
	}{
		{
			name:     "matching defaults",
			desired:  minimalNestedModel(),
			expected: true,
		},
		{
			name: "different threshold",
			desired: func() NestedModel {
				m := minimalNestedModel()
				m.Threshold = types.Int64Value(5)
				return m
			}(),
			expected: false,
		},
		{
			name: "different name",
			desired: func() NestedModel {
				m := minimalNestedModel()
				m.Name = types.StringValue("other_healthcheck")
				return m
			}(),
			expected: false,
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
			name:     "identical minimal models",
			a:        minimalNestedModel(),
			b:        minimalNestedModel(),
			expected: true,
		},
		{
			name:     "identical full models",
			a:        fullNestedModel(),
			b:        fullNestedModel(),
			expected: true,
		},
		{
			name: "different host",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Host = types.StringValue("a.example.com")
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Host = types.StringValue("b.example.com")
				return m
			}(),
			expected: false,
		},
		{
			name: "headers in different order are still equal",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Headers = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a: 1"), types.StringValue("b: 2")})
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Headers = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("b: 2"), types.StringValue("a: 1")})
				return m
			}(),
			expected: true,
		},
		{
			name: "different headers",
			a: func() NestedModel {
				m := minimalNestedModel()
				m.Headers = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a: 1")})
				return m
			}(),
			b: func() NestedModel {
				m := minimalNestedModel()
				m.Headers = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("b: 2")})
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
				{Name: types.StringValue("hc_b")},
				{Name: types.StringValue("hc_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("hc_a")},
				{Name: types.StringValue("hc_b")},
			},
			expected: true,
		},
		{
			name: "different lengths",
			a: []NestedModel{
				{Name: types.StringValue("hc_a")},
			},
			b: []NestedModel{
				{Name: types.StringValue("hc_a")},
				{Name: types.StringValue("hc_b")},
			},
			expected: false,
		},
		{
			name: "different content",
			a: []NestedModel{
				{Name: types.StringValue("hc_a"), Threshold: types.Int64Value(3)},
			},
			b: []NestedModel{
				{Name: types.StringValue("hc_a"), Threshold: types.Int64Value(5)},
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
			order:    []NestedModel{{Name: types.StringValue("hc_a")}},
			expected: []NestedModel{},
		},
		{
			name: "items match order exactly",
			items: []NestedModel{
				{Name: types.StringValue("hc_b"), Host: types.StringValue("b.example.com")},
				{Name: types.StringValue("hc_a"), Host: types.StringValue("a.example.com")},
			},
			order: []NestedModel{
				{Name: types.StringValue("hc_a")},
				{Name: types.StringValue("hc_b")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("hc_a"), Host: types.StringValue("a.example.com")},
				{Name: types.StringValue("hc_b"), Host: types.StringValue("b.example.com")},
			},
		},
		{
			name: "items not in order are appended",
			items: []NestedModel{
				{Name: types.StringValue("hc_a"), Host: types.StringValue("a.example.com")},
				{Name: types.StringValue("hc_c"), Host: types.StringValue("c.example.com")},
			},
			order: []NestedModel{
				{Name: types.StringValue("hc_a")},
			},
			expected: []NestedModel{
				{Name: types.StringValue("hc_a"), Host: types.StringValue("a.example.com")},
				{Name: types.StringValue("hc_c"), Host: types.StringValue("c.example.com")},
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

func TestHeadersRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
	}{
		{name: "nil headers", headers: nil},
		{name: "empty headers", headers: []string{}},
		{name: "single header", headers: []string{"a: 1"}},
		{name: "multiple headers", headers: []string{"a: 1", "b: 2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := headersFromSlice(tt.headers)
			result := headersToSlice(set)
			if len(tt.headers) == 0 {
				assert.Empty(t, result)
				assert.True(t, set.IsNull())
			} else {
				assert.ElementsMatch(t, tt.headers, result)
			}
		})
	}
}
