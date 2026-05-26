package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceComputePackageModel struct {
	Content        types.String `tfsdk:"content"`
	Filename       types.String `tfsdk:"filename"`
	SourceCodeHash types.String `tfsdk:"source_code_hash"`
}

func computePackageCommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"content": schema.StringAttribute{
			Optional:    true,
			Description: "The contents of the Compute deployment package as a base64-encoded string. Conflicts with `filename`.",
		},
		"filename": schema.StringAttribute{
			Optional:    true,
			Description: "The path to the Compute deployment package on the local filesystem. Conflicts with `content`.",
		},
		"source_code_hash": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Hash of the package contents used to detect local package changes.",
		},
	}
}

func computePackageNestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Compute package attached to this service version. At most one package block is supported.",
		NestedObject: schema.NestedBlockObject{
			Attributes: computePackageCommonAttributes(),
		},
	}
}

func computePackagesEqual(a, b []serviceComputePackageModel) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 {
		return true
	}

	return packageStringValuesEqual(a[0].Content, b[0].Content) &&
		packageStringValuesEqual(a[0].Filename, b[0].Filename) &&
		packageStringValuesEqual(a[0].SourceCodeHash, b[0].SourceCodeHash)
}

func packageStringValuesEqual(a, b types.String) bool {
	if a.IsNull() && b.IsNull() {
		return true
	}
	if a.IsUnknown() && b.IsUnknown() {
		return true
	}
	if a.IsNull() != b.IsNull() || a.IsUnknown() != b.IsUnknown() {
		return false
	}
	if a.IsNull() || a.IsUnknown() {
		return true
	}
	return a.ValueString() == b.ValueString()
}

func validateComputePackageInput(packages []serviceComputePackageModel) error {
	if len(packages) == 0 {
		return nil
	}

	if len(packages) > 1 {
		return fmt.Errorf("only one package block is supported")
	}

	pkg := packages[0]

	contentSet := !pkg.Content.IsNull() && !pkg.Content.IsUnknown() && pkg.Content.ValueString() != ""
	filenameSet := !pkg.Filename.IsNull() && !pkg.Filename.IsUnknown() && pkg.Filename.ValueString() != ""

	switch {
	case contentSet && filenameSet:
		return fmt.Errorf("package content and filename cannot both be set")
	case !contentSet && !filenameSet:
		return fmt.Errorf("package requires exactly one of content or filename")
	default:
		return nil
	}
}

func updateComputePackage(ctx context.Context, client *fastly.Client, serviceID string, version int, packages []serviceComputePackageModel) error {
	if len(packages) == 0 {
		return nil
	}

	if err := validateComputePackageInput(packages); err != nil {
		return err
	}

	pkg := packages[0]
	input := &fastly.UpdatePackageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	}

	if !pkg.Content.IsNull() && !pkg.Content.IsUnknown() && pkg.Content.ValueString() != "" {
		decoded, err := base64.StdEncoding.DecodeString(pkg.Content.ValueString())
		if err != nil {
			return fmt.Errorf("error decoding base64 package content for service %s version %d: %w", serviceID, version, err)
		}
		input.PackageContent = decoded
	}

	if !pkg.Filename.IsNull() && !pkg.Filename.IsUnknown() && pkg.Filename.ValueString() != "" {
		input.PackagePath = fastly.ToPointer(pkg.Filename.ValueString())
	}

	_, err := client.UpdatePackage(ctx, input)
	if err != nil {
		return fmt.Errorf("error updating Compute package for service %s version %d: %w", serviceID, version, err)
	}

	return nil
}

func readComputePackageForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int, current []serviceComputePackageModel) ([]serviceComputePackageModel, error) {
	if len(current) == 0 {
		return current, nil
	}

	pkg, err := client.GetPackage(ctx, &fastly.GetPackageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		if isFastlyNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	result := current[0]
	if pkg != nil && pkg.Metadata != nil && pkg.Metadata.FilesHash != nil && *pkg.Metadata.FilesHash != "" {
		result.SourceCodeHash = types.StringValue(*pkg.Metadata.FilesHash)
	} else if result.SourceCodeHash.IsUnknown() {
		result.SourceCodeHash = types.StringNull()
	}

	return []serviceComputePackageModel{result}, nil
}
