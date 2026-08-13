package loggingblobstorage

import (
	fastly "github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

// buildCommonCreateInput sets the Create fields shared by VCL and Compute
// services. BuildCreateInput and BuildComputeCreateInput layer their
// service-type-specific fields on top of this.
func buildCommonCreateInput(serviceID string, version int, m commonModel) *fastly.CreateBlobStorageInput {
	input := &fastly.CreateBlobStorageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		Container:      new(service.StringValue(m.Container)),
	}

	input.AccountName = fastly.NullString(service.StringValue(m.AccountName()))
	input.SASToken = fastly.NullString(service.StringValue(m.SASToken()))
	input.Path = new(service.StringValue(m.Path))
	input.Period = fastly.NullInt(int(service.Int64Value(m.Period)))
	input.CompressionCodec = fastly.NullString(service.StringValue(m.CompressionCodec))
	// Only send an explicitly configured gzip_level. DefaultGzipLevel (-1) means
	// unset: the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level when omitted. fastly.NullInt is
	// not used here because it treats 0 as unset too, which would silently drop
	// an explicit "no compression" (gzip_level = 0).
	if gl := service.Int64Value(m.GzipLevel); gl != DefaultGzipLevel {
		input.GzipLevel = new(int(gl))
	}
	input.MessageType = fastly.NullString(service.StringValue(m.MessageType))
	input.TimestampFormat = fastly.NullString(service.StringValue(m.TimestampFormat))
	input.PublicKey = fastly.NullString(service.StringValue(m.PublicKey))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	if fmb := service.Int64Value(m.FileMaxBytes); fmb != 0 {
		v := int(fmb)
		input.FileMaxBytes = &v
	}

	return input
}

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateBlobStorageInput {
	input := buildCommonCreateInput(serviceID, version, m.commonModel)
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	input.Placement = fastly.NullString(service.StringValue(m.Placement))
	input.ResponseCondition = fastly.NullString(service.StringValue(m.ResponseCondition))
	return input
}

// ClearVCLOnlyCreateFields nils out format, format_version, placement, and
// response_condition on a CreateBlobStorageInput. The standalone
// fastly_service_logging_blobstorage resource shares one schema across both
// service types, so this is called instead of BuildComputeCreateInput to
// strip the VCL-only fields once the service is confirmed to be Compute.
func ClearVCLOnlyCreateFields(input *fastly.CreateBlobStorageInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// ClearVCLOnlyUpdateFields is ClearVCLOnlyCreateFields for
// UpdateBlobStorageInput.
func ClearVCLOnlyUpdateFields(input *fastly.UpdateBlobStorageInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// BuildComputeCreateInput is BuildCreateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeCreateInput(serviceID string, version int, m ComputeNestedModel) *fastly.CreateBlobStorageInput {
	return buildCommonCreateInput(serviceID, version, m.commonModel)
}

// buildCommonUpdateInput sets the Update fields shared by VCL and Compute
// services. BuildUpdateInput and BuildComputeUpdateInput layer their
// service-type-specific fields on top of this.
func buildCommonUpdateInput(serviceID string, version int, m commonModel) *fastly.UpdateBlobStorageInput {
	input := &fastly.UpdateBlobStorageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		Container:      new(service.StringValue(m.Container)),
	}

	// Credentials default to "" and can be cleared. Always send a concrete
	// value via new() rather than fastly.NullString, which maps "" to nil,
	// omits the field (account_name,omitempty / sas_token,omitempty), and
	// leaves the previously-set credential in place.
	input.AccountName = new(service.StringValue(m.AccountName()))
	input.SASToken = new(service.StringValue(m.SASToken()))
	input.Path = new(service.StringValue(m.Path))
	input.Period = fastly.NullInt(int(service.Int64Value(m.Period)))
	input.CompressionCodec = new(service.StringValue(m.CompressionCodec))
	// Only send an explicitly configured gzip_level. DefaultGzipLevel (-1) means
	// unset: the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level when omitted.
	if gl := service.Int64Value(m.GzipLevel); gl != DefaultGzipLevel {
		input.GzipLevel = new(int(gl))
	}
	input.MessageType = fastly.NullString(service.StringValue(m.MessageType))
	input.TimestampFormat = fastly.NullString(service.StringValue(m.TimestampFormat))
	input.PublicKey = new(service.StringValue(m.PublicKey))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	fmb := int(service.Int64Value(m.FileMaxBytes))
	input.FileMaxBytes = &fmb

	return input
}

// BuildComputeUpdateInput is BuildUpdateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeUpdateInput(serviceID string, version int, m ComputeNestedModel) *fastly.UpdateBlobStorageInput {
	return buildCommonUpdateInput(serviceID, version, m.commonModel)
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateBlobStorageInput {
	input := buildCommonUpdateInput(serviceID, version, m.commonModel)
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	// placement can be cleared back to unset / nil (distinct from "none" — see
	// schema.go). UpdateBlobStorageInput.Placement is a *Nullable[string]
	// specifically so this can be sent as a real JSON null: omitting the field
	// leaves the previous value in place, and sending a literal empty string
	// gets stored as "" rather than reverting to null/auto-placement — neither
	// actually clears it.
	if v := service.StringValue(m.Placement); v != "" {
		input.Placement = fastly.NewNullable(v)
	} else {
		input.Placement = fastly.NullValue[string]()
	}
	input.ResponseCondition = new(service.StringValue(m.ResponseCondition))
	return input
}
