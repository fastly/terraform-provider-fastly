package loggingsplunk

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(s *fastly.Splunk) NestedModel {
	m := NestedModel{}

	if s == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(s.Name))
	m.URL = types.StringValue(fastly.ToValue(s.URL))
	m.Authentication = NewAuthenticationObject(
		service.StringPointerOrDefault(s.Token, ""),
	)
	m.TLS = NewTLSObject(
		service.StringPointerOrDefault(s.TLSCACert, ""),
		service.StringPointerOrDefault(s.TLSClientCert, ""),
		service.StringPointerOrDefault(s.TLSClientKey, ""),
		service.StringPointerOrDefault(s.TLSHostname, DefaultTLSHostname),
	)
	m.UseTLS = service.BoolPointerOrDefault(s.UseTLS, DefaultUseTLS)
	m.ProcessingRegion = service.StringPointerOrDefault(s.ProcessingRegion, DefaultProcessingRegion)
	m.RequestMaxBytes = service.Int64PointerOrDefault(s.RequestMaxBytes, DefaultRequestMaxBytes)
	m.RequestMaxEntries = service.Int64PointerOrDefault(s.RequestMaxEntries, DefaultRequestMaxEntries)
	m.Format = service.StringPointerOrDefault(s.Format, constants.LoggingSplunkDefaultFormat)
	m.FormatVersion = service.Int64PointerOrDefault(s.FormatVersion, DefaultFormatVersion)
	m.Placement = service.StringPointerOrNull(s.Placement)
	m.ResponseCondition = service.StringPointerOrDefault(s.ResponseCondition, DefaultResponseCondition)

	return m
}

// ResetVCLOnlyToDefaults restores the VCL-only fields to their schema defaults
// after a flatten. On a Compute service they are never sent, so the API's own
// values are discarded rather than reported as a diff against the plan.
func ResetVCLOnlyToDefaults(m *NestedModel) {
	m.Format = types.StringValue(constants.LoggingSplunkDefaultFormat)
	m.FormatVersion = types.Int64Value(DefaultFormatVersion)
	m.Placement = types.StringNull()
	m.ResponseCondition = types.StringValue(DefaultResponseCondition)
}

// FlattenToComputeNestedModel is FlattenToNestedModel for Compute services: it
// carries over only the attributes ComputeNestedModel exposes.
func FlattenToComputeNestedModel(s *fastly.Splunk) ComputeNestedModel {
	return ComputeNestedModel{commonModel: FlattenToNestedModel(s).commonModel}
}

func flatten(ctx context.Context, s *fastly.Splunk, m *Model) {
	if s == nil {
		tflog.Warn(ctx, "flatten called with nil Splunk logging endpoint")
		return
	}

	id := fastly.ToValue(s.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(s.ServiceVersion)) + "-" + fastly.ToValue(s.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(s.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(s.ServiceVersion)))

	m.NestedModel = FlattenToNestedModel(s)

	tflog.Debug(ctx, "Flattened Splunk logging endpoint state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
