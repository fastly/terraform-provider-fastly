package requestsetting

import (
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:             types.StringValue("request-setting"),
		Action:           types.StringNull(),
		BypassBusyWait:   types.BoolValue(false),
		DefaultHost:      types.StringValue(""),
		ForceMiss:        types.BoolValue(false),
		ForceSSL:         types.BoolValue(false),
		HashKeys:         types.StringValue(""),
		MaxStaleAge:      types.Int64Value(0),
		RequestCondition: types.StringValue(""),
		TimerSupport:     types.BoolValue(false),
		XFF:              types.StringNull(),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:             types.StringValue("request-setting"),
		Action:           types.StringValue("lookup"),
		BypassBusyWait:   types.BoolValue(true),
		DefaultHost:      types.StringValue("host.example.com"),
		ForceMiss:        types.BoolValue(true),
		ForceSSL:         types.BoolValue(true),
		HashKeys:         types.StringValue("field1,field2"),
		MaxStaleAge:      types.Int64Value(120),
		RequestCondition: types.StringValue("request-condition"),
		TimerSupport:     types.BoolValue(true),
		XFF:              types.StringValue("append"),
	}
}

func TestToModel(t *testing.T) {
	action := fastly.RequestSettingActionLookup
	xff := fastly.RequestSettingXFFAppend
	api := &fastly.RequestSetting{
		Name:             new("request-setting"),
		Action:           &action,
		BypassBusyWait:   new(true),
		DefaultHost:      new("host.example.com"),
		ForceMiss:        new(true),
		ForceSSL:         new(true),
		HashKeys:         new("field1,field2"),
		MaxStaleAge:      new(120),
		RequestCondition: new("request-condition"),
		TimerSupport:     new(true),
		XForwardedFor:    &xff,
	}

	result := ops{}.ToModel(api)

	assert.True(t, fullNestedModel().ModelsEqual(result))
}

func TestToModel_emptyOptionalFields(t *testing.T) {
	emptyAction := fastly.RequestSettingAction("")
	emptyXFF := fastly.RequestSettingXFF("")
	api := &fastly.RequestSetting{
		Name:             new("request-setting"),
		Action:           &emptyAction,
		BypassBusyWait:   new(false),
		DefaultHost:      new(""),
		ForceMiss:        new(false),
		ForceSSL:         new(false),
		HashKeys:         new(""),
		MaxStaleAge:      new(0),
		RequestCondition: new(""),
		TimerSupport:     new(false),
		XForwardedFor:    &emptyXFF,
	}

	result := ops{}.ToModel(api)

	assert.Equal(t, "request-setting", result.Name.ValueString())
	assert.True(t, result.Action.IsNull())
	assert.True(t, result.XFF.IsNull())
	assert.Equal(t, "", result.DefaultHost.ValueString())
	assert.Equal(t, "", result.HashKeys.ValueString())
	assert.Equal(t, "", result.RequestCondition.ValueString())
	assert.Equal(t, int64(0), result.MaxStaleAge.ValueInt64())
}

func TestToModel_nilOptionalFields(t *testing.T) {
	api := &fastly.RequestSetting{
		Name:        new("request-setting"),
		MaxStaleAge: new(0),
	}

	result := ops{}.ToModel(api)

	assert.True(t, result.Action.IsNull())
	assert.True(t, result.XFF.IsNull())
	assert.Equal(t, "", result.DefaultHost.ValueString())
	assert.Equal(t, "", result.HashKeys.ValueString())
	assert.Equal(t, "", result.RequestCondition.ValueString())
}

func TestOpsEqual(t *testing.T) {
	action := fastly.RequestSettingActionLookup
	xff := fastly.RequestSettingXFFAppend
	remote := &fastly.RequestSetting{
		Name:             new("request-setting"),
		Action:           &action,
		BypassBusyWait:   new(true),
		DefaultHost:      new("host.example.com"),
		ForceMiss:        new(true),
		ForceSSL:         new(true),
		HashKeys:         new("field1,field2"),
		MaxStaleAge:      new(120),
		RequestCondition: new("request-condition"),
		TimerSupport:     new(true),
		XForwardedFor:    &xff,
	}

	assert.True(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestOpsEqual_mismatch(t *testing.T) {
	remote := &fastly.RequestSetting{
		Name:        new("request-setting"),
		MaxStaleAge: new(0),
	}

	assert.False(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestActionPointer(t *testing.T) {
	tests := []struct {
		name     string
		value    types.String
		expected *fastly.RequestSettingAction
	}{
		{name: "null", value: types.StringNull(), expected: nil},
		{name: "unknown", value: types.StringUnknown(), expected: nil},
		{name: "empty", value: types.StringValue(""), expected: nil},
		{name: "lookup", value: types.StringValue("lookup"), expected: new(fastly.RequestSettingActionLookup)},
		{name: "uppercase", value: types.StringValue("PASS"), expected: new(fastly.RequestSettingActionPass)},
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

func TestXFFPointer(t *testing.T) {
	tests := []struct {
		name     string
		value    types.String
		expected *fastly.RequestSettingXFF
	}{
		{name: "null", value: types.StringNull(), expected: nil},
		{name: "unknown", value: types.StringUnknown(), expected: nil},
		{name: "empty", value: types.StringValue(""), expected: nil},
		{name: "append", value: types.StringValue("append"), expected: new(fastly.RequestSettingXFFAppend)},
		{name: "uppercase", value: types.StringValue("CLEAR"), expected: new(fastly.RequestSettingXFFClear)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := xffPointer(tt.value)
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
