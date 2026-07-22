package loggingnewrelicotlp

import (
	fastly "github.com/fastly/go-fastly/v16/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateNewRelicOTLPInput {
	input := &fastly.CreateNewRelicOTLPInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		Token:          new(service.StringValue(m.Token)),
	}

	input.Region = fastly.NullString(service.StringValue(m.Region))
	input.URL = fastly.NullString(service.StringValue(m.URL))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	input.Placement = fastly.NullString(service.StringValue(m.Placement))
	input.ResponseCondition = fastly.NullString(service.StringValue(m.ResponseCondition))

	return input
}

// BuildComputeCreateInput is BuildCreateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeCreateInput(serviceID string, version int, m ComputeNestedModel) *fastly.CreateNewRelicOTLPInput {
	input := &fastly.CreateNewRelicOTLPInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		Token:          new(service.StringValue(m.Token)),
	}

	input.Region = fastly.NullString(service.StringValue(m.Region))
	input.URL = fastly.NullString(service.StringValue(m.URL))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	return input
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateNewRelicOTLPInput {
	input := &fastly.UpdateNewRelicOTLPInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		Token:          new(service.StringValue(m.Token)),
	}

	input.Region = fastly.NullString(service.StringValue(m.Region))
	// url and response_condition default to "" and can be cleared, so always send
	// them as a concrete value on update. fastly.NullString is not used because it
	// maps "" to nil, which omits the field (url,omitempty) and leaves a
	// previously-set value in place — producing an inconsistent-result error when
	// the user clears the attribute.
	input.URL = new(service.StringValue(m.URL))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	input.Placement = fastly.NullString(service.StringValue(m.Placement))
	input.ResponseCondition = new(service.StringValue(m.ResponseCondition))

	return input
}

// BuildComputeUpdateInput is BuildUpdateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeUpdateInput(serviceID string, version int, m ComputeNestedModel) *fastly.UpdateNewRelicOTLPInput {
	input := &fastly.UpdateNewRelicOTLPInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		Token:          new(service.StringValue(m.Token)),
	}

	input.Region = fastly.NullString(service.StringValue(m.Region))
	// url defaults to "" and can be cleared, so always send it as a concrete
	// value on update — see BuildUpdateInput.
	input.URL = new(service.StringValue(m.URL))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	return input
}

// ClearVCLOnlyCreateFields nils out format, format_version, placement, and
// response_condition on a CreateNewRelicOTLPInput. The standalone
// fastly_service_logging_newrelicotlp resource shares one schema across both
// service types, so this is called instead of BuildComputeCreateInput to strip
// the VCL-only fields once the service is confirmed to be Compute.
func ClearVCLOnlyCreateFields(input *fastly.CreateNewRelicOTLPInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// ClearVCLOnlyUpdateFields is ClearVCLOnlyCreateFields for
// UpdateNewRelicOTLPInput.
func ClearVCLOnlyUpdateFields(input *fastly.UpdateNewRelicOTLPInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}
