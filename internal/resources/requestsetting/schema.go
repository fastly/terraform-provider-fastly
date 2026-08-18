package requestsetting

import (
	"context"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NestedModel struct {
	Name             types.String `tfsdk:"name"`
	Action           types.String `tfsdk:"action"`
	BypassBusyWait   types.Bool   `tfsdk:"bypass_busy_wait"`
	DefaultHost      types.String `tfsdk:"default_host"`
	ForceMiss        types.Bool   `tfsdk:"force_miss"`
	ForceSSL         types.Bool   `tfsdk:"force_ssl"`
	HashKeys         types.String `tfsdk:"hash_keys"`
	MaxStaleAge      types.Int64  `tfsdk:"max_stale_age"`
	RequestCondition types.String `tfsdk:"request_condition"`
	TimerSupport     types.Bool   `tfsdk:"timer_support"`
	XFF              types.String `tfsdk:"xff"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.Action) == service.StringValue(other.Action) &&
		service.BoolValue(n.BypassBusyWait) == service.BoolValue(other.BypassBusyWait) &&
		service.StringValue(n.DefaultHost) == service.StringValue(other.DefaultHost) &&
		service.BoolValue(n.ForceMiss) == service.BoolValue(other.ForceMiss) &&
		service.BoolValue(n.ForceSSL) == service.BoolValue(other.ForceSSL) &&
		service.StringValue(n.HashKeys) == service.StringValue(other.HashKeys) &&
		service.Int64Value(n.MaxStaleAge) == service.Int64Value(other.MaxStaleAge) &&
		service.StringValue(n.RequestCondition) == service.StringValue(other.RequestCondition) &&
		service.BoolValue(n.TimerSupport) == service.BoolValue(other.TimerSupport) &&
		service.StringValue(n.XFF) == service.StringValue(other.XFF)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Unique name to refer to this Request Setting. Changing this attribute will delete and recreate the resource.",
		},
		"action": schema.StringAttribute{
			Optional:    true,
			Description: "Allows you to terminate request handling and immediately perform an action. When set it can be `lookup` or `pass` (ignore the cache completely).",
			Validators: []validator.String{
				stringvalidator.OneOfCaseInsensitive("lookup", "pass"),
			},
		},
		"bypass_busy_wait": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Disable collapsed forwarding, so you don't wait for other objects to origin. Default `false`.",
		},
		"default_host": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Sets the host header.",
		},
		"force_miss": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Force a cache miss for the request. Default `false`.",
		},
		"force_ssl": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Forces the request to use SSL (redirects a non-SSL request to SSL). Default `false`.",
		},
		"hash_keys": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Comma separated list of varnish request object fields that should be in the hash key.",
		},
		"max_stale_age": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(0),
			Description: "How old an object is allowed to be to serve `stale-if-error` or `stale-while-revalidate`, in seconds. Default `0`.",
		},
		"request_condition": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Name of already defined `condition` to determine if this request setting should be applied. Should be unique across multiple instances of `request_setting`, including any left unset (Fastly rejects more than one request setting sharing the same, or unset, `request_condition`).",
		},
		"timer_support": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Injects the X-Timer info into the request for viewing origin fetch durations. Default `false`.",
		},
		"xff": schema.StringAttribute{
			Optional:    true,
			Description: "X-Forwarded-For, should be `clear`, `leave`, `append`, `append_all`, or `overwrite`.",
			Validators: []validator.String{
				stringvalidator.OneOfCaseInsensitive("clear", "leave", "append", "append_all", "overwrite"),
			},
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Request settings attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.RequestSetting, error) {
	return client.ListRequestSettings(ctx, &fastly.ListRequestSettingsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.RequestSetting) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteRequestSetting(ctx, &fastly.DeleteRequestSettingInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

// Create omits Action/XFF entirely when unset, rather than sending an empty string, since
// both are enums the Fastly API validates and may reject a blank value on creation. Update
// (below) always sends both, since that's the only way to clear a previously configured
// value back to unset.
func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.RequestSetting, error) {
	name := service.StringValue(desired.Name)
	defaultHost := service.StringValue(desired.DefaultHost)
	hashKeys := service.StringValue(desired.HashKeys)
	maxStaleAge := int(service.Int64Value(desired.MaxStaleAge))
	requestCondition := service.StringValue(desired.RequestCondition)
	bypassBusyWait := fastly.Compatibool(service.BoolValue(desired.BypassBusyWait))
	forceMiss := fastly.Compatibool(service.BoolValue(desired.ForceMiss))
	forceSSL := fastly.Compatibool(service.BoolValue(desired.ForceSSL))
	timerSupport := fastly.Compatibool(service.BoolValue(desired.TimerSupport))

	return client.CreateRequestSetting(ctx, &fastly.CreateRequestSettingInput{
		ServiceID:        serviceID,
		ServiceVersion:   version,
		Name:             &name,
		Action:           actionPointer(desired.Action),
		BypassBusyWait:   &bypassBusyWait,
		DefaultHost:      &defaultHost,
		ForceMiss:        &forceMiss,
		ForceSSL:         &forceSSL,
		HashKeys:         &hashKeys,
		MaxStaleAge:      &maxStaleAge,
		RequestCondition: &requestCondition,
		TimerSupport:     &timerSupport,
		XForwardedFor:    xffPointer(desired.XFF),
	})
}

// actionPointer returns nil for a null/unknown/empty action, so Create omits the field
// rather than sending an invalid empty string for an enum the Fastly API validates. The
// value is lowercased since the validator accepts any case (e.g. LOOKUP) but the API expects
// the lowercase enum value.
func actionPointer(v types.String) *fastly.RequestSettingAction {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	action := fastly.RequestSettingAction(strings.ToLower(v.ValueString()))
	return &action
}

// xffPointer returns nil for a null/unknown/empty value, so Create omits the field rather
// than sending an invalid empty string for an enum the Fastly API validates.
func xffPointer(v types.String) *fastly.RequestSettingXFF {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	xff := fastly.RequestSettingXFF(strings.ToLower(v.ValueString()))
	return &xff
}

func (o ops) Equal(desired NestedModel, remote *fastly.RequestSetting) bool {
	return desired.ModelsEqual(o.ToModel(remote))
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.RequestSetting, error) {
	action := fastly.RequestSettingAction(strings.ToLower(service.StringValue(desired.Action)))
	xff := fastly.RequestSettingXFF(strings.ToLower(service.StringValue(desired.XFF)))
	defaultHost := service.StringValue(desired.DefaultHost)
	hashKeys := service.StringValue(desired.HashKeys)
	maxStaleAge := int(service.Int64Value(desired.MaxStaleAge))
	requestCondition := service.StringValue(desired.RequestCondition)
	bypassBusyWait := fastly.Compatibool(service.BoolValue(desired.BypassBusyWait))
	forceMiss := fastly.Compatibool(service.BoolValue(desired.ForceMiss))
	forceSSL := fastly.Compatibool(service.BoolValue(desired.ForceSSL))
	timerSupport := fastly.Compatibool(service.BoolValue(desired.TimerSupport))

	return client.UpdateRequestSetting(ctx, &fastly.UpdateRequestSettingInput{
		ServiceID:        serviceID,
		ServiceVersion:   version,
		Name:             service.StringValue(desired.Name),
		Action:           &action,
		BypassBusyWait:   &bypassBusyWait,
		DefaultHost:      &defaultHost,
		ForceMiss:        &forceMiss,
		ForceSSL:         &forceSSL,
		HashKeys:         &hashKeys,
		MaxStaleAge:      &maxStaleAge,
		RequestCondition: &requestCondition,
		TimerSupport:     &timerSupport,
		XForwardedFor:    &xff,
	})
}

func (o ops) ToModel(api *fastly.RequestSetting) NestedModel {
	model := NestedModel{
		Name:             types.StringValue(fastly.ToValue(api.Name)),
		BypassBusyWait:   types.BoolValue(fastly.ToValue(api.BypassBusyWait)),
		DefaultHost:      types.StringValue(fastly.ToValue(api.DefaultHost)),
		ForceMiss:        types.BoolValue(fastly.ToValue(api.ForceMiss)),
		ForceSSL:         types.BoolValue(fastly.ToValue(api.ForceSSL)),
		HashKeys:         types.StringValue(fastly.ToValue(api.HashKeys)),
		MaxStaleAge:      types.Int64Value(int64(fastly.ToValue(api.MaxStaleAge))),
		RequestCondition: types.StringValue(fastly.ToValue(api.RequestCondition)),
		TimerSupport:     types.BoolValue(fastly.ToValue(api.TimerSupport)),
	}
	if api.Action != nil && *api.Action != "" {
		model.Action = types.StringValue(string(*api.Action))
	} else {
		model.Action = types.StringNull()
	}
	if api.XForwardedFor != nil && *api.XForwardedFor != "" {
		model.XFF = types.StringValue(string(*api.XForwardedFor))
	} else {
		model.XFF = types.StringNull()
	}
	return model
}

var reconciler = &reconcile.Resource[NestedModel, fastly.RequestSetting]{
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
