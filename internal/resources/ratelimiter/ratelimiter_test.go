package ratelimiter

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/terraform-provider-fastly/internal/resources/dictionary"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:               types.StringValue("rate-limiter"),
		Action:             types.StringValue("log_only"),
		ClientKey:          stringListValue("req.http.Fastly-Client-IP"),
		FeatureRevision:    types.Int64Value(1),
		HTTPMethods:        stringListValue("GET"),
		LoggerType:         types.StringNull(),
		PenaltyBoxDuration: types.Int64Value(10),
		RateLimiterID:      types.StringValue("abc123"),
		Response:           types.ObjectNull(responseAttributeTypes),
		ResponseObjectName: types.StringNull(),
		RpsLimit:           types.Int64Value(100),
		URIDictionaryName:  types.StringNull(),
		WindowSize:         types.Int64Value(60),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:               types.StringValue("rate-limiter"),
		Action:             types.StringValue("response"),
		ClientKey:          stringListValue("req.http.Fastly-Client-IP", "req.http.User-Agent"),
		FeatureRevision:    types.Int64Value(2),
		HTTPMethods:        stringListValue("GET", "POST"),
		LoggerType:         types.StringValue("s3"),
		PenaltyBoxDuration: types.Int64Value(20),
		RateLimiterID:      types.StringValue("abc123"),
		Response: NewResponseObject(
			types.StringValue("rate limited"),
			types.StringValue("text/plain"),
			types.Int64Value(429),
		),
		ResponseObjectName: types.StringValue("my-response-object"),
		RpsLimit:           types.Int64Value(500),
		URIDictionaryName:  types.StringValue("my-dictionary"),
		WindowSize:         types.Int64Value(10),
	}
}

func stringListValue(values ...string) types.List {
	elems := make([]attr.Value, len(values))
	for i, v := range values {
		elems[i] = types.StringValue(v)
	}
	return types.ListValueMust(types.StringType, elems)
}

func TestToModel(t *testing.T) {
	action := fastly.ERLActionResponse
	loggerType := fastly.ERLLogS3
	windowSize := fastly.ERLSize10
	api := &fastly.ERL{
		Name:               new("rate-limiter"),
		Action:             &action,
		ClientKey:          []*string{new("req.http.Fastly-Client-IP"), new("req.http.User-Agent")},
		FeatureRevision:    new(2),
		HTTPMethods:        []*string{new("GET"), new("POST")},
		LoggerType:         &loggerType,
		PenaltyBoxDuration: new(20),
		RateLimiterID:      new("abc123"),
		Response: &fastly.ERLResponse{
			ERLContent:     new("rate limited"),
			ERLContentType: new("text/plain"),
			ERLStatus:      new(429),
		},
		ResponseObjectName: new("my-response-object"),
		RpsLimit:           new(500),
		URIDictionaryName:  new("my-dictionary"),
		WindowSize:         &windowSize,
	}

	result := ops{}.ToModel(api)

	assert.True(t, fullNestedModel().ModelsEqual(result))
}

func TestToModel_emptyOptionalFields(t *testing.T) {
	action := fastly.ERLActionLogOnly
	windowSize := fastly.ERLSize60
	api := &fastly.ERL{
		Name:               new("rate-limiter"),
		Action:             &action,
		ClientKey:          []*string{new("req.http.Fastly-Client-IP")},
		FeatureRevision:    new(1),
		HTTPMethods:        []*string{new("GET")},
		PenaltyBoxDuration: new(10),
		RateLimiterID:      new("abc123"),
		RpsLimit:           new(100),
		WindowSize:         &windowSize,
	}

	result := ops{}.ToModel(api)

	assert.Equal(t, "rate-limiter", result.Name.ValueString())
	assert.True(t, result.LoggerType.IsNull())
	assert.True(t, result.ResponseObjectName.IsNull())
	assert.True(t, result.URIDictionaryName.IsNull())
	assert.True(t, result.Response.IsNull())
}

