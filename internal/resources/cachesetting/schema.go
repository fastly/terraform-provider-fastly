package cachesetting

import (
	"context"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NestedModel struct {
	Name           types.String `tfsdk:"name"`
	Action         types.String `tfsdk:"action"`
	CacheCondition types.String `tfsdk:"cache_condition"`
	StaleTTL       types.Int64  `tfsdk:"stale_ttl"`
	TTL            types.Int64  `tfsdk:"ttl"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.Action) == service.StringValue(other.Action) &&
		service.StringValue(n.CacheCondition) == service.StringValue(other.CacheCondition) &&
		service.Int64Value(n.StaleTTL) == service.Int64Value(other.StaleTTL) &&
		service.Int64Value(n.TTL) == service.Int64Value(other.TTL)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Unique name for this Cache Setting. Changing this attribute will delete and recreate the resource.",
		},
		"action": schema.StringAttribute{
			Optional:    true,
			Description: "One of `cache`, `pass`, or `restart`, as defined on Fastly's documentation under [\"Caching action descriptions\"](https://docs.fastly.com/en/guides/controlling-caching#caching-action-descriptions).",
			Validators: []validator.String{
				stringvalidator.OneOfCaseInsensitive("cache", "pass", "restart"),
			},
		},
		"cache_condition": schema.StringAttribute{
			Optional:    true,
			Description: "Name of already defined `condition` used to test whether this settings object should be used. This `condition` must be of type `CACHE`.",
		},
		"stale_ttl": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(0),
			Description: "Max \"Time To Live\" (in seconds) for stale (unreachable) objects. Default `0`.",
		},
		"ttl": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(0),
			Description: "The Time-To-Live (TTL, in seconds) for the object. Default `0`.",
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Cache settings attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.CacheSetting, error) {
	return client.ListCacheSettings(ctx, &fastly.ListCacheSettingsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.CacheSetting) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteCacheSetting(ctx, &fastly.DeleteCacheSettingInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

// Create omits Action entirely when unset, rather than sending an empty string, since the
// Fastly API validates Action as one of cache/pass/restart and may reject a blank value on
// creation. Update (below) always sends Action, since that's the only way to clear a
// previously configured value back to unset.
func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.CacheSetting, error) {
	name := service.StringValue(desired.Name)
	cacheCondition := service.StringValue(desired.CacheCondition)
	ttl := int(service.Int64Value(desired.TTL))
	staleTTL := int(service.Int64Value(desired.StaleTTL))

	return client.CreateCacheSetting(ctx, &fastly.CreateCacheSettingInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           &name,
		Action:         actionPointer(desired.Action),
		CacheCondition: &cacheCondition,
		TTL:            &ttl,
		StaleTTL:       &staleTTL,
	})
}

// actionPointer returns nil for a null/unknown/empty action, so Create omits the field
// rather than sending an invalid empty string for an enum the Fastly API validates. The
// value is lowercased since the validator accepts any case (e.g. PASS) but the API expects
// the lowercase enum value.
func actionPointer(v types.String) *fastly.CacheSettingAction {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	action := fastly.CacheSettingAction(strings.ToLower(v.ValueString()))
	return &action
}

func (o ops) Equal(desired NestedModel, remote *fastly.CacheSetting) bool {
	return desired.ModelsEqual(o.ToModel(remote))
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.CacheSetting, error) {
	action := fastly.CacheSettingAction(strings.ToLower(service.StringValue(desired.Action)))
	cacheCondition := service.StringValue(desired.CacheCondition)
	ttl := int(service.Int64Value(desired.TTL))
	staleTTL := int(service.Int64Value(desired.StaleTTL))

	return client.UpdateCacheSetting(ctx, &fastly.UpdateCacheSettingInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(desired.Name),
		Action:         &action,
		CacheCondition: &cacheCondition,
		TTL:            &ttl,
		StaleTTL:       &staleTTL,
	})
}

func (o ops) ToModel(api *fastly.CacheSetting) NestedModel {
	model := NestedModel{
		Name:     types.StringValue(fastly.ToValue(api.Name)),
		TTL:      types.Int64Value(int64(fastly.ToValue(api.TTL))),
		StaleTTL: types.Int64Value(int64(fastly.ToValue(api.StaleTTL))),
	}
	if api.Action != nil && *api.Action != "" {
		model.Action = types.StringValue(string(*api.Action))
	} else {
		model.Action = types.StringNull()
	}
	if api.CacheCondition != nil && *api.CacheCondition != "" {
		model.CacheCondition = types.StringValue(*api.CacheCondition)
	} else {
		model.CacheCondition = types.StringNull()
	}
	return model
}

var reconciler = &reconcile.Resource[NestedModel, fastly.CacheSetting]{
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
