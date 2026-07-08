package resourcelink

import (
	"context"
	"maps"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NestedModel struct {
	Name       types.String `tfsdk:"name"`
	ResourceID types.String `tfsdk:"resource_id"`
	LinkID     types.String `tfsdk:"link_id"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.ResourceID) == service.StringValue(other.ResourceID)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name the service will use to open the linked resource from Compute code (e.g. a KV Store or Config Store SDK lookup). This is an alias and does not need to match the name of the underlying resource.",
		},
		"resource_id": schema.StringAttribute{
			Required:    true,
			Description: "The ID of the shared resource to link (e.g. the ID of a KV Store or Config Store).",
		},
		"link_id": schema.StringAttribute{
			Computed:    true,
			Description: "An alphanumeric string identifying this resource link.",
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
		},
	}
	maps.Copy(attrs, CommonAttributes())
	// resource_id identifies which underlying resource is linked; the update API can only
	// rename an existing link, so pointing it at a different resource requires replacement.
	resourceIDAttr := attrs["resource_id"].(schema.StringAttribute)
	resourceIDAttr.PlanModifiers = []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
	attrs["resource_id"] = resourceIDAttr
	return attrs
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Shared resources (such as KV Stores or Config Stores) linked to this service, making them accessible from Compute code.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Resource, error) {
	return client.ListResources(ctx, &fastly.ListResourcesInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

// GetName keys reconciliation on the linked resource's ID rather than the alias name,
// since that's the identity the Fastly API treats as stable: UpdateResource can rename
// a link in place, but it cannot repoint an existing link at a different resource.
func (o ops) GetName(api *fastly.Resource) string {
	return fastly.ToValue(api.ResourceID)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, resourceID string) error {
	linkID, err := findLinkID(ctx, client, serviceID, version, resourceID)
	if err != nil {
		return err
	}
	if linkID == "" {
		return nil
	}

	return client.DeleteResource(ctx, &fastly.DeleteResourceInput{
		ResourceID:     linkID,
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Resource, error) {
	return client.CreateResource(ctx, BuildCreateInput(serviceID, version, desired))
}

func (o ops) Equal(desired NestedModel, remote *fastly.Resource) bool {
	return desired.ModelsEqual(FlattenToNestedModel(remote))
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Resource, error) {
	linkID, err := findLinkID(ctx, client, serviceID, version, service.StringValue(desired.ResourceID))
	if err != nil {
		return nil, err
	}

	return client.UpdateResource(ctx, &fastly.UpdateResourceInput{
		ResourceID:     linkID,
		Name:           desired.Name.ValueStringPointer(),
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) ToModel(api *fastly.Resource) NestedModel {
	return FlattenToNestedModel(api)
}

// findLinkID resolves the link's own ID (what the Fastly API calls "resource_id" in
// Get/Update/Delete requests) from the ID of the resource it points at (what the API
// instead calls "resource_id" in Create requests). Confusingly, these are two different
// values that share a field name depending on which endpoint you're calling.
func findLinkID(ctx context.Context, client *fastly.Client, serviceID string, version int, resourceID string) (string, error) {
	items, err := ops{}.List(ctx, client, serviceID, version)
	if err != nil {
		return "", err
	}

	for _, item := range items {
		if fastly.ToValue(item.ResourceID) == resourceID {
			return fastly.ToValue(item.LinkID), nil
		}
	}

	return "", nil
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Resource]{
	Ops: ops{},
	GetName: func(m NestedModel) string {
		return service.StringValue(m.ResourceID)
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
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.ResourceID) }, NestedModel.ModelsEqual, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	return reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.ResourceID) })
}
