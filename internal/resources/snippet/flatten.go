package snippet

import (
	"context"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func FlattenToNestedModel(api *fastly.Snippet) NestedModel {
	if api == nil {
		return NestedModel{}
	}

	return NestedModel{
		Name:     types.StringValue(fastly.ToValue(api.Name)),
		Type:     types.StringValue(string(fastly.ToValue(api.Type))),
		Priority: types.Int64Value(parsePriority(api.Priority)),
		Content:  types.StringValue(fastly.ToValue(api.Content)),
	}
}

func flatten(ctx context.Context, s *fastly.Snippet, m *Model) {
	if s == nil {
		tflog.Warn(ctx, "flatten called with nil VCL snippet")
		return
	}

	m.ID = types.StringValue(ID(fastly.ToValue(s.ServiceID), fastly.ToValue(s.ServiceVersion), fastly.ToValue(s.Name)))
	m.Service = types.StringValue(fastly.ToValue(s.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(s.ServiceVersion)))
	m.NestedModel = FlattenToNestedModel(s)

	tflog.Debug(ctx, "Flattened VCL snippet state", map[string]any{
		"id":      m.ID.ValueString(),
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
