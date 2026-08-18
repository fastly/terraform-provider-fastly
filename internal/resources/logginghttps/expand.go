package logginghttps

import (
	fastly "github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

// buildCommonCreateInput sets the Create fields shared by VCL and Compute
// services. BuildCreateInput and BuildComputeCreateInput layer their
// service-type-specific fields on top of this.
func buildCommonCreateInput(serviceID string, version int, m commonModel) *fastly.CreateHTTPSInput {
	input := &fastly.CreateHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		URL:            new(service.StringValue(m.URL)),
	}

	input.ContentType = fastly.NullString(service.StringValue(m.ContentType))
	input.CompressionCodec = fastly.NullString(service.StringValue(m.CompressionCodec))
	// Only send an explicitly configured gzip_level. DefaultGzipLevel (-1) means
	// unset: the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level when omitted. fastly.NullInt is
	// not used here because it treats 0 as unset too, which would silently drop
	// an explicit "no compression" (gzip_level = 0).
	if gl := service.Int64Value(m.GzipLevel); gl != DefaultGzipLevel {
		input.GzipLevel = new(int(gl))
	}
	input.HeaderName = fastly.NullString(service.StringValue(m.HeaderName))
	input.HeaderValue = fastly.NullString(service.StringValue(m.HeaderValue))
	input.JSONFormat = fastly.NullString(service.StringValue(m.JSONFormat))
	input.MessageType = fastly.NullString(service.StringValue(m.MessageType))
	input.Method = fastly.NullString(service.StringValue(m.Method))
	input.Period = fastly.NullInt(int(service.Int64Value(m.Period)))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))
	input.RequestMaxBytes = fastly.NullInt(int(service.Int64Value(m.RequestMaxBytes)))
	input.RequestMaxEntries = fastly.NullInt(int(service.Int64Value(m.RequestMaxEntries)))
	input.TLSCACert = fastly.NullString(service.StringValue(m.TLSCACert()))
	input.TLSClientCert = fastly.NullString(service.StringValue(m.TLSClientCert()))
	input.TLSClientKey = fastly.NullString(service.StringValue(m.TLSClientKey()))
	input.TLSHostname = fastly.NullString(service.StringValue(m.TLSHostname()))

	return input
}

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateHTTPSInput {
	input := buildCommonCreateInput(serviceID, version, m.commonModel)
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	input.Placement = fastly.NullString(service.StringValue(m.Placement))
	input.ResponseCondition = fastly.NullString(service.StringValue(m.ResponseCondition))
	return input
}

// BuildComputeCreateInput is BuildCreateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeCreateInput(serviceID string, version int, m ComputeNestedModel) *fastly.CreateHTTPSInput {
	return buildCommonCreateInput(serviceID, version, m.commonModel)
}

// ClearVCLOnlyCreateFields nils out format, format_version, placement, and
// response_condition on a CreateHTTPSInput. The standalone
// fastly_service_logging_https resource shares one schema across both service
// types, so this is called instead of BuildComputeCreateInput to strip the
// VCL-only fields once the service is confirmed to be Compute.
func ClearVCLOnlyCreateFields(input *fastly.CreateHTTPSInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// ClearVCLOnlyUpdateFields is ClearVCLOnlyCreateFields for UpdateHTTPSInput.
func ClearVCLOnlyUpdateFields(input *fastly.UpdateHTTPSInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// buildCommonUpdateInput sets the Update fields shared by VCL and Compute
// services. BuildUpdateInput and BuildComputeUpdateInput layer their
// service-type-specific fields on top of this.
func buildCommonUpdateInput(serviceID string, version int, m commonModel) *fastly.UpdateHTTPSInput {
	input := &fastly.UpdateHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		URL:            new(service.StringValue(m.URL)),
	}

	// content_type, header_name, header_value, and compression_codec default to
	// "" and can be cleared. Always send a concrete value via new() rather than
	// fastly.NullString, which maps "" to nil, omits the field, and leaves the
	// previously-set value in place.
	input.ContentType = new(service.StringValue(m.ContentType))
	input.CompressionCodec = new(service.StringValue(m.CompressionCodec))
	// Only send an explicitly configured gzip_level. DefaultGzipLevel (-1) means
	// unset: the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level when omitted.
	if gl := service.Int64Value(m.GzipLevel); gl != DefaultGzipLevel {
		input.GzipLevel = new(int(gl))
	}
	input.HeaderName = new(service.StringValue(m.HeaderName))
	input.HeaderValue = new(service.StringValue(m.HeaderValue))
	input.JSONFormat = fastly.NullString(service.StringValue(m.JSONFormat))
	input.MessageType = fastly.NullString(service.StringValue(m.MessageType))
	input.Method = fastly.NullString(service.StringValue(m.Method))
	input.Period = fastly.NullInt(int(service.Int64Value(m.Period)))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))
	// request_max_bytes/request_max_entries default to 0 ("unbounded"), a
	// legitimate explicit value a practitioner may set back to after raising it,
	// so — like content_type/header_name/header_value/compression_codec above —
	// they must always be sent rather than omitted when zero (fastly.NullInt
	// maps 0 to nil).
	input.RequestMaxBytes = new(int(service.Int64Value(m.RequestMaxBytes)))
	input.RequestMaxEntries = new(int(service.Int64Value(m.RequestMaxEntries)))
	// tls.* default to "" and are clearable, same reasoning as above.
	input.TLSCACert = new(service.StringValue(m.TLSCACert()))
	input.TLSClientCert = new(service.StringValue(m.TLSClientCert()))
	input.TLSClientKey = new(service.StringValue(m.TLSClientKey()))
	input.TLSHostname = new(service.StringValue(m.TLSHostname()))

	return input
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateHTTPSInput {
	input := buildCommonUpdateInput(serviceID, version, m.commonModel)
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	// placement can be cleared back to unset / nil (distinct from "none" — see
	// schema.go). UpdateHTTPSInput.Placement is a *Nullable[string] specifically
	// so this can be sent as a real JSON null: omitting the field leaves the
	// previous value in place, and sending a literal empty string gets stored as
	// "" rather than reverting to null/auto-placement — neither actually clears
	// it.
	if v := service.StringValue(m.Placement); v != "" {
		input.Placement = fastly.NewNullable(v)
	} else {
		input.Placement = fastly.NullValue[string]()
	}
	input.ResponseCondition = new(service.StringValue(m.ResponseCondition))
	return input
}

// BuildComputeUpdateInput is BuildUpdateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeUpdateInput(serviceID string, version int, m ComputeNestedModel) *fastly.UpdateHTTPSInput {
	return buildCommonUpdateInput(serviceID, version, m.commonModel)
}
