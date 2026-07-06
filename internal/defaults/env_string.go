package defaults

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvString returns a default handler that reads from an environment variable,
// falling back to fallback if the variable is unset or empty.
func EnvString(envVar, fallback string) defaults.String {
	return envStringDefault{envVar: envVar, fallback: fallback}
}

type envStringDefault struct {
	envVar   string
	fallback string
}

func (d envStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to the %s environment variable, or %q if unset", d.envVar, d.fallback)
}

func (d envStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value defaults to the `%s` environment variable, or `%s` if unset", d.envVar, d.fallback)
}

func (d envStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	if v := os.Getenv(d.envVar); v != "" {
		resp.PlanValue = types.StringValue(v)
		return
	}
	resp.PlanValue = types.StringValue(d.fallback)
}
