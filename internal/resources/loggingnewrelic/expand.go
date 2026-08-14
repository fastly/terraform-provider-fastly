package loggingnewrelic

import (
	fastly "github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

// buildCommonCreateInput sets the Create fields shared by VCL and Compute
// services. BuildCreateInput and BuildComputeCreateInput layer their
// service-type-specific fields on top of this.
func buildCommonCreateInput(serviceID string, version int, m commonModel) *fastly.CreateNewRelicInput {
	input := &fastly.CreateNewRelicInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		Token:          new(service.StringValue(m.Token())),
	}

	input.Region = fastly.NullString(service.StringValue(m.Region))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	return input
}

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateNewRelicInput {
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
func BuildComputeCreateInput(serviceID string, version int, m ComputeNestedModel) *fastly.CreateNewRelicInput {
	return buildCommonCreateInput(serviceID, version, m.commonModel)
}

// buildCommonUpdateInput sets the Update fields shared by VCL and Compute
// services. BuildUpdateInput and BuildComputeUpdateInput layer their
// service-type-specific fields on top of this.
func buildCommonUpdateInput(serviceID string, version int, m commonModel) *fastly.UpdateNewRelicInput {
	input := &fastly.UpdateNewRelicInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		Token:          new(service.StringValue(m.Token())),
	}

	input.Region = fastly.NullString(service.StringValue(m.Region))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	return input
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateNewRelicInput {
	input := buildCommonUpdateInput(serviceID, version, m.commonModel)
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	// placement can be cleared back to unset / nil (distinct from "none" — see
	// schema.go). UpdateNewRelicInput.Placement is a *Nullable[string]
	// specifically so this can be sent as a real JSON null: omitting the field
	// leaves the previous value in place, and sending a literal empty string
	// gets stored as "" rather than reverting to null/auto-placement — neither
	// actually clears it.
	if v := service.StringValue(m.Placement); v != "" {
		input.Placement = fastly.NewNullable(v)
	} else {
		input.Placement = fastly.NullValue[string]()
	}
	// response_condition defaults to "" and can be cleared, so always send it as
	// a concrete value on update. fastly.NullString is not used because it maps
	// "" to nil, which omits the field (response_condition,omitempty) and leaves
	// a previously-set value in place — producing an inconsistent-result error
	// when the user clears the attribute.
	input.ResponseCondition = new(service.StringValue(m.ResponseCondition))
	return input
}

// BuildComputeUpdateInput is BuildUpdateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeUpdateInput(serviceID string, version int, m ComputeNestedModel) *fastly.UpdateNewRelicInput {
	return buildCommonUpdateInput(serviceID, version, m.commonModel)
}

// ClearVCLOnlyCreateFields nils out format, format_version, placement, and
// response_condition on a CreateNewRelicInput. The standalone
// fastly_service_logging_newrelic resource shares one schema across both
// service types, so this is called instead of BuildComputeCreateInput to
// strip the VCL-only fields once the service is confirmed to be Compute.
func ClearVCLOnlyCreateFields(input *fastly.CreateNewRelicInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// ClearVCLOnlyUpdateFields is ClearVCLOnlyCreateFields for UpdateNewRelicInput.
func ClearVCLOnlyUpdateFields(input *fastly.UpdateNewRelicInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}
