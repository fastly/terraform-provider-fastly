package defaults

import (
	"context"
	"testing"

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
