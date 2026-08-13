package dictionary

import (
	"context"
	"fmt"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultWriteOnly    = false
	DefaultForceDestroy = false
)

type NestedModel struct {
	Name         types.String `tfsdk:"name"`
	DictionaryID types.String `tfsdk:"dictionary_id"`
	WriteOnly    types.Bool   `tfsdk:"write_only"`
	ForceDestroy types.Bool   `tfsdk:"force_destroy"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.BoolValue(n.WriteOnly) == service.BoolValue(other.WriteOnly)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A unique name to identify this dictionary. Must be unique within the service.",
		},
		"dictionary_id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of the dictionary.",
		},
		"write_only": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultWriteOnly),
			Description: "Determines if items in the dictionary are readable or not. Default `false`. Changing this attribute deletes and recreates the dictionary, discarding its current items, so it is subject to the same `force_destroy` requirement as removing the dictionary.",
		},
		"force_destroy": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultForceDestroy),
			Description: "Allow the dictionary to be deleted or have `write_only` changed, even if it still contains items. Dictionary items are not recoverable once deleted, so this defaults to `false`.",
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Edge dictionaries attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Dictionary, error) {
	return client.ListDictionaries(ctx, &fastly.ListDictionariesInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.Dictionary) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteDictionary(ctx, &fastly.DeleteDictionaryInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Dictionary, error) {
	return client.CreateDictionary(ctx, &fastly.CreateDictionaryInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(desired.Name)),
		WriteOnly:      new(fastly.Compatibool(service.BoolValue(desired.WriteOnly))),
	})
}

func (o ops) Equal(desired NestedModel, remote *fastly.Dictionary) bool {
	return desired.ModelsEqual(o.ToModel(remote))
}

// Update changes write_only by deleting and recreating the dictionary: the Fastly API
// rejects an in-place update of write_only ("Not allowed to modify field 'write_only'"),
// even though UpdateDictionaryInput accepts the field. Name never differs between desired
// and remote here, since the reconciler only calls Update for a name match.
// ReconcileWithPrevious guards this path with the same force_destroy/emptiness check used
// for removing a dictionary outright, since this also discards any existing items.
func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Dictionary, error) {
	if err := o.Delete(ctx, client, serviceID, version, service.StringValue(desired.Name)); err != nil {
		return nil, err
	}
	return o.Create(ctx, client, serviceID, version, desired)
}

func (o ops) ToModel(api *fastly.Dictionary) NestedModel {
	return NestedModel{
		Name:         types.StringValue(fastly.ToValue(api.Name)),
		DictionaryID: types.StringValue(fastly.ToValue(api.DictionaryID)),
		WriteOnly:    types.BoolValue(fastly.ToValue(api.WriteOnly)),
		// ForceDestroy is configuration-only, not returned by the API.
		// Set to the default value here; preserved through ReadForVersionWithPlan.
		ForceDestroy: types.BoolValue(DefaultForceDestroy),
	}
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Dictionary]{
	Ops: ops{},
	GetName: func(m NestedModel) string {
		return service.StringValue(m.Name)
	},
	Sortable: true,
}

// ReadForVersion reads dictionaries from a service version without plan context.
// Note: force_destroy will not be preserved when using this function.
// Use ReadForVersionWithPlan when reading as part of resource state management.
func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	return ReadForVersionWithPlan(ctx, client, serviceID, version, nil)
}

// ReadForVersionWithPlan reads dictionaries from a service version and preserves
// configuration-only fields (like force_destroy) from the provided plan. This ensures
// that fields which don't round-trip through the API maintain their configured values
// in Terraform state.
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
			model.ForceDestroy = planItem.ForceDestroy
		}
		result = append(result, model)
	}

	return result, nil
}

// Reconcile applies desired dictionaries with no force_destroy/emptiness guard against a
// previous state. It is only safe to call when nothing previously existed (e.g. resource
// Create), since it will delete-and-recreate or drop a dictionary that still has
// items. Any caller reconciling against pre-existing dictionaries (e.g. Update) must use
// ReconcileWithPrevious instead, or items can be silently discarded.
func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	return reconciler.Run(ctx, client, serviceID, version, desired)
}

// ReconcileWithPrevious reconciles dictionaries while validating force_destroy requirements
// for any change that discards a dictionary's items: removing the dictionary from config
// outright, or changing write_only, which Update handles via delete-then-create since the
// API rejects an in-place write_only change. Either case either needs force_destroy=true in
// the previous state or the dictionary must contain no items, since its items are not
// recoverable once deleted. The guard mechanics (map previous/desired by name, check
// force_destroy, check emptiness, error) are shared with cdnacl via reconcile.GuardedRun.
//
// A write_only dictionary can't be inspected for items via the API, so the emptiness check
// is skipped for it - it's treated as non-empty and force_destroy is required unconditionally,
// matching the old SDKv2 provider's behavior.
func ReconcileWithPrevious(ctx context.Context, client *fastly.Client, serviceID string, version int, previous, desired []NestedModel) error {
	return reconciler.GuardedRun(
		ctx, client, serviceID, version, previous, desired,
		func(m NestedModel) bool { return service.BoolValue(m.ForceDestroy) },
		func(prev, desired NestedModel, stillPresent bool) bool {
			return !stillPresent || service.BoolValue(prev.WriteOnly) != service.BoolValue(desired.WriteOnly)
		},
		func(ctx context.Context, prev NestedModel) (bool, error) {
			if service.BoolValue(prev.WriteOnly) {
				return false, nil
			}
			return isDictionaryEmpty(ctx, client, serviceID, service.StringValue(prev.DictionaryID))
		},
		func(name string, prev NestedModel) error {
			if service.BoolValue(prev.WriteOnly) {
				return fmt.Errorf("cannot delete or change write_only for dictionary %q (ID: %s): it is write_only, so its contents cannot be inspected to verify it is empty; set force_destroy to true for this change to be applied", name, service.StringValue(prev.DictionaryID))
			}
			return fmt.Errorf("cannot delete dictionary %q (ID: %s): it contains items that must be removed first, or set force_destroy to true for this change to be applied", name, service.StringValue(prev.DictionaryID))
		},
	)
}

// Equal reports whether two dictionary lists are equivalent from the Terraform resource's
// point of view, unlike NestedModel.ModelsEqual (used for the live API comparison in
// ops.Equal), this also considers force_destroy so that a force_destroy-only config change
// is detected as a change - otherwise the caller's no-op path would reorder items from stale
// prior state and never apply the new force_destroy value.
func Equal(a, b []NestedModel) bool {
	equalIncludingForceDestroy := func(a, b NestedModel) bool {
		return a.ModelsEqual(b) && service.BoolValue(a.ForceDestroy) == service.BoolValue(b.ForceDestroy)
	}
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.Name) }, equalIncludingForceDestroy, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	return reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
}

func isDictionaryEmpty(ctx context.Context, client *fastly.Client, serviceID, dictionaryID string) (bool, error) {
	items, err := client.ListDictionaryItems(ctx, &fastly.ListDictionaryItemsInput{
		ServiceID:    serviceID,
		DictionaryID: dictionaryID,
	})
	if err != nil {
		return false, err
	}

	return len(items) == 0, nil
}
