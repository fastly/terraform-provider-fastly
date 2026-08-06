package loggingbigquery

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(bq *fastly.BigQuery) NestedModel {
	m := NestedModel{}

	if bq == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(bq.Name))
	m.ProjectID = types.StringValue(fastly.ToValue(bq.ProjectID))
	m.Dataset = types.StringValue(fastly.ToValue(bq.Dataset))
	m.Table = types.StringValue(fastly.ToValue(bq.Table))
	m.Authentication = NewAuthenticationObject(
		service.StringPointerOrDefault(bq.AccountName, ""),
		service.StringPointerOrDefault(bq.User, ""),
		service.StringPointerOrDefault(bq.SecretKey, ""),
	)
	m.Template = service.StringPointerOrDefault(bq.Template, DefaultTemplate)
	m.ProcessingRegion = service.StringPointerOrDefault(bq.ProcessingRegion, DefaultProcessingRegion)
	m.Format = service.StringPointerOrDefault(bq.Format, constants.LoggingBigQueryDefaultFormat)
	m.FormatVersion = service.Int64PointerOrDefault(bq.FormatVersion, DefaultFormatVersion)
	m.Placement = service.StringPointerOrNull(bq.Placement)
	m.ResponseCondition = service.StringPointerOrDefault(bq.ResponseCondition, DefaultResponseCondition)

	return m
}

// ResetVCLOnlyToDefaults restores the VCL-only fields to their schema defaults
// after a flatten. On a Compute service they are never sent, so the API's own
// values are discarded rather than reported as a diff against the plan.
func ResetVCLOnlyToDefaults(m *NestedModel) {
	m.Format = types.StringValue(constants.LoggingBigQueryDefaultFormat)
	m.FormatVersion = types.Int64Value(DefaultFormatVersion)
	m.Placement = types.StringNull()
	m.ResponseCondition = types.StringValue(DefaultResponseCondition)
}

// FlattenToComputeNestedModel is FlattenToNestedModel for Compute services: it
// carries over only the attributes ComputeNestedModel exposes.
func FlattenToComputeNestedModel(bq *fastly.BigQuery) ComputeNestedModel {
	return ComputeNestedModel{commonModel: FlattenToNestedModel(bq).commonModel}
}

func flatten(ctx context.Context, bq *fastly.BigQuery, m *Model) {
	if bq == nil {
		tflog.Warn(ctx, "flatten called with nil BigQuery logging endpoint")
		return
	}

	id := fastly.ToValue(bq.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(bq.ServiceVersion)) + "-" + fastly.ToValue(bq.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(bq.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(bq.ServiceVersion)))

	m.NestedModel = FlattenToNestedModel(bq)

	tflog.Debug(ctx, "Flattened BigQuery logging endpoint state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
