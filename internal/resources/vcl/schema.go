package vcl

import (
	"context"
	"maps"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NestedModel struct {
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
	Main    types.Bool   `tfsdk:"main"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		ContentEqual(service.StringValue(n.Content), service.StringValue(other.Content)) &&
		service.BoolValue(n.Main) == service.BoolValue(other.Main)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A unique name for this custom VCL file. Included VCL files must be referenced by this exact name from the main VCL file.",
		},
		"content": schema.StringAttribute{
			Required:    true,
			Description: "The custom VCL source code to upload. Commonly configured with file(\"${path.module}/main.vcl\") or templatefile(...).",
		},
		"main": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultMain),
			Description: "Whether this custom VCL file is the main configuration. Exactly one configured custom VCL file must be marked as main.",
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
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"version": schema.Int64Attribute{
			Required:    true,
			Description: "Writable Fastly service version to modify.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
	}
	maps.Copy(attrs, CommonAttributes())

	nameAttr := attrs["name"].(schema.StringAttribute)
	nameAttr.PlanModifiers = []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
	attrs["name"] = nameAttr

	mainAttr := attrs["main"].(schema.BoolAttribute)
	mainAttr.PlanModifiers = []planmodifier.Bool{
		boolplanmodifier.RequiresReplace(),
	}
	attrs["main"] = mainAttr

	return attrs
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Custom VCL files attached to this service. This is modeled as a list so Terraform can render changes to `content` as an in-place string diff instead of a whole set element delete/add.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.VCL, error) {
	return client.ListVCLs(ctx, &fastly.ListVCLsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.VCL) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteVCL(ctx, &fastly.DeleteVCLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.VCL, error) {
	return create(ctx, client, serviceID, version, desired)
}

func (o ops) Equal(desired NestedModel, remote *fastly.VCL) bool {
	return desired.ModelsEqual(FlattenToNestedModel(remote))
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.VCL, error) {
	remote, err := client.GetVCL(ctx, &fastly.GetVCLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(desired.Name),
	})
	if err != nil {
		return nil, err
	}

	// go-fastly's UpdateVCLInput updates content but not main. When the main flag
	// changes inside an automatic service resource, recreate the file with the
	// desired flag while keeping Terraform's user-visible diff focused on the
	// changed nested attributes.
	if service.BoolValue(desired.Main) != service.BoolValue(FlattenToNestedModel(remote).Main) {
		if err := o.Delete(ctx, client, serviceID, version, service.StringValue(desired.Name)); err != nil {
			return nil, err
		}
		return o.Create(ctx, client, serviceID, version, desired)
	}

	content := service.StringValue(desired.Content)
	return client.UpdateVCL(ctx, &fastly.UpdateVCLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(desired.Name),
		Content:        &content,
	})
}

func (o ops) ToModel(api *fastly.VCL) NestedModel {
	return FlattenToNestedModel(api)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.VCL]{
	Ops: ops{},
	GetName: func(m NestedModel) string {
		return service.StringValue(m.Name)
	},
	Sortable: true,
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	return reconciler.ReadForVersion(ctx, client, serviceID, version)
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	if err := Validate(desired); err != nil {
		return err
	}
	return reconciler.Run(ctx, client, serviceID, version, desired)
}

func Equal(a, b []NestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.Name) }, NestedModel.ModelsEqual, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	return reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
}

func create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.VCL, error) {
	name := service.StringValue(desired.Name)
	content := service.StringValue(desired.Content)
	main := service.BoolValue(desired.Main)

	return client.CreateVCL(ctx, &fastly.CreateVCLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           &name,
		Content:        &content,
		Main:           &main,
	})
}
