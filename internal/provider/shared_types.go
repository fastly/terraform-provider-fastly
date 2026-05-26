package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func stringValue(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func int64Value(v types.Int64) int64 {
	if v.IsNull() || v.IsUnknown() {
		return 0
	}
	return v.ValueInt64()
}

func boolValue(v types.Bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	return v.ValueBool()
}

func stringPointerOrDefault(v *string, fallback string) types.String {
	if v == nil {
		return types.StringValue(fallback)
	}
	return types.StringValue(*v)
}

func int64PointerOrDefault(v *int, fallback int64) types.Int64 {
	if v == nil {
		return types.Int64Value(fallback)
	}
	return types.Int64Value(int64(*v))
}

func int64PointerOrNull(v *int) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}

func boolPointerOrDefault(v *bool, defaultValue bool) types.Bool {
	if v == nil {
		return types.BoolValue(defaultValue)
	}
	return types.BoolValue(*v)
}
