package loggingnewrelic

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(n *fastly.NewRelic) NestedModel {
	m := NestedModel{}

	if n == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(n.Name))
	m.Authentication = NewAuthenticationObject(
		service.StringPointerOrDefault(n.Token, ""),
	)
	m.Region = service.StringPointerOrDefault(n.Region, DefaultRegion)
	m.ProcessingRegion = service.StringPointerOrDefault(n.ProcessingRegion, DefaultProcessingRegion)
	m.Format = service.StringPointerOrDefault(n.Format, constants.LoggingNewRelicDefaultFormat)
	m.FormatVersion = service.Int64PointerOrDefault(n.FormatVersion, DefaultFormatVersion)
	m.Placement = service.StringPointerOrNull(n.Placement)
	m.ResponseCondition = service.StringPointerOrDefault(n.ResponseCondition, DefaultResponseCondition)

	return m
}

// ResetVCLOnlyToDefaults restores the VCL-only fields to their schema defaults
// after a flatten. On a Compute service they are never sent, so the API's own
// values are discarded rather than reported as a diff against the plan.
func ResetVCLOnlyToDefaults(m *NestedModel) {
	m.Format = types.StringValue(constants.LoggingNewRelicDefaultFormat)
	m.FormatVersion = types.Int64Value(DefaultFormatVersion)
	m.Placement = types.StringNull()
	m.ResponseCondition = types.StringValue(DefaultResponseCondition)
}

// FlattenToComputeNestedModel is FlattenToNestedModel for Compute services: it
// carries over only the attributes ComputeNestedModel exposes.
func FlattenToComputeNestedModel(n *fastly.NewRelic) ComputeNestedModel {
	return ComputeNestedModel{commonModel: FlattenToNestedModel(n).commonModel}
}

func flatten(ctx context.Context, n *fastly.NewRelic, m *Model) {
	if n == nil {
		tflog.Warn(ctx, "flatten called with nil New Relic logging endpoint")
		return
	}

	id := fastly.ToValue(n.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(n.ServiceVersion)) + "-" + fastly.ToValue(n.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(n.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(n.ServiceVersion)))

	m.NestedModel = FlattenToNestedModel(n)

	tflog.Debug(ctx, "Flattened New Relic logging endpoint state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
