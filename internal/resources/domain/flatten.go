package domain

import (
	"context"
	"strconv"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func flatten(ctx context.Context, d *fastly.Domain, m *Model) {
	if d == nil {
		tflog.Warn(ctx, "flatten called with nil domain")
		return
	}

	id := *d.ServiceID + "-" + strconv.Itoa(*d.ServiceVersion) + "-" + *d.Name
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(*d.ServiceID)
	m.Version = types.Int64Value(int64(*d.ServiceVersion))
	m.Name = types.StringValue(*d.Name)

	if d.Comment != nil && *d.Comment != "" {
		m.Comment = types.StringValue(*d.Comment)
	} else {
		m.Comment = types.StringNull()
	}

	tflog.Debug(ctx, "Flattened service domain state", map[string]any{
		"id":      id,
		"service": *d.ServiceID,
		"version": *d.ServiceVersion,
		"name":    *d.Name,
		"comment": d.Comment,
	})
}
