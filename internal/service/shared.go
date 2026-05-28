package service

import (
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	TypeVCL     = "vcl"
	TypeCompute = "wasm"
)

var nonIdentifierRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func TypeLabel(serviceType string) string {
	switch serviceType {
	case TypeVCL:
		return "CDN"
	case TypeCompute:
		return "Compute"
	default:
		return serviceType
	}
}

func TypeSupported(serviceType string, supportedTypes ...string) bool {
	for _, supported := range supportedTypes {
		if serviceType == supported {
			return true
		}
	}
	return false
}

func SupportedTypeLabels(supportedTypes []string) string {
	if len(supportedTypes) == 0 {
		return ""
	}

	out := ""
	for i, supported := range supportedTypes {
		if i > 0 {
			out += ", "
		}
		out += TypeLabel(supported)
	}
	return out
}

// ToIdentifier converts a generated name to a valid HCL identifier.
func ToIdentifier(name string) string {
	id := nonIdentifierRe.ReplaceAllString(name, "_")
	id = strings.ToLower(id)
	if len(id) > 0 && id[0] >= '0' && id[0] <= '9' {
		id = "_" + id
	}
	return id
}

func ToGeneratedResourceName(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}

	id := ToIdentifier(strings.Join(nonEmpty, "_"))
	if id == "" || id == "_" {
		return "resource"
	}
	return id
}

func StringValue(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func Int64Value(v types.Int64) int64 {
	if v.IsNull() || v.IsUnknown() {
		return 0
	}
	return v.ValueInt64()
}

func BoolValue(v types.Bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	return v.ValueBool()
}

func StringPointerOrDefault(v *string, fallback string) types.String {
	if v == nil {
		return types.StringValue(fallback)
	}
	return types.StringValue(*v)
}

func Int64PointerOrDefault(v *int, fallback int64) types.Int64 {
	if v == nil {
		return types.Int64Value(fallback)
	}
	return types.Int64Value(int64(*v))
}

func Int64PointerOrNull(v *int) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}

func BoolPointerOrDefault(v *bool, defaultValue bool) types.Bool {
	if v == nil {
		return types.BoolValue(defaultValue)
	}
	return types.BoolValue(*v)
}
