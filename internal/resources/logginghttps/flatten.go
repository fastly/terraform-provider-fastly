package logginghttps

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(h *fastly.HTTPS) NestedModel {
	m := NestedModel{}

	if h == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(h.Name))
	m.URL = types.StringValue(fastly.ToValue(h.URL))
	m.ContentType = service.StringPointerOrDefault(h.ContentType, DefaultContentType)
	m.CompressionCodec = service.StringPointerOrDefault(h.CompressionCodec, DefaultCompressionCodec)
	m.GzipLevel = service.Int64PointerOrDefault(h.GzipLevel, DefaultGzipLevel)
	m.HeaderName = service.StringPointerOrDefault(h.HeaderName, DefaultHeaderName)
	m.HeaderValue = service.StringPointerOrDefault(h.HeaderValue, DefaultHeaderValue)
	m.JSONFormat = service.StringPointerOrDefault(h.JSONFormat, DefaultJSONFormat)
	m.MessageType = service.StringPointerOrDefault(h.MessageType, DefaultMessageType)
	m.Method = service.StringPointerOrDefault(h.Method, DefaultMethod)
	m.Period = service.Int64PointerOrDefault(h.Period, DefaultPeriod)
	m.ProcessingRegion = service.StringPointerOrDefault(h.ProcessingRegion, DefaultProcessingRegion)
	m.RequestMaxBytes = service.Int64PointerOrDefault(h.RequestMaxBytes, DefaultRequestMaxBytes)
	m.RequestMaxEntries = service.Int64PointerOrDefault(h.RequestMaxEntries, DefaultRequestMaxEntries)
	m.TLS = NewTLSObject(
		service.StringPointerOrDefault(h.TLSCACert, ""),
		service.StringPointerOrDefault(h.TLSClientCert, ""),
		service.StringPointerOrDefault(h.TLSClientKey, ""),
		service.StringPointerOrDefault(h.TLSHostname, DefaultTLSHostname),
	)
	m.Format = service.StringPointerOrDefault(h.Format, constants.LoggingHTTPSDefaultFormat)
	m.FormatVersion = service.Int64PointerOrDefault(h.FormatVersion, DefaultFormatVersion)
	m.Placement = service.StringPointerOrNull(h.Placement)
	m.ResponseCondition = service.StringPointerOrDefault(h.ResponseCondition, DefaultResponseCondition)

	return m
}

// ResetVCLOnlyToDefaults restores the VCL-only fields to their schema defaults
// after a flatten. On a Compute service they are never sent, so the API's own
// values are discarded rather than reported as a diff against the plan.
func ResetVCLOnlyToDefaults(m *NestedModel) {
	m.Format = types.StringValue(constants.LoggingHTTPSDefaultFormat)
	m.FormatVersion = types.Int64Value(DefaultFormatVersion)
	m.Placement = types.StringNull()
	m.ResponseCondition = types.StringValue(DefaultResponseCondition)
}

// FlattenToComputeNestedModel is FlattenToNestedModel for Compute services: it
// carries over only the attributes ComputeNestedModel exposes.
func FlattenToComputeNestedModel(h *fastly.HTTPS) ComputeNestedModel {
	return ComputeNestedModel{commonModel: FlattenToNestedModel(h).commonModel}
}

// preserveGzipSentinelCommon restores the gzip_level "unset" sentinel after a
// flatten. When gzip_level was not configured (desired == DefaultGzipLevel),
// the API's auto-managed value is discarded so the provider does not report a
// permanent diff against the sentinel.
func preserveGzipSentinelCommon(m *commonModel, desired commonModel) {
	if service.Int64Value(desired.GzipLevel) == DefaultGzipLevel {
		m.GzipLevel = types.Int64Value(DefaultGzipLevel)
	}
}

func preserveGzipSentinel(m *NestedModel, desired NestedModel) {
	preserveGzipSentinelCommon(&m.commonModel, desired.commonModel)
}

func preserveGzipSentinelCompute(m *ComputeNestedModel, desired ComputeNestedModel) {
	preserveGzipSentinelCommon(&m.commonModel, desired.commonModel)
}

// inferGzipSentinelOnImport approximates the unset sentinel when there is no
// desired/prior model to compare against — a freshly imported resource, or an
// endpoint discovered on read that isn't tracked in config/state yet. The API
// always returns a concrete gzip_level even when it was never configured, and
// with no compression_codec set that auto-managed value is 0, indistinguishable
// from an explicit gzip_level = 0. Treating that case as unset is the better
// default: a genuine explicit 0 self-corrects with one harmless update on the
// next apply, whereas leaving 0 in state would permanently diverge from the
// -1 a Terraform-managed create/read always produces for "never configured".
func inferGzipSentinelOnImport(m *commonModel) {
	if service.StringValue(m.CompressionCodec) == "" && service.Int64Value(m.GzipLevel) == 0 {
		m.GzipLevel = types.Int64Value(DefaultGzipLevel)
	}
}

// preserveGzipSentinelList applies preserveGzipSentinel to each read model using
// the matching desired/prior model (by name), falling back to
// inferGzipSentinelOnImport for models with no match (e.g. freshly imported).
func preserveGzipSentinelList(read, desired []NestedModel) {
	desiredByName := make(map[string]NestedModel, len(desired))
	for _, d := range desired {
		desiredByName[service.StringValue(d.Name)] = d
	}
	for i := range read {
		if d, ok := desiredByName[service.StringValue(read[i].Name)]; ok {
			preserveGzipSentinel(&read[i], d)
		} else {
			inferGzipSentinelOnImport(&read[i].commonModel)
		}
	}
}

// preserveGzipSentinelListCompute is preserveGzipSentinelList for Compute's
// ComputeNestedModel.
func preserveGzipSentinelListCompute(read, desired []ComputeNestedModel) {
	desiredByName := make(map[string]ComputeNestedModel, len(desired))
	for _, d := range desired {
		desiredByName[service.StringValue(d.Name)] = d
	}
	for i := range read {
		if d, ok := desiredByName[service.StringValue(read[i].Name)]; ok {
			preserveGzipSentinelCompute(&read[i], d)
		} else {
			inferGzipSentinelOnImport(&read[i].commonModel)
		}
	}
}

func flatten(ctx context.Context, h *fastly.HTTPS, m *Model) {
	if h == nil {
		tflog.Warn(ctx, "flatten called with nil HTTPS logging endpoint")
		return
	}

	id := fastly.ToValue(h.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(h.ServiceVersion)) + "-" + fastly.ToValue(h.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(h.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(h.ServiceVersion)))

	m.NestedModel = FlattenToNestedModel(h)

	tflog.Debug(ctx, "Flattened HTTPS logging endpoint state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
