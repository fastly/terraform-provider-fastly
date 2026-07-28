package condition

import (
	"context"
	"maps"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const DefaultPriority = 10

type NestedModel struct {
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Statement types.String `tfsdk:"statement"`
	Priority  types.Int64  `tfsdk:"priority"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.Type) == service.StringValue(other.Type) &&
		service.StringValue(n.Statement) == service.StringValue(other.Statement) &&
		service.Int64Value(n.Priority) == service.Int64Value(other.Priority)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A name to refer to this condition. Changing this attribute will delete and recreate the resource.",
		},
		"type": schema.StringAttribute{
			Required:    true,
			Description: "Type of condition. Must be one of `REQUEST`, `RESPONSE`, `CACHE`, or `PREFETCH`.",
			Validators: []validator.String{
				stringvalidator.OneOf("REQUEST", "RESPONSE", "CACHE", "PREFETCH"),
			},
		},
		"statement": schema.StringAttribute{
			Required:    true,
			Description: "A conditional expression in VCL used to determine if the condition is met.",
		},
		"priority": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultPriority),
			Description: "A number used to determine the order in which multiple conditions execute. Lower numbers execute first. Default `10`.",
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
		Description: "Conditions attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Condition, error) {
	return client.ListConditions(ctx, &fastly.ListConditionsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.Condition) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteCondition(ctx, &fastly.DeleteConditionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Condition, error) {
	return client.CreateCondition(ctx, BuildCreateInput(serviceID, version, desired))
}

func (o ops) Equal(desired NestedModel, remote *fastly.Condition) bool {
	remoteModel := FlattenToNestedModel(remote)
	return desired.ModelsEqual(remoteModel)
}

// Update applies changes to a condition. The Fastly API does not support changing a condition's
// type in place, so a type change is applied as a delete followed by a create instead of a PUT.
func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Condition, error) {
	remote, err := client.GetCondition(ctx, &fastly.GetConditionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(desired.Name),
	})
	if err != nil {
		return nil, err
	}

	if fastly.ToValue(remote.Type) != service.StringValue(desired.Type) {
		if err := client.DeleteCondition(ctx, &fastly.DeleteConditionInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Name:           service.StringValue(desired.Name),
		}); err != nil {
			return nil, err
		}
		return o.Create(ctx, client, serviceID, version, desired)
	}

	return client.UpdateCondition(ctx, BuildUpdateInput(serviceID, version, desired))
}

func (o ops) ToModel(api *fastly.Condition) NestedModel {
	return FlattenToNestedModel(api)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Condition]{
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
	return reconciler.Run(ctx, client, serviceID, version, desired)
}

func Equal(a, b []NestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.Name) }, NestedModel.ModelsEqual, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	return reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
}

// trimStatement strips leading/trailing whitespace so HEREDOC-formatted VCL statements with a
// trailing newline don't produce a permanent diff against the API's stored value.
func trimStatement(s string) string {
	return strings.TrimSpace(s)
}
