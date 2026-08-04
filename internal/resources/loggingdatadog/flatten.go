package loggingdatadog

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(d *fastly.Datadog) NestedModel {
	m := NestedModel{}

	if d == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(d.Name))
	m.Authentication = NewAuthenticationObject(
		service.StringPointerOrDefault(d.Token, ""),
	)
	m.Region = service.StringPointerOrDefault(d.Region, DefaultRegion)
	m.ProcessingRegion = service.StringPointerOrDefault(d.ProcessingRegion, DefaultProcessingRegion)
	m.Format = service.StringPointerOrDefault(d.Format, constants.LoggingDatadogDefaultFormat)
	m.FormatVersion = service.Int64PointerOrDefault(d.FormatVersion, DefaultFormatVersion)
	m.Placement = service.StringPointerOrNull(d.Placement)
	m.ResponseCondition = service.StringPointerOrDefault(d.ResponseCondition, DefaultResponseCondition)

	return m
}

// FlattenToComputeNestedModel is FlattenToNestedModel for Compute services: it
// carries over only the attributes ComputeNestedModel exposes.
func FlattenToComputeNestedModel(d *fastly.Datadog) ComputeNestedModel {
	return ComputeNestedModel{commonModel: FlattenToNestedModel(d).commonModel}
}

func flatten(ctx context.Context, d *fastly.Datadog, m *Model) {
	if d == nil {
		tflog.Warn(ctx, "flatten called with nil Datadog logging endpoint")
		return
	}

	id := fastly.ToValue(d.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(d.ServiceVersion)) + "-" + fastly.ToValue(d.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(d.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(d.ServiceVersion)))

	m.NestedModel = FlattenToNestedModel(d)

	tflog.Debug(ctx, "Flattened Datadog logging endpoint state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
