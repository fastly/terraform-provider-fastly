package loggingnewrelicotlp

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(n *fastly.NewRelicOTLP) NestedModel {
	m := NestedModel{}

	if n == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(n.Name))
	m.Token = types.StringValue(fastly.ToValue(n.Token))
	m.Region = service.StringPointerOrDefault(n.Region, DefaultRegion)
	m.URL = service.StringPointerOrDefault(n.URL, DefaultURL)
	m.ProcessingRegion = service.StringPointerOrDefault(n.ProcessingRegion, DefaultProcessingRegion)
	m.Format = service.StringPointerOrDefault(n.Format, constants.LoggingNewRelicOTLPDefaultFormat)
	m.FormatVersion = service.Int64PointerOrDefault(n.FormatVersion, DefaultFormatVersion)
	m.Placement = service.StringPointerOrDefault(n.Placement, DefaultPlacement)
	m.ResponseCondition = service.StringPointerOrDefault(n.ResponseCondition, DefaultResponseCondition)

	return m
}

// FlattenToComputeNestedModel is FlattenToNestedModel for Compute services: it
// carries over only the attributes ComputeNestedModel exposes.
func FlattenToComputeNestedModel(n *fastly.NewRelicOTLP) ComputeNestedModel {
	return ComputeNestedModel{commonModel: FlattenToNestedModel(n).commonModel}
}

func flatten(ctx context.Context, n *fastly.NewRelicOTLP, m *Model) {
	if n == nil {
		tflog.Warn(ctx, "flatten called with nil New Relic OTLP logging endpoint")
		return
	}

	id := fastly.ToValue(n.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(n.ServiceVersion)) + "-" + fastly.ToValue(n.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(n.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(n.ServiceVersion)))

	m.NestedModel = FlattenToNestedModel(n)

	tflog.Debug(ctx, "Flattened New Relic OTLP logging endpoint state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
