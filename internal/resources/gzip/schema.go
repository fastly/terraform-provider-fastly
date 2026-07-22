package gzip

import (
	"context"
	"sort"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NestedModel struct {
	Name           types.String `tfsdk:"name"`
	CacheCondition types.String `tfsdk:"cache_condition"`
	ContentTypes   types.List   `tfsdk:"content_types"`
	Extensions     types.List   `tfsdk:"extensions"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.CacheCondition) == service.StringValue(other.CacheCondition) &&
		stringListEqual(n.ContentTypes, other.ContentTypes) &&
		stringListEqual(n.Extensions, other.Extensions)
}

// stringListEqual treats a null/unknown/empty list as equivalent to any other
// empty representation, since the Fastly API silently defaults content_types
// and extensions when they're unset, which would otherwise appear as drift.
func stringListEqual(a, b types.List) bool {
	if listUnset(a) && listUnset(b) {
		return true
	}
	return a.Equal(b)
}

// listUnset reports whether a list should be treated as "not provided" -
// null, unknown, or an explicit empty list all send the same empty wire
// value to the Fastly API (see joinStringList).
func listUnset(l types.List) bool {
	return l.IsNull() || l.IsUnknown() || len(l.Elements()) == 0
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A name to refer to this gzip condition. Changing this attribute will delete and recreate the resource.",
		},
		"cache_condition": schema.StringAttribute{
			Optional:    true,
			Description: "Name of already defined `condition` controlling when this gzip configuration applies. This `condition` must be of type `CACHE`.",
		},
		"content_types": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "The content-type for each type of content you wish to have dynamically gzip'ed. Example: `[\"text/html\", \"text/css\"]`.",
		},
		"extensions": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "File extensions for each file type to dynamically gzip. Example: `[\"css\", \"js\"]`.",
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Gzip configurations attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Gzip, error) {
	return client.ListGzips(ctx, &fastly.ListGzipsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.Gzip) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteGzip(ctx, &fastly.DeleteGzipInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Gzip, error) {
	name := service.StringValue(desired.Name)
	cacheCondition := service.StringValue(desired.CacheCondition)

	return client.CreateGzip(ctx, &fastly.CreateGzipInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           &name,
		CacheCondition: &cacheCondition,
		ContentTypes:   joinStringList(desired.ContentTypes),
		Extensions:     joinStringList(desired.Extensions),
	})
}

// Equal ignores content_types/extensions on the remote side when desired leaves them
// unset (null, unknown, or an explicit empty list), since the Fastly API silently
// substitutes a large default list for either field when it's omitted or empty.
// Comparing against that default would otherwise produce a permanent diff for any
// gzip config that doesn't set them explicitly.
func (o ops) Equal(desired NestedModel, remote *fastly.Gzip) bool {
	remoteModel := o.ToModel(remote)
	if listUnset(desired.ContentTypes) {
		remoteModel.ContentTypes = desired.ContentTypes
	}
	if listUnset(desired.Extensions) {
		remoteModel.Extensions = desired.Extensions
	}
	return desired.ModelsEqual(remoteModel)
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Gzip, error) {
	cacheCondition := service.StringValue(desired.CacheCondition)

	return client.UpdateGzip(ctx, &fastly.UpdateGzipInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(desired.Name),
		CacheCondition: &cacheCondition,
		ContentTypes:   joinStringList(desired.ContentTypes),
		Extensions:     joinStringList(desired.Extensions),
	})
}

func (o ops) ToModel(api *fastly.Gzip) NestedModel {
	model := NestedModel{
		Name:         types.StringValue(fastly.ToValue(api.Name)),
		ContentTypes: stringListValue(api.ContentTypes),
		Extensions:   stringListValue(api.Extensions),
	}
	if api.CacheCondition != nil && *api.CacheCondition != "" {
		model.CacheCondition = types.StringValue(*api.CacheCondition)
	} else {
		model.CacheCondition = types.StringNull()
	}
	return model
}

// joinStringList converts a Terraform list of strings into the space-separated
// wire format the Fastly API expects, always returning a non-nil pointer. An
// omitted or emptied list sends an empty string, which the API treats as unset
// and responds with its own default list rather than clearing the remote value.
// Once a value has been set remotely, sending an empty string again does not
// revert it back to that default - see ops.Equal, which skips Update when the
// desired value is unset so the remote value is left untouched.
func joinStringList(l types.List) *string {
	if l.IsNull() || l.IsUnknown() {
		empty := ""
		return &empty
	}

	elems := l.Elements()
	parts := make([]string, 0, len(elems))
	for _, e := range elems {
		if s, ok := e.(types.String); ok {
			parts = append(parts, s.ValueString())
		}
	}

	joined := strings.Join(parts, " ")
	return &joined
}

// stringListValue converts the Fastly API's space-separated wire format back
// into a Terraform list, mapping an unset/empty value to null so config that
// omits the attribute doesn't drift against state.
func stringListValue(raw *string) types.List {
	if raw == nil || *raw == "" {
		return types.ListNull(types.StringType)
	}

	parts := strings.Split(*raw, " ")
	elems := make([]attr.Value, len(parts))
	for i, p := range parts {
		elems[i] = types.StringValue(p)
	}
	return types.ListValueMust(types.StringType, elems)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Gzip]{
	Ops: ops{},
	GetName: func(m NestedModel) string {
		return service.StringValue(m.Name)
	},
	Sortable: true,
}

// ReadForVersion reads gzip configurations without plan awareness. Prefer
// ReadForVersionWithPlan when reading as part of resource state management.
func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	return ReadForVersionWithPlan(ctx, client, serviceID, version, nil)
}

// ReadForVersionWithPlan reads gzip configurations from a service version and nulls out
// content_types/extensions on any item whose plan left them unset (null, unknown, or an
// explicit empty list), since the Fastly API silently substitutes a large default list
// for either field when it's omitted or empty. Without this, config that never sets these
// fields would show a permanent diff, and an explicit empty list would produce a
// "provider produced inconsistent result after apply" error against the planned value.
func ReadForVersionWithPlan(ctx context.Context, client *fastly.Client, serviceID string, version int, plan []NestedModel) ([]NestedModel, error) {
	remote, err := ops{}.List(ctx, client, serviceID, version)
	if err != nil {
		return nil, err
	}

	planByName := make(map[string]NestedModel, len(plan))
	for _, p := range plan {
		planByName[service.StringValue(p.Name)] = p
	}

	result := make([]NestedModel, 0, len(remote))
	for _, item := range remote {
		model := ops{}.ToModel(item)
		if planItem, exists := planByName[service.StringValue(model.Name)]; exists {
			if listUnset(planItem.ContentTypes) {
				model.ContentTypes = planItem.ContentTypes
			}
			if listUnset(planItem.Extensions) {
				model.Extensions = planItem.Extensions
			}
		}
		result = append(result, model)
	}

	sort.Slice(result, func(i, j int) bool {
		return service.StringValue(result[i].Name) < service.StringValue(result[j].Name)
	})

	return result, nil
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
