package defaults

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEnvString(t *testing.T) {
	const envVar = "TEST_ENVSTRING_VAR"

	t.Run("uses fallback when env var unset", func(t *testing.T) {
		t.Setenv(envVar, "")
		resp := &defaults.StringResponse{}
		EnvString(envVar, "fallback").DefaultString(context.Background(), defaults.StringRequest{}, resp)
		if got := resp.PlanValue; got != types.StringValue("fallback") {
			t.Fatalf("expected fallback, got %s", got)
		}
	})

	t.Run("uses env var when set", func(t *testing.T) {
		t.Setenv(envVar, "from-env")
		resp := &defaults.StringResponse{}
		EnvString(envVar, "fallback").DefaultString(context.Background(), defaults.StringRequest{}, resp)
		if got := resp.PlanValue; got != types.StringValue("from-env") {
			t.Fatalf("expected from-env, got %s", got)
		}
	})
}

func TestEnvStringDeprecatedFallback(t *testing.T) {
	const envVar = "TEST_ENVSTRING_NEW_VAR"
	const deprecatedEnvVar = "TEST_ENVSTRING_OLD_VAR"

	t.Run("uses fallback when neither env var is set", func(t *testing.T) {
		t.Setenv(envVar, "")
		t.Setenv(deprecatedEnvVar, "")
		resp := &defaults.StringResponse{}
		EnvStringDeprecatedFallback(envVar, deprecatedEnvVar, "fallback").DefaultString(context.Background(), defaults.StringRequest{}, resp)
		if got := resp.PlanValue; got != types.StringValue("fallback") {
			t.Fatalf("expected fallback, got %s", got)
		}
		if resp.Diagnostics.HasError() || len(resp.Diagnostics) != 0 {
			t.Fatalf("expected no diagnostics, got %v", resp.Diagnostics)
		}
	})

	t.Run("uses new env var when set, no warning", func(t *testing.T) {
		t.Setenv(envVar, "from-new")
		t.Setenv(deprecatedEnvVar, "from-old")
		resp := &defaults.StringResponse{}
		EnvStringDeprecatedFallback(envVar, deprecatedEnvVar, "fallback").DefaultString(context.Background(), defaults.StringRequest{}, resp)
		if got := resp.PlanValue; got != types.StringValue("from-new") {
			t.Fatalf("expected from-new, got %s", got)
		}
		if len(resp.Diagnostics) != 0 {
			t.Fatalf("expected no diagnostics, got %v", resp.Diagnostics)
		}
	})

	t.Run("falls back to deprecated env var with a warning", func(t *testing.T) {
		t.Setenv(envVar, "")
		t.Setenv(deprecatedEnvVar, "from-old")
		resp := &defaults.StringResponse{}
		EnvStringDeprecatedFallback(envVar, deprecatedEnvVar, "fallback").DefaultString(context.Background(), defaults.StringRequest{}, resp)
		if got := resp.PlanValue; got != types.StringValue("from-old") {
			t.Fatalf("expected from-old, got %s", got)
		}
		if len(resp.Diagnostics) != 1 || resp.Diagnostics[0].Severity() != diag.SeverityWarning {
			t.Fatalf("expected exactly one warning diagnostic, got %v", resp.Diagnostics)
		}
	})
}
