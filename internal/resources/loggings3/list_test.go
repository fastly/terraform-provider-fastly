package loggings3

import (
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/service"
)

// TestSetResourceAttrs_computeResetsVCLOnlyFields covers the same Compute
// normalization gap as TestResetVCLOnlyToDefaults (see loggings3_test.go),
// but for the list resource: the API still returns its own values for the
// VCL-only fields on a Compute-attached endpoint, and those must be reset to
// schema defaults here exactly as they are in Read/Import, or the list
// resource emits data the standalone resource's own schema rejects.
func TestSetResourceAttrs_computeResetsVCLOnlyFields(t *testing.T) {
	ctx := context.Background()

	s := &fastly.S3{
		ServiceID:         new("service-id"),
		ServiceVersion:    new(1),
		Name:              new("logger"),
		BucketName:        new("bucket"),
		Format:            new("api-assigned-compute-format"),
		FormatVersion:     new(2),
		Placement:         new("none"),
		ResponseCondition: new("some-condition"),
	}

	tests := []struct {
		name              string
		serviceType       string
		wantFormat        types.String
		wantFormatVersion types.Int64
		wantPlacement     types.String
		wantRespCondition types.String
	}{
		{
			name:              "vcl service keeps the API's values",
			serviceType:       service.TypeVCL,
			wantFormat:        types.StringValue("api-assigned-compute-format"),
			wantFormatVersion: types.Int64Value(2),
			wantPlacement:     types.StringValue("none"),
			wantRespCondition: types.StringValue("some-condition"),
		},
		{
			name:              "compute service resets VCL-only fields to schema defaults",
			serviceType:       service.TypeCompute,
			wantFormat:        types.StringValue(constants.LoggingS3DefaultFormat),
			wantFormatVersion: types.Int64Value(DefaultFormatVersion),
			wantPlacement:     types.StringNull(),
			wantRespCondition: types.StringValue(DefaultResponseCondition),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTestListResult(ctx)

			diags := setResourceAttrs(ctx, &result, s, "service-id", 1, tt.serviceType)
			require.False(t, diags.HasError(), diags)

			var got Model
			require.False(t, result.Resource.Get(ctx, &got).HasError())

			require.Equal(t, tt.wantFormat, got.Format)
			require.Equal(t, tt.wantFormatVersion, got.FormatVersion)
			require.Equal(t, tt.wantPlacement, got.Placement)
			require.Equal(t, tt.wantRespCondition, got.ResponseCondition)
		})
	}
}

// buildTestListResult builds a list.ListResult backed by the standalone
// resource's own schema, mirroring how List() constructs one via
// listidentity.NewResult but without requiring a full list.ListRequest.
func buildTestListResult(ctx context.Context) list.ListResult {
	s := schema.Schema{Attributes: ResourceAttributes()}
	return list.ListResult{
		Resource: &tfsdk.Resource{
			Raw:    tftypes.NewValue(s.Type().TerraformType(ctx), nil),
			Schema: s,
		},
	}
}