func TestOpsEqual(t *testing.T) {
	action := fastly.ERLActionResponse
	loggerType := fastly.ERLLogS3
	windowSize := fastly.ERLSize10
	remote := &fastly.ERL{
		Name:               new("rate-limiter"),
		Action:             &action,
		ClientKey:          []*string{new("req.http.Fastly-Client-IP"), new("req.http.User-Agent")},
		FeatureRevision:    new(2),
		HTTPMethods:        []*string{new("GET"), new("POST")},
		LoggerType:         &loggerType,
		PenaltyBoxDuration: new(20),
		RateLimiterID:      new("abc123"),
		Response: &fastly.ERLResponse{
			ERLContent:     new("rate limited"),
			ERLContentType: new("text/plain"),
			ERLStatus:      new(429),
		},
		ResponseObjectName: new("my-response-object"),
		RpsLimit:           new(500),
		URIDictionaryName:  new("my-dictionary"),
		WindowSize:         &windowSize,
	}

	assert.True(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestOpsEqual_caseInsensitiveAction(t *testing.T) {
	action := fastly.ERLActionLogOnly
	windowSize := fastly.ERLSize60
	remote := &fastly.ERL{
		Name:               new("rate-limiter"),
		Action:             &action,
		ClientKey:          []*string{new("req.http.Fastly-Client-IP")},
		FeatureRevision:    new(1),
		HTTPMethods:        []*string{new("GET")},
		PenaltyBoxDuration: new(10),
		RateLimiterID:      new("abc123"),
		RpsLimit:           new(100),
		WindowSize:         &windowSize,
	}

	desired := minimalNestedModel()
	desired.Action = types.StringValue("LOG_ONLY")

	assert.True(t, ops{}.Equal(desired, remote))
}

func TestOpsEqual_mismatch(t *testing.T) {
	action := fastly.ERLActionLogOnly
	windowSize := fastly.ERLSize60
	remote := &fastly.ERL{
		Name:               new("rate-limiter"),
		Action:             &action,
		ClientKey:          []*string{new("req.http.Fastly-Client-IP")},
		FeatureRevision:    new(1),
		HTTPMethods:        []*string{new("GET")},
		PenaltyBoxDuration: new(10),
		RateLimiterID:      new("abc123"),
		RpsLimit:           new(100),
		WindowSize:         &windowSize,
	}

	assert.False(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestActionPointer(t *testing.T) {
	tests := []struct {
		name     string
		value    types.String
		expected *fastly.ERLAction
	}{
		{name: "null", value: types.StringNull(), expected: nil},
		{name: "unknown", value: types.StringUnknown(), expected: nil},
		{name: "empty", value: types.StringValue(""), expected: nil},
		{name: "log_only", value: types.StringValue("log_only"), expected: new(fastly.ERLActionLogOnly)},
		{name: "uppercase", value: types.StringValue("RESPONSE"), expected: new(fastly.ERLActionResponse)},
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

func TestLoggerTypePointer(t *testing.T) {
	tests := []struct {
		name     string
		value    types.String
		expected *fastly.ERLLogger
	}{
		{name: "null", value: types.StringNull(), expected: nil},
		{name: "empty", value: types.StringValue(""), expected: nil},
		{name: "s3", value: types.StringValue("s3"), expected: new(fastly.ERLLogS3)},
		{name: "uppercase", value: types.StringValue("BIGQUERY"), expected: new(fastly.ERLLogBigQuery)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loggerTypePointer(tt.value)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else if assert.NotNil(t, result) {
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestWindowSizePointer(t *testing.T) {
	result := windowSizePointer(types.Int64Value(10))
	assert.NotNil(t, result)
	assert.Equal(t, fastly.ERLSize10, *result)
}

func TestResponseType(t *testing.T) {
	t.Run("null object", func(t *testing.T) {
		result := responseType(types.ObjectNull(responseAttributeTypes))
		assert.Nil(t, result)
	})

	t.Run("configured object", func(t *testing.T) {
		obj := NewResponseObject(types.StringValue("body"), types.StringValue("application/json"), types.Int64Value(429))
		result := responseType(obj)
		if assert.NotNil(t, result) {
			assert.Equal(t, "body", *result.ERLContent)
			assert.Equal(t, "application/json", *result.ERLContentType)
			assert.Equal(t, 429, *result.ERLStatus)
		}
	})
}

func TestStringSliceToList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		result := stringSliceToList(nil)
		assert.True(t, result.IsNull())
	})

	t.Run("values", func(t *testing.T) {
		result := stringSliceToList([]*string{new("a"), new("b")})
		assert.Equal(t, []string{"a", "b"}, *listToStringSlice(result))
	})
}

func TestListToStringSlice(t *testing.T) {
	l := stringListValue("a", "b")
	assert.Equal(t, []string{"a", "b"}, *listToStringSlice(l))
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

func TestOpsDelete_noMatchingID(t *testing.T) {
	o := &ops{}

	err := o.Delete(context.Background(), nil, "service-id", 1, "missing")

	assert.NoError(t, err)
}

func TestNeedsRecreate(t *testing.T) {
	tests := []struct {
		name     string
		desired  NestedModel
		remote   *fastly.ERL
		expected bool
	}{
		{
			name:     "no remote",
			desired:  minimalNestedModel(),
			remote:   nil,
			expected: false,
		},
		{
			name:    "uri_dictionary_name cleared",
			desired: minimalNestedModel(),
			remote: &fastly.ERL{
				URIDictionaryName: new("my-dictionary"),
			},
			expected: true,
		},
		{
			name:    "response_object_name cleared",
			desired: minimalNestedModel(),
			remote: &fastly.ERL{
				ResponseObjectName: new("my-response-object"),
			},
			expected: true,
		},
		{
			name:    "uri_dictionary_name unchanged",
			desired: fullNestedModel(),
			remote: &fastly.ERL{
				URIDictionaryName:  new("my-dictionary"),
				ResponseObjectName: new("my-response-object"),
			},
			expected: false,
		},
		{
			name:     "neither field was ever set",
			desired:  minimalNestedModel(),
			remote:   &fastly.ERL{},
			expected: false,
		},
		{
			name:    "response cleared",
			desired: minimalNestedModel(),
			remote: &fastly.ERL{
				Response: &fastly.ERLResponse{
					ERLContent:     new("rate limited"),
					ERLContentType: new("text/plain"),
					ERLStatus:      new(429),
				},
			},
			expected: true,
		},
		{
			name:    "response unchanged",
			desired: fullNestedModel(),
			remote: &fastly.ERL{
				Response: &fastly.ERLResponse{
					ERLContent:     new("rate limited"),
					ERLContentType: new("text/plain"),
					ERLStatus:      new(429),
				},
			},
			expected: false,
		},
		{
			name:    "remote response missing a sub-field doesn't count as set",
			desired: minimalNestedModel(),
			remote: &fastly.ERL{
				Response: &fastly.ERLResponse{
					ERLContent: new("rate limited"),
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, needsRecreate(tt.desired, tt.remote))
		})
	}
}

func TestCaseInsensitiveState_PlanModifyString(t *testing.T) {
	tests := []struct {
		name     string
		state    types.String
		config   types.String
		expected types.String
	}{
		{
			name:     "config differs only in case from state",
			state:    types.StringValue("response"),
			config:   types.StringValue("RESPONSE"),
			expected: types.StringValue("response"),
		},
		{
			name:     "config matches state exactly",
			state:    types.StringValue("response"),
			config:   types.StringValue("response"),
			expected: types.StringValue("response"),
		},
		{
			name:     "config is a different value",
			state:    types.StringValue("response"),
			config:   types.StringValue("log_only"),
			expected: types.StringValue("log_only"),
		},
		{
			name:     "no prior state",
			state:    types.StringNull(),
			config:   types.StringValue("RESPONSE"),
			expected: types.StringValue("RESPONSE"),
		},
		{
			name:     "unknown config value",
			state:    types.StringValue("response"),
			config:   types.StringUnknown(),
			expected: types.StringUnknown(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				StateValue:  tt.state,
				ConfigValue: tt.config,
			}
			resp := &planmodifier.StringResponse{PlanValue: tt.config}

			caseInsensitiveState{}.PlanModifyString(context.Background(), req, resp)

			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	t.Run("unique names", func(t *testing.T) {
		first := minimalNestedModel()
		first.Name = types.StringValue("first")
		second := minimalNestedModel()
		second.Name = types.StringValue("second")

		assert.NoError(t, ValidateConfig([]NestedModel{first, second}))
	})

	t.Run("duplicate names", func(t *testing.T) {
		first := minimalNestedModel()
		first.Name = types.StringValue("duplicate")
		second := minimalNestedModel()
		second.Name = types.StringValue("duplicate")

		err := ValidateConfig([]NestedModel{first, second})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), `"duplicate"`)
		}
	})

	t.Run("skips unknown names", func(t *testing.T) {
		first := minimalNestedModel()
		first.Name = types.StringUnknown()
		second := minimalNestedModel()
		second.Name = types.StringUnknown()

		assert.NoError(t, ValidateConfig([]NestedModel{first, second}))
	})

	t.Run("response action with response block set", func(t *testing.T) {
		item := fullNestedModel()
		item.Action = types.StringValue("response")

		assert.NoError(t, ValidateConfig([]NestedModel{item}))
	})

	t.Run("response action missing response block", func(t *testing.T) {
		item := minimalNestedModel()
		item.Action = types.StringValue("RESPONSE")
		item.Response = types.ObjectNull(responseAttributeTypes)

		err := ValidateConfig([]NestedModel{item})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "response is required")
		}
	})

	t.Run("response_object action with name set", func(t *testing.T) {
		item := fullNestedModel()
		item.Action = types.StringValue("response_object")

		assert.NoError(t, ValidateConfig([]NestedModel{item}))
	})

	t.Run("response_object action missing name", func(t *testing.T) {
		item := minimalNestedModel()
		item.Action = types.StringValue("response_object")
		item.ResponseObjectName = types.StringNull()

		err := ValidateConfig([]NestedModel{item})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "response_object_name is required")
		}
	})

	t.Run("log_only action requires neither field", func(t *testing.T) {
		item := minimalNestedModel()
		item.Action = types.StringValue("log_only")

		assert.NoError(t, ValidateConfig([]NestedModel{item}))
	})
}

func TestValidateDictionaryReferences(t *testing.T) {
	dict := func(name string) dictionary.NestedModel {
		return dictionary.NestedModel{Name: types.StringValue(name)}
	}

	t.Run("no uri_dictionary_name set", func(t *testing.T) {
		item := minimalNestedModel()
		assert.NoError(t, ValidateDictionaryReferences([]NestedModel{item}, nil))
	})

	t.Run("references a configured dictionary", func(t *testing.T) {
		item := minimalNestedModel()
		item.URIDictionaryName = types.StringValue("my-dictionary")

		assert.NoError(t, ValidateDictionaryReferences([]NestedModel{item}, []dictionary.NestedModel{dict("my-dictionary")}))
	})

	t.Run("references a dictionary that isn't configured", func(t *testing.T) {
		item := minimalNestedModel()
		item.Name = types.StringValue("rate-limiter")
		item.URIDictionaryName = types.StringValue("missing-dictionary")

		err := ValidateDictionaryReferences([]NestedModel{item}, []dictionary.NestedModel{dict("other-dictionary")})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), `"rate-limiter"`)
			assert.Contains(t, err.Error(), `"missing-dictionary"`)
		}
	})

	t.Run("references a dictionary when none are configured", func(t *testing.T) {
		item := minimalNestedModel()
		item.URIDictionaryName = types.StringValue("missing-dictionary")

		err := ValidateDictionaryReferences([]NestedModel{item}, nil)
		assert.Error(t, err)
	})

	t.Run("skips unknown uri_dictionary_name", func(t *testing.T) {
		item := minimalNestedModel()
		item.URIDictionaryName = types.StringUnknown()

		assert.NoError(t, ValidateDictionaryReferences([]NestedModel{item}, nil))
	})
}

func TestMatchOrder(t *testing.T) {
	a := NestedModel{Name: types.StringValue("a")}
	b := NestedModel{Name: types.StringValue("b")}

	result := MatchOrder([]NestedModel{b, a}, []NestedModel{a, b})

	assert.Equal(t, []NestedModel{a, b}, result)
}
