package logginggcs

import (
	"context"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

// buildCommonCreateInput sets the Create fields shared by VCL and Compute
// services. BuildCreateInput and BuildComputeCreateInput layer their
// service-type-specific fields on top of this.
func buildCommonCreateInput(serviceID string, version int, m commonModel) *fastly.CreateGCSInput {
	input := &fastly.CreateGCSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		Bucket:         new(service.StringValue(m.BucketName)),
	}

	// account_name and email/secret_key are alternative auth methods, but not
	// mutually exclusive on the API: it accepts both being set at once (each
	// is sent whenever configured, with no client-side check against the
	// other), and does not reject the request.
	input.AccountName = fastly.NullString(service.StringValue(m.AccountName()))
	input.User = fastly.NullString(service.StringValue(m.Email()))
	input.SecretKey = fastly.NullString(service.StringValue(m.SecretKey()))
	input.ProjectID = fastly.NullString(service.StringValue(m.ProjectID))
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
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	return input
}

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateGCSInput {
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
func BuildComputeCreateInput(serviceID string, version int, m ComputeNestedModel) *fastly.CreateGCSInput {
	return buildCommonCreateInput(serviceID, version, m.commonModel)
}

// ClearVCLOnlyCreateFields nils out format, format_version, placement, and
// response_condition on a CreateGCSInput. The standalone
// fastly_service_logging_gcs resource shares one schema across both service
// types, so this is called instead of BuildComputeCreateInput to strip the
// VCL-only fields once the service is confirmed to be Compute.
func ClearVCLOnlyCreateFields(input *fastly.CreateGCSInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// ClearVCLOnlyUpdateFields is ClearVCLOnlyCreateFields for UpdateGCSInput.
func ClearVCLOnlyUpdateFields(input *fastly.UpdateGCSInput) {
	input.Format = nil
	input.FormatVersion = nil
	input.Placement = nil
	input.ResponseCondition = nil
}

// buildCommonUpdateInput sets the Update fields shared by VCL and Compute
// services. BuildUpdateInput and BuildComputeUpdateInput layer their
// service-type-specific fields on top of this.
func buildCommonUpdateInput(serviceID string, version int, m commonModel) *fastly.UpdateGCSInput {
	input := &fastly.UpdateGCSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		NewName:        new(service.StringValue(m.Name)),
		Bucket:         new(service.StringValue(m.BucketName)),
	}

	// Unlike email/secret_key below, account_name can't be cleared via update:
	// the API rejects an explicit empty account_name on update, so it must be
	// omitted (fastly.NullString) rather than sent as "" (new()). See
	// UpdateOrRecreate/needsRecreateForAccountNameClear.
	input.AccountName = fastly.NullString(service.StringValue(m.AccountName()))
	input.User = new(service.StringValue(m.Email()))
	input.SecretKey = new(service.StringValue(m.SecretKey()))
	input.ProjectID = new(service.StringValue(m.ProjectID))
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
	input.ProcessingRegion = fastly.NullString(service.StringValue(m.ProcessingRegion))

	return input
}

// BuildComputeUpdateInput is BuildUpdateInput for Compute services: it never
// sets format, format_version, placement, or response_condition, since those
// only affect generated VCL and Compute services don't have any.
func BuildComputeUpdateInput(serviceID string, version int, m ComputeNestedModel) *fastly.UpdateGCSInput {
	return buildCommonUpdateInput(serviceID, version, m.commonModel)
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateGCSInput {
	input := buildCommonUpdateInput(serviceID, version, m.commonModel)
	input.Format = fastly.NullString(service.StringValue(m.Format))
	input.FormatVersion = fastly.NullInt(int(service.Int64Value(m.FormatVersion)))
	// placement can be cleared back to unset / nil (distinct from "none" — see
	// schema.go). UpdateGCSInput.Placement is a *Nullable[string] specifically
	// so this can be sent as a real JSON null: omitting the field leaves the
	// previous value in place, and sending a literal empty string gets stored
	// as "" rather than reverting to null/auto-placement — neither actually
	// clears it.
	if v := service.StringValue(m.Placement); v != "" {
		input.Placement = fastly.NewNullable(v)
	} else {
		input.Placement = fastly.NullValue[string]()
	}
	input.ResponseCondition = new(service.StringValue(m.ResponseCondition))
	return input
}

// UpdateOrRecreate applies opts via UpdateGCS, unless recreate is true, in
// which case it deletes and recreates the endpoint via createInput instead.
// recreate must be true exactly when the desired account_name is empty but
// the endpoint currently being changed has a non-empty one: the API rejects
// an explicit empty account_name on update (see buildCommonUpdateInput), so
// omitting it there just leaves the old account_name in place, silently
// diverging from a plan that shows it cleared in favor of email/secret_key.
func UpdateOrRecreate(ctx context.Context, client *fastly.Client, recreate bool, opts *fastly.UpdateGCSInput, createInput *fastly.CreateGCSInput) (*fastly.GCS, error) {
	if !recreate {
		return client.UpdateGCS(ctx, opts)
	}

	if err := client.DeleteGCS(ctx, &fastly.DeleteGCSInput{
		ServiceID:      opts.ServiceID,
		ServiceVersion: opts.ServiceVersion,
		Name:           opts.Name,
	}); err != nil {
		return nil, err
	}
	return client.CreateGCS(ctx, createInput)
}

// needsRecreateForAccountNameClear reports whether clearing account_name to
// desiredAccountName requires UpdateOrRecreate to delete+recreate this
// endpoint rather than update it in place. The reconcile path (unlike the
// standalone resource's Update, which already has the prior state on hand)
// only has the desired model, so the endpoint's current account_name is
// looked up directly.
func needsRecreateForAccountNameClear(ctx context.Context, client *fastly.Client, serviceID string, version int, name, desiredAccountName types.String) (bool, error) {
	if service.StringValue(desiredAccountName) != "" {
		return false, nil
	}

	remote, err := client.GetGCS(ctx, &fastly.GetGCSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(name),
	})
	if err != nil {
		return false, err
	}

	return fastly.ToValue(remote.AccountName) != "", nil
}
