package vcl

import (
	"fmt"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultMain = false
)

// ContentEqual compares VCL source using Fastly's harmless trailing-newline
// round-trip behavior as equivalent while preserving all internal whitespace.
//
// The provider must not rewrite configured VCL source during planning, because
// content is a user-configured string. This helper is only used for provider-side
// reconciliation decisions so an API round trip that differs only by trailing
// newlines does not force an unnecessary service-version clone.
func ContentEqual(a, b string) bool {
	return strings.TrimRight(a, "\n") == strings.TrimRight(b, "\n")
}

func FlattenToNestedModel(api *fastly.VCL) NestedModel {
	if api == nil {
		return NestedModel{}
	}

	return NestedModel{
		Name:    types.StringValue(fastly.ToValue(api.Name)),
		Content: types.StringValue(fastly.ToValue(api.Content)),
		Main:    service.BoolPointerOrDefault(api.Main, DefaultMain),
	}
}

// ValidateConfig performs plan-time custom VCL validation while tolerating
// unknown values. Terraform's schema validation will handle required/null
// attributes, and apply-time Validate handles fully known values before API
// reconciliation.
func ValidateConfig(vcls []NestedModel) error {
	if len(vcls) == 0 {
		return nil
	}

	seenNames := make(map[string]struct{}, len(vcls))
	mainCount := 0
	allMainKnown := true

	for _, item := range vcls {
		if item.Main.IsUnknown() || item.Main.IsNull() {
			allMainKnown = false
		} else if item.Main.ValueBool() {
			mainCount++
		}

		if item.Name.IsUnknown() || item.Name.IsNull() {
			continue
		}

		name := strings.TrimSpace(item.Name.ValueString())
		if name == "" {
			return fmt.Errorf("custom VCL name cannot be empty")
		}

		if _, ok := seenNames[name]; ok {
			return fmt.Errorf("duplicate custom VCL name %q; names must be unique within a service version", name)
		}
		seenNames[name] = struct{}{}
	}

	if mainCount > 1 {
		return fmt.Errorf("only one custom VCL file can have main = true")
	}

	if allMainKnown && mainCount == 0 {
		return fmt.Errorf("one custom VCL file must have main = true")
	}

	return nil
}

func Validate(vcls []NestedModel) error {
	if len(vcls) == 0 {
		return nil
	}

	seenNames := make(map[string]struct{}, len(vcls))
	mainCount := 0

	for _, item := range vcls {
		name := strings.TrimSpace(service.StringValue(item.Name))
		if name == "" {
			return fmt.Errorf("custom VCL name cannot be empty")
		}

		if _, ok := seenNames[name]; ok {
			return fmt.Errorf("duplicate custom VCL name %q; names must be unique within a service version", name)
		}
		seenNames[name] = struct{}{}

		if service.BoolValue(item.Main) {
			mainCount++
		}
	}

	if mainCount == 0 {
		return fmt.Errorf("one custom VCL file must have main = true")
	}

	if mainCount > 1 {
		return fmt.Errorf("only one custom VCL file can have main = true")
	}

	return nil
}

func ID(serviceID string, version int, name string) string {
	return fmt.Sprintf("%s-%d-%s", serviceID, version, name)
}

func MatchOrderPreservePlanContent(items, plan []NestedModel) []NestedModel {
	ordered := MatchOrder(items, plan)

	planByName := make(map[string]NestedModel, len(plan))
	for _, item := range plan {
		planByName[service.StringValue(item.Name)] = item
	}

	for i := range ordered {
		name := service.StringValue(ordered[i].Name)
		if planned, ok := planByName[name]; ok {
			ordered[i].Content = planned.Content
		}
	}

	return ordered
}
