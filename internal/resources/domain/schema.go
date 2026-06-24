package domain

import (
	"context"
	"maps"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NestedModel struct {
	Name    types.String `tfsdk:"name"`
	Comment types.String `tfsdk:"comment"`
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The domain that this service responds to.",
		},
		"comment": schema.StringAttribute{
			Optional:    true,
			Description: "Optional comment for the domain.",
		},
	}
}

func ResourceAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Terraform resource identifier.",
		},
		"service_id": schema.StringAttribute{
			Required:    true,
			Description: "Fastly service ID.",
		},
		"version": schema.Int64Attribute{
			Required:    true,
			Description: "Writable Fastly service version to modify.",
		},
	}
	maps.Copy(attrs, CommonAttributes())
	return attrs
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Domains attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Domain, error) {
	return client.ListDomains(ctx, &fastly.ListDomainsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.Domain) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteDomain(ctx, &fastly.DeleteDomainInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Domain, error) {
	comment := ""
	if !desired.Comment.IsNull() && !desired.Comment.IsUnknown() {
		comment = desired.Comment.ValueString()
	}

	name := desired.Name.ValueString()
	input := &fastly.CreateDomainInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           &name,
	}
	if !desired.Comment.IsNull() && !desired.Comment.IsUnknown() {
		input.Comment = &comment
	}

	return client.CreateDomain(ctx, input)
}

func (o ops) Equal(desired NestedModel, remote *fastly.Domain) bool {
	if desired.Name.ValueString() != fastly.ToValue(remote.Name) {
		return false
	}

	desiredComment := ""
	if !desired.Comment.IsNull() && !desired.Comment.IsUnknown() {
		desiredComment = desired.Comment.ValueString()
	}

	remoteComment := ""
	if remote.Comment != nil {
		remoteComment = *remote.Comment
	}

	return desiredComment == remoteComment
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Domain, error) {
	comment := ""
	if !desired.Comment.IsNull() && !desired.Comment.IsUnknown() {
		comment = desired.Comment.ValueString()
	}

	name := desired.Name.ValueString()
	return client.UpdateDomain(ctx, &fastly.UpdateDomainInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
		Comment:        &comment,
	})
}

func (o ops) ToModel(api *fastly.Domain) NestedModel {
	model := NestedModel{
		Name: types.StringValue(fastly.ToValue(api.Name)),
	}
	if api.Comment != nil && *api.Comment != "" {
		model.Comment = types.StringValue(*api.Comment)
	} else {
		model.Comment = types.StringNull()
	}
	return model
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Domain]{
	Ops: ops{},
	GetName: func(m NestedModel) string {
		return m.Name.ValueString()
	},
	Sortable: true,
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	return reconciler.ReadForVersion(ctx, client, serviceID, version)
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	return reconciler.Run(ctx, client, serviceID, version, desired)
}

func Equal(a, b []NestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return m.Name.ValueString() }, modelsEqual, true)
}

func modelsEqual(a, b NestedModel) bool {
	if a.Name.ValueString() != b.Name.ValueString() {
		return false
	}

	ac := ""
	if !a.Comment.IsNull() && !a.Comment.IsUnknown() {
		ac = a.Comment.ValueString()
	}

	bc := ""
	if !b.Comment.IsNull() && !b.Comment.IsUnknown() {
		bc = b.Comment.ValueString()
	}

	return ac == bc
}
