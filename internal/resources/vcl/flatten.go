package vcl

import (
	"context"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func flatten(ctx context.Context, v *fastly.VCL, m *Model) {
	if v == nil {
		tflog.Warn(ctx, "flatten called with nil VCL")
		return
	}

	m.ID = types.StringValue(ID(fastly.ToValue(v.ServiceID), fastly.ToValue(v.ServiceVersion), fastly.ToValue(v.Name)))
	m.Service = types.StringValue(fastly.ToValue(v.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(v.ServiceVersion)))
	m.NestedModel = FlattenToNestedModel(v)

	tflog.Debug(ctx, "Flattened custom VCL state", map[string]any{
		"id":      m.ID.ValueString(),
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
		"main":    m.Main.ValueBool(),
	})
}
