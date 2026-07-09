package resourcelink

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func FlattenToNestedModel(a *fastly.Resource) NestedModel {
	m := NestedModel{}

	if a == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(a.Name))
	m.ResourceID = types.StringValue(fastly.ToValue(a.ResourceID))
	m.LinkID = types.StringValue(fastly.ToValue(a.LinkID))

	return m
}

func flatten(ctx context.Context, a *fastly.Resource, m *Model) {
	if a == nil {
		tflog.Warn(ctx, "flatten called with nil resource link")
		return
	}

	id := fastly.ToValue(a.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(a.ServiceVersion)) + "-" + fastly.ToValue(a.LinkID)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(a.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(a.ServiceVersion)))
	m.NestedModel = FlattenToNestedModel(a)

	tflog.Debug(ctx, "Flattened resource link state", map[string]any{
		"id":          id,
		"service":     m.Service.ValueString(),
		"version":     m.Version.ValueInt64(),
		"name":        m.Name.ValueString(),
		"resource_id": m.ResourceID.ValueString(),
		"link_id":     m.LinkID.ValueString(),
	})
}
