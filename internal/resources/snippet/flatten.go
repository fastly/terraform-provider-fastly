package snippet

import (
	"context"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func FlattenToNestedModel(api *fastly.Snippet) (NestedModel, error) {
	if api == nil {
		return NestedModel{}, nil
	}

	priority, err := parsePriority(api.Priority)
	if err != nil {
		return NestedModel{}, err
	}

	return NestedModel{
		Name:     types.StringValue(fastly.ToValue(api.Name)),
		Type:     types.StringValue(string(fastly.ToValue(api.Type))),
		Priority: types.Int64Value(priority),
		Content:  types.StringValue(fastly.ToValue(api.Content)),
	}, nil
}

func flatten(ctx context.Context, s *fastly.Snippet, m *Model) error {
	if s == nil {
		tflog.Warn(ctx, "flatten called with nil VCL snippet")
		return nil
	}

	nestedModel, err := FlattenToNestedModel(s)
	if err != nil {
		return err
	}

	m.ID = types.StringValue(ID(fastly.ToValue(s.ServiceID), fastly.ToValue(s.ServiceVersion), fastly.ToValue(s.Name)))
	m.Service = types.StringValue(fastly.ToValue(s.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(s.ServiceVersion)))
	m.NestedModel = nestedModel

	tflog.Debug(ctx, "Flattened VCL snippet state", map[string]any{
		"id":      m.ID.ValueString(),
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})

	return nil
}
