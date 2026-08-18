package header

import (
	"context"

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

const (
	DefaultCacheCondition    = ""
	DefaultIgnoreIfSet       = false
	DefaultPriority          = 100
	DefaultRegex             = ""
	DefaultRequestCondition  = ""
	DefaultResponseCondition = ""
	DefaultSource            = ""
	DefaultSubstitution      = ""
)

type NestedModel struct {
	Name              types.String `tfsdk:"name"`
	Action            types.String `tfsdk:"action"`
	Type              types.String `tfsdk:"type"`
	Destination       types.String `tfsdk:"destination"`
	CacheCondition    types.String `tfsdk:"cache_condition"`
	IgnoreIfSet       types.Bool   `tfsdk:"ignore_if_set"`
	Priority          types.Int64  `tfsdk:"priority"`
	Regex             types.String `tfsdk:"regex"`
	RequestCondition  types.String `tfsdk:"request_condition"`
	ResponseCondition types.String `tfsdk:"response_condition"`
	Source            types.String `tfsdk:"source"`
	Substitution      types.String `tfsdk:"substitution"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.Action) == service.StringValue(other.Action) &&
		service.StringValue(n.Type) == service.StringValue(other.Type) &&
		service.StringValue(n.Destination) == service.StringValue(other.Destination) &&
		service.StringValue(n.CacheCondition) == service.StringValue(other.CacheCondition) &&
		service.BoolValue(n.IgnoreIfSet) == service.BoolValue(other.IgnoreIfSet) &&
		service.Int64Value(n.Priority) == service.Int64Value(other.Priority) &&
		service.StringValue(n.Regex) == service.StringValue(other.Regex) &&
		service.StringValue(n.RequestCondition) == service.StringValue(other.RequestCondition) &&
		service.StringValue(n.ResponseCondition) == service.StringValue(other.ResponseCondition) &&
		service.StringValue(n.Source) == service.StringValue(other.Source) &&
		service.StringValue(n.Substitution) == service.StringValue(other.Substitution)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Unique name for this Header attribute. Changing this attribute will delete and recreate the resource.",
		},
		"action": schema.StringAttribute{
			Required:    true,
			Description: "The Header manipulation action to take; must be one of `set`, `append`, `delete`, `regex`, or `regex_repeat`.",
			Validators: []validator.String{
				stringvalidator.OneOf("set", "append", "delete", "regex", "regex_repeat"),
			},
		},
		"type": schema.StringAttribute{
			Required:    true,
			Description: "The Request type on which to apply the selected Action; must be one of `request`, `fetch`, `cache`, or `response`.",
			Validators: []validator.String{
				stringvalidator.OneOf("request", "fetch", "cache", "response"),
			},
		},
		"destination": schema.StringAttribute{
			Required:    true,
			Description: "The name of the header that is going to be affected by the Action.",
		},
		"cache_condition": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultCacheCondition),
			Description: "Name of already defined `condition` to apply. This `condition` must be of type `CACHE`.",
		},
		"ignore_if_set": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultIgnoreIfSet),
			Description: "Don't add the header if it is already present. Only applies to the `set` action. Default `false`.",
		},
		"priority": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultPriority),
			Description: "Lower priorities execute first. Default `100`.",
		},
		"regex": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultRegex),
			Description: "Regular expression to use. Only applies to the `regex` and `regex_repeat` actions.",
		},
		"request_condition": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultRequestCondition),
			Description: "Name of already defined `condition` to apply. This `condition` must be of type `REQUEST`.",
		},
		"response_condition": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultResponseCondition),
			Description: "Name of already defined `condition` to apply. This `condition` must be of type `RESPONSE`. For detailed information about Conditionals, see [Fastly's Documentation on Conditionals](https://docs.fastly.com/en/guides/using-conditions).",
		},
		"source": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultSource),
			Description: "Variable to be used as a source for the header content. Does not apply to the `delete` action.",
		},
		"substitution": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultSubstitution),
			Description: "Value to substitute in place of regular expression. Only applies to the `regex` and `regex_repeat` actions.",
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Header manipulations attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Header, error) {
	return client.ListHeaders(ctx, &fastly.ListHeadersInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.Header) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteHeader(ctx, &fastly.DeleteHeaderInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Header, error) {
	return client.CreateHeader(ctx, BuildCreateInput(serviceID, version, desired))
}

func (o ops) Equal(desired NestedModel, remote *fastly.Header) bool {
	return desired.ModelsEqual(o.ToModel(remote))
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Header, error) {
	return client.UpdateHeader(ctx, BuildUpdateInput(serviceID, version, desired))
}

func (o ops) ToModel(api *fastly.Header) NestedModel {
	return FlattenToNestedModel(api)
}

func BuildCreateInput(serviceID string, version int, desired NestedModel) *fastly.CreateHeaderInput {
	name := service.StringValue(desired.Name)
	destination := service.StringValue(desired.Destination)
	cacheCondition := service.StringValue(desired.CacheCondition)
	regex := service.StringValue(desired.Regex)
	requestCondition := service.StringValue(desired.RequestCondition)
	responseCondition := service.StringValue(desired.ResponseCondition)
	source := service.StringValue(desired.Source)
	substitution := service.StringValue(desired.Substitution)
	priority := int(service.Int64Value(desired.Priority))
	ignoreIfSet := fastly.Compatibool(service.BoolValue(desired.IgnoreIfSet))

	return &fastly.CreateHeaderInput{
		ServiceID:         serviceID,
		ServiceVersion:    version,
		Name:              &name,
		Action:            headerActionPointer(desired.Action),
		Type:              headerTypePointer(desired.Type),
		Destination:       &destination,
		CacheCondition:    &cacheCondition,
		IgnoreIfSet:       &ignoreIfSet,
		Priority:          &priority,
		Regex:             &regex,
		RequestCondition:  &requestCondition,
		ResponseCondition: &responseCondition,
		Source:            &source,
		Substitution:      &substitution,
	}
}

func BuildUpdateInput(serviceID string, version int, desired NestedModel) *fastly.UpdateHeaderInput {
	destination := service.StringValue(desired.Destination)
	cacheCondition := service.StringValue(desired.CacheCondition)
	regex := service.StringValue(desired.Regex)
	requestCondition := service.StringValue(desired.RequestCondition)
	responseCondition := service.StringValue(desired.ResponseCondition)
	source := service.StringValue(desired.Source)
	substitution := service.StringValue(desired.Substitution)
	priority := int(service.Int64Value(desired.Priority))
	ignoreIfSet := fastly.Compatibool(service.BoolValue(desired.IgnoreIfSet))

	return &fastly.UpdateHeaderInput{
		ServiceID:         serviceID,
		ServiceVersion:    version,
		Name:              service.StringValue(desired.Name),
		Action:            headerActionPointer(desired.Action),
		Type:              headerTypePointer(desired.Type),
		Destination:       &destination,
		CacheCondition:    &cacheCondition,
		IgnoreIfSet:       &ignoreIfSet,
		Priority:          &priority,
		Regex:             &regex,
		RequestCondition:  &requestCondition,
		ResponseCondition: &responseCondition,
		Source:            &source,
		Substitution:      &substitution,
	}
}

func FlattenToNestedModel(api *fastly.Header) NestedModel {
	return NestedModel{
		Name:              types.StringValue(fastly.ToValue(api.Name)),
		Action:            types.StringValue(string(fastly.ToValue(api.Action))),
		Type:              types.StringValue(string(fastly.ToValue(api.Type))),
		Destination:       types.StringValue(fastly.ToValue(api.Destination)),
		CacheCondition:    types.StringValue(fastly.ToValue(api.CacheCondition)),
		IgnoreIfSet:       types.BoolValue(fastly.ToValue(api.IgnoreIfSet)),
		Priority:          types.Int64Value(int64(fastly.ToValue(api.Priority))),
		Regex:             types.StringValue(fastly.ToValue(api.Regex)),
		RequestCondition:  types.StringValue(fastly.ToValue(api.RequestCondition)),
		ResponseCondition: types.StringValue(fastly.ToValue(api.ResponseCondition)),
		Source:            types.StringValue(fastly.ToValue(api.Source)),
		Substitution:      types.StringValue(fastly.ToValue(api.Substitution)),
	}
}

// headerActionPointer looks up the API HeaderAction for a validated config value. desired.Action
// is always one of the schema's stringvalidator.OneOf values by the time Create/Update run, so
// this cast cannot produce an invalid enum value.
func headerActionPointer(v types.String) *fastly.HeaderAction {
	action := fastly.HeaderAction(service.StringValue(v))
	return &action
}

// headerTypePointer looks up the API HeaderType for a validated config value. desired.Type is
// always one of the schema's stringvalidator.OneOf values by the time Create/Update run, so this
// cast cannot produce an invalid enum value.
func headerTypePointer(v types.String) *fastly.HeaderType {
	t := fastly.HeaderType(service.StringValue(v))
	return &t
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Header]{
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
