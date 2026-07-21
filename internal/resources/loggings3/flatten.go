package loggings3

import (
	"context"
	"strconv"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func FlattenToNestedModel(s *fastly.S3) NestedModel {
	m := NestedModel{}

	if s == nil {
		return m
	}

	m.Name = types.StringValue(fastly.ToValue(s.Name))
	m.BucketName = types.StringValue(fastly.ToValue(s.BucketName))
	m.Authentication = NewAuthenticationObject(
		service.StringPointerOrDefault(s.AccessKey, ""),
		service.StringPointerOrDefault(s.SecretKey, ""),
		service.StringPointerOrDefault(s.IAMRole, ""),
	)
	m.Domain = service.StringPointerOrDefault(s.Domain, DefaultDomain)
	m.Path = service.StringPointerOrDefault(s.Path, DefaultPath)
	m.Period = service.Int64PointerOrDefault(s.Period, DefaultPeriod)
	m.GzipLevel = service.Int64PointerOrDefault(s.GzipLevel, DefaultGzipLevel)
	m.CompressionCodec = service.StringPointerOrDefault(s.CompressionCodec, DefaultCompressionCodec)
	m.Format = service.StringPointerOrDefault(s.Format, constants.LoggingS3DefaultFormat)
	m.FormatVersion = service.Int64PointerOrDefault(s.FormatVersion, DefaultFormatVersion)
	m.MessageType = service.StringPointerOrDefault(s.MessageType, DefaultMessageType)
	m.TimestampFormat = service.StringPointerOrDefault(s.TimestampFormat, DefaultTimestampFormat)
	m.Placement = service.StringPointerOrDefault(s.Placement, DefaultPlacement)
	m.ResponseCondition = service.StringPointerOrDefault(s.ResponseCondition, DefaultResponseCondition)
	m.PublicKey = service.StringPointerOrDefault(s.PublicKey, DefaultPublicKey)
	m.ProcessingRegion = service.StringPointerOrDefault(s.ProcessingRegion, DefaultProcessingRegion)
	m.ServerSideEncryptionKMSKeyID = service.StringPointerOrDefault(s.ServerSideEncryptionKMSKeyID, DefaultServerSideEncryptionKMSKeyID)
	m.FileMaxBytes = service.Int64PointerOrDefault(s.FileMaxBytes, 0)

	if s.ACL != nil {
		m.ACL = types.StringValue(string(*s.ACL))
	} else {
		m.ACL = types.StringValue(DefaultACL)
	}
	if s.Redundancy != nil {
		m.Redundancy = types.StringValue(string(*s.Redundancy))
	} else {
		m.Redundancy = types.StringValue(DefaultRedundancy)
	}
	if s.ServerSideEncryption != nil {
		m.ServerSideEncryption = types.StringValue(string(*s.ServerSideEncryption))
	} else {
		m.ServerSideEncryption = types.StringValue(DefaultServerSideEncryption)
	}

	return m
}

// FlattenToComputeNestedModel is FlattenToNestedModel for Compute services: it
// carries over only the attributes ComputeNestedModel exposes.
func FlattenToComputeNestedModel(s *fastly.S3) ComputeNestedModel {
	return ComputeNestedModel{commonModel: FlattenToNestedModel(s).commonModel}
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

func flatten(ctx context.Context, s *fastly.S3, m *Model) {
	if s == nil {
		tflog.Warn(ctx, "flatten called with nil S3 logging endpoint")
		return
	}

	id := fastly.ToValue(s.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(s.ServiceVersion)) + "-" + fastly.ToValue(s.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(s.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(s.ServiceVersion)))

	m.NestedModel = FlattenToNestedModel(s)

	tflog.Debug(ctx, "Flattened S3 logging endpoint state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
