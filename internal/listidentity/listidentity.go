// Package listidentity builds list results for resources that intentionally
// do not implement resource.ResourceWithIdentity.
//
// list.ListRequest.NewListResult derives the identity's Terraform type from
// req.ResourceIdentitySchema, which the framework only populates when the
// underlying managed resource implements resource.ResourceWithIdentity. For
// resources that don't (identity was removed from the versioned resources in
// this provider because a resource's identity can't safely contain a mutable
// field like service version), that schema is nil and NewListResult panics.
// NewResult builds the same result without that dependency, using a fixed,
// empty identity schema so list.ListResult.Identity stays non-nil (required
// by the framework) with no structured identity attributes.
package listidentity

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var emptySchema = identityschema.Schema{Attributes: map[string]identityschema.Attribute{}}

// NewResult builds a list.ListResult in place of req.NewListResult, for
// managed resources with no identity schema.
func NewResult(ctx context.Context, req list.ListRequest) list.ListResult {
	return list.ListResult{
		Resource: &tfsdk.Resource{
			Raw:    tftypes.NewValue(req.ResourceSchema.Type().TerraformType(ctx), nil),
			Schema: req.ResourceSchema,
		},
		Identity: &tfsdk.ResourceIdentity{
			Raw:    tftypes.NewValue(emptySchema.Type().TerraformType(ctx), nil),
			Schema: emptySchema,
		},
		Diagnostics: nil,
	}
}
