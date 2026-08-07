package loggingbigquery

import (
	fastly "github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateBigQueryInput {
	input := &fastly.CreateBigQueryInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		ProjectID:      new(service.StringValue(m.ProjectID)),
		Dataset:        new(service.StringValue(m.Dataset)),
		Table:          new(service.StringValue(m.Table)),
	}

	input.AccountName = fastly.NullString(service.StringValue(m.AccountName()))
	input.User = fastly.NullString(service.StringValue(m.Email()))
	input.SecretKey = fastly.NullString(service.StringValue(m.SecretKey()))
	input.Template = fastly.NullString(service.StringValue(m.Template))
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
func BuildComputeCreateInput(serviceID string, version int, m ComputeNestedModel) *fastly.CreateBigQueryInput {
	input := &fastly.CreateBigQueryInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		ProjectID:      new(service.StringValue(m.ProjectID)),
		Dataset:        new(service.StringValue(m.Dataset)),
		Table:          new(service.StringValue(m.Table)),
	}

	input.AccountName = fastly.NullString(service.StringValue(m.AccountName()))
	input.User = fastly.NullString(service.StringValue(m.Email()))
	input.SecretKey = fastly.NullString(service.StringValue(m.SecretKey()))
	input.Template = fastly.NullString(service.StringValue(m.Template))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	return input
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateBigQueryInput {
	input := &fastly.UpdateBigQueryInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		ProjectID:      new(service.StringValue(m.ProjectID)),
		Dataset:        new(service.StringValue(m.Dataset)),
		Table:          new(service.StringValue(m.Table)),
	}

	// Unlike email/secret_key below, account_name can't be cleared via update:
	// the API rejects an explicit empty string ("Invalid account_name ''"), so
	// it must be omitted (fastly.NullString) rather than sent as "" (new()).
	input.AccountName = fastly.NullString(service.StringValue(m.AccountName()))
	input.User = new(service.StringValue(m.Email()))
	input.SecretKey = new(service.StringValue(m.SecretKey()))
	input.Template = new(service.StringValue(m.Template))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	// placement can be cleared back to unset / nil (distinct from "none" — see
	// schema.go). UpdateBigQueryInput.Placement is a *Nullable[string]
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

// BuildComputeUpdateInput is BuildUpdateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeUpdateInput(serviceID string, version int, m ComputeNestedModel) *fastly.UpdateBigQueryInput {
	input := &fastly.UpdateBigQueryInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		ProjectID:      new(service.StringValue(m.ProjectID)),
		Dataset:        new(service.StringValue(m.Dataset)),
		Table:          new(service.StringValue(m.Table)),
	}

	// See BuildUpdateInput for why AccountName uses fastly.NullString here while
	// User/SecretKey use new().
	input.AccountName = fastly.NullString(service.StringValue(m.AccountName()))
	input.User = new(service.StringValue(m.Email()))
	input.SecretKey = new(service.StringValue(m.SecretKey()))
	input.Template = new(service.StringValue(m.Template))
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	return input
}

// ClearVCLOnlyCreateFields nils out format, format_version, placement, and
// response_condition on a CreateBigQueryInput. The standalone
// fastly_service_logging_bigquery resource shares one schema across both
// service types, so this is called instead of BuildComputeCreateInput to strip
// the VCL-only fields once the service is confirmed to be Compute.
func ClearVCLOnlyCreateFields(input *fastly.CreateBigQueryInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// ClearVCLOnlyUpdateFields is ClearVCLOnlyCreateFields for UpdateBigQueryInput.
func ClearVCLOnlyUpdateFields(input *fastly.UpdateBigQueryInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}
