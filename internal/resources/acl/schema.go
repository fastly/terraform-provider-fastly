package acl

import (
	"context"
	"fmt"
	"maps"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultForceDestroy = false
)

type NestedModel struct {
	Name         types.String `tfsdk:"name"`
	ACLID        types.String `tfsdk:"acl_id"`
	ForceDestroy types.Bool   `tfsdk:"force_destroy"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.ACLID) == service.StringValue(other.ACLID)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A unique name to identify this ACL. Must be unique within the service.",
		},
		"acl_id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of the ACL.",
		},
		"force_destroy": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultForceDestroy),
			Description: "Allow the ACL to be deleted, even if it contains entries. Default `false`.",
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
	// Add RequiresReplace to name for standalone resource only
	nameAttr := attrs["name"].(schema.StringAttribute)
	nameAttr.PlanModifiers = []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
	attrs["name"] = nameAttr
	return attrs
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "ACLs attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.ACL, error) {
	return client.ListACLs(ctx, &fastly.ListACLsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.ACL) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteACL(ctx, &fastly.DeleteACLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.ACL, error) {
	input := BuildCreateInput(serviceID, version, desired)
	return client.CreateACL(ctx, input)
}

func (o ops) Equal(desired NestedModel, remote *fastly.ACL) bool {
	remoteModel := FlattenToNestedModel(remote)
	return desired.ModelsEqual(remoteModel)
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.ACL, error) {
	return nil, nil
}

func (o ops) ToModel(api *fastly.ACL) NestedModel {
	return FlattenToNestedModel(api)
}

func (o ops) PreserveComputed(desired NestedModel, remote *fastly.ACL) NestedModel {
	result := FlattenToNestedModel(remote)
	result.ForceDestroy = desired.ForceDestroy
	return result
}

var reconciler = &reconcile.Resource[NestedModel, fastly.ACL]{
	Ops: ops{},
	GetName: func(m NestedModel) string {
		return service.StringValue(m.Name)
	},
	Sortable: true,
}

// ReadForVersion reads ACLs from a service version without plan context.
// Note: force_destroy will not be preserved when using this function.
// Use ReadForVersionWithPlan when reading as part of resource state management.
func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	return ReadForVersionWithPlan(ctx, client, serviceID, version, nil)
}

// ReadForVersionWithPlan reads ACLs from a service version and preserves configuration-only
// fields (like force_destroy) from the provided plan. This ensures that fields which don't
// round-trip through the API maintain their configured values in Terraform state.
func ReadForVersionWithPlan(ctx context.Context, client *fastly.Client, serviceID string, version int, plan []NestedModel) ([]NestedModel, error) {
	remote, err := ops{}.List(ctx, client, serviceID, version)
	if err != nil {
		return nil, err
	}

	planByName := make(map[string]NestedModel)
	for _, p := range plan {
		planByName[service.StringValue(p.Name)] = p
	}

	result := make([]NestedModel, 0, len(remote))
	for _, item := range remote {
		model := FlattenToNestedModel(item)
		if planItem, exists := planByName[service.StringValue(model.Name)]; exists {
			model.ForceDestroy = planItem.ForceDestroy
		}
		result = append(result, model)
	}

	return result, nil
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	return reconciler.Run(ctx, client, serviceID, version, desired)
}

// ReconcileWithPrevious reconciles ACLs while validating force_destroy requirements for deletions.
// It checks that ACLs being removed either have force_destroy=true in their previous state or are empty.
func ReconcileWithPrevious(ctx context.Context, client *fastly.Client, serviceID string, version int, previous, desired []NestedModel) error {
	previousByName := make(map[string]NestedModel)
	for _, p := range previous {
		previousByName[service.StringValue(p.Name)] = p
	}

	desiredByName := make(map[string]NestedModel)
	for _, d := range desired {
		desiredByName[service.StringValue(d.Name)] = d
	}

	for name, prevACL := range previousByName {
		if _, exists := desiredByName[name]; !exists {
			if !service.BoolValue(prevACL.ForceDestroy) {
				isEmpty, err := isACLEmpty(ctx, serviceID, service.StringValue(prevACL.ACLID), client)
				if err != nil {
					return fmt.Errorf("error checking if ACL is empty before removal: %w", err)
				}

				if !isEmpty {
					return fmt.Errorf("cannot delete ACL %q (ID: %s): list is not empty. The ACL contains entries that must be removed first, or set force_destroy to true before removing the ACL", name, service.StringValue(prevACL.ACLID))
				}
			}
		}
	}

	return reconciler.Run(ctx, client, serviceID, version, desired)
}

func Equal(a, b []NestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.Name) }, NestedModel.ModelsEqual, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	return reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
}
