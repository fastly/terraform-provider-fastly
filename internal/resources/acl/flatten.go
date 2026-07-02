package acl

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func FlattenToNestedModel(a *fastly.ACL) NestedModel {
	m := NestedModel{}

	if a == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(a.Name))
	m.ACLID = types.StringValue(fastly.ToValue(a.ACLID))
	// ForceDestroy is configuration-only, not returned by API
	// Set to default value; preserved through ReadForVersionWithPlan
	m.ForceDestroy = types.BoolValue(DefaultForceDestroy)

	return m
}

func flatten(ctx context.Context, a *fastly.ACL, m *Model) {
	if a == nil {
		tflog.Warn(ctx, "flatten called with nil ACL")
		return
	}

	id := fastly.ToValue(a.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(a.ServiceVersion)) + "-" + fastly.ToValue(a.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(a.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(a.ServiceVersion)))

	// Preserve force_destroy from the model before flattening, since it's configuration-only
	forceDestroy := m.ForceDestroy
	m.NestedModel = FlattenToNestedModel(a)
	// If ForceDestroy was already set in the model (e.g., from prior state), preserve it
	if !forceDestroy.IsNull() {
		m.ForceDestroy = forceDestroy
	}

	tflog.Debug(ctx, "Flattened service ACL state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
		"acl_id":  m.ACLID.ValueString(),
	})
}
