package condition

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(c *fastly.Condition) NestedModel {
	m := NestedModel{}

	if c == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(c.Name))
	m.Type = types.StringValue(fastly.ToValue(c.Type))
	m.Statement = types.StringValue(fastly.ToValue(c.Statement))
	m.Priority = service.Int64PointerOrDefault(c.Priority, DefaultPriority)

	return m
}

func flatten(ctx context.Context, c *fastly.Condition, m *Model) {
	if c == nil {
		tflog.Warn(ctx, "flatten called with nil condition")
		return
	}

	id := fastly.ToValue(c.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(c.ServiceVersion)) + "-" + fastly.ToValue(c.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(c.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(c.ServiceVersion)))

	m.NestedModel = FlattenToNestedModel(c)

	tflog.Debug(ctx, "Flattened service condition state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
