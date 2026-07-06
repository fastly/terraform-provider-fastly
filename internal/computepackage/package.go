package computepackage

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/fastly/terraform-provider-fastly/internal/errors"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	Content        types.String `tfsdk:"content"`
	Filename       types.String `tfsdk:"filename"`
	SourceCodeHash types.String `tfsdk:"source_code_hash"`
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"content": schema.StringAttribute{
			Optional:    true,
			Description: "The contents of the Compute deployment package as a base64-encoded string. Conflicts with `filename`.",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("filename"),
				),
			},
		},
		"filename": schema.StringAttribute{
			Optional:    true,
			Description: "The path to the Compute deployment package on the local filesystem. Conflicts with `content`.",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("content"),
				),
			},
		},
		"source_code_hash": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Hash of the package contents used to detect local package changes.",
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Compute package attached to this service version. At most one package block is supported.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

// Equal reports whether two compute package model slices represent the same package configuration.
// At most one package block is supported, so only the first element is compared when present.
// SourceCodeHash is only compared if it's explicitly set (not null/unknown) in BOTH packages,
// allowing users to optionally use it to detect package changes while ignoring API-populated values.
func Equal(a, b []Model) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 {
		return true
	}

	aModel := a[0]
	bModel := b[0]

	// Always compare content and filename
	if !stringValuesEqual(aModel.Content, bModel.Content) ||
		!stringValuesEqual(aModel.Filename, bModel.Filename) {
		return false
	}

	// Only compare SourceCodeHash if it's explicitly set in BOTH packages
	// This allows the API to populate the hash without triggering changes
	aHashSet := !aModel.SourceCodeHash.IsNull() && !aModel.SourceCodeHash.IsUnknown()
	bHashSet := !bModel.SourceCodeHash.IsNull() && !bModel.SourceCodeHash.IsUnknown()

	if aHashSet && bHashSet {
		return stringValuesEqual(aModel.SourceCodeHash, bModel.SourceCodeHash)
	}

	return true
}

func stringValuesEqual(a, b types.String) bool {
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

func ValidateInput(packages []Model) error {
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

func Update(ctx context.Context, client *fastly.Client, serviceID string, version int, packages []Model) error {
	if len(packages) == 0 {
		return nil
	}

	if err := ValidateInput(packages); err != nil {
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
		input.PackagePath = new(pkg.Filename.ValueString())
	}

	_, err := client.UpdatePackage(ctx, input)
	if err != nil {
		return fmt.Errorf("error updating Compute package for service %s version %d: %w", serviceID, version, err)
	}

	return nil
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int, current []Model) ([]Model, error) {
	if len(current) == 0 {
		return current, nil
	}

	pkg, err := client.GetPackage(ctx, &fastly.GetPackageInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		if errors.IsNotFound(err) {
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

	return []Model{result}, nil
}
