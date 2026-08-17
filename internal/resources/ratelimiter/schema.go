package ratelimiter

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/resources/dictionary"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const DefaultFeatureRevision = 1

// uppercaseRe matches HTTP methods that are entirely uppercase (e.g. POST, PUT), matching the
// legacy provider's http_methods validation.
var uppercaseRe = regexp.MustCompile(`^[A-Z]+$`)

// caseInsensitiveState preserves the prior state value when the configured value is
// case-insensitively equal to it. actionPointer/loggerTypePointer always lowercase the value
// sent to the API, and ToModel reads state back in that lowercase form, so without this a
// differently-cased config value (e.g. "RESPONSE") would never converge with state and
// Terraform would show a persistent plan diff on every run.
type caseInsensitiveState struct{}

func (m caseInsensitiveState) Description(_ context.Context) string {
	return "Preserves the prior state value when the configured value differs only in case."
}

func (m caseInsensitiveState) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m caseInsensitiveState) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.EqualFold(req.StateValue.ValueString(), req.ConfigValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

type NestedModel struct {
	Name               types.String `tfsdk:"name"`
	Action             types.String `tfsdk:"action"`
	ClientKey          types.List   `tfsdk:"client_key"`
	FeatureRevision    types.Int64  `tfsdk:"feature_revision"`
	HTTPMethods        types.List   `tfsdk:"http_methods"`
	LoggerType         types.String `tfsdk:"logger_type"`
	PenaltyBoxDuration types.Int64  `tfsdk:"penalty_box_duration"`
	RateLimiterID      types.String `tfsdk:"rate_limiter_id"`
	Response           types.Object `tfsdk:"response"`
	ResponseObjectName types.String `tfsdk:"response_object_name"`
	RpsLimit           types.Int64  `tfsdk:"rps_limit"`
	URIDictionaryName  types.String `tfsdk:"uri_dictionary_name"`
	WindowSize         types.Int64  `tfsdk:"window_size"`
}

// ModelsEqual excludes RateLimiterID, which is API-assigned and not part of
// the desired configuration - see ops.Equal.
func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		strings.EqualFold(service.StringValue(n.Action), service.StringValue(other.Action)) &&
		n.ClientKey.Equal(other.ClientKey) &&
		service.Int64Value(n.FeatureRevision) == service.Int64Value(other.FeatureRevision) &&
		n.HTTPMethods.Equal(other.HTTPMethods) &&
		strings.EqualFold(service.StringValue(n.LoggerType), service.StringValue(other.LoggerType)) &&
		service.Int64Value(n.PenaltyBoxDuration) == service.Int64Value(other.PenaltyBoxDuration) &&
		n.Response.Equal(other.Response) &&
		service.StringValue(n.ResponseObjectName) == service.StringValue(other.ResponseObjectName) &&
		service.Int64Value(n.RpsLimit) == service.Int64Value(other.RpsLimit) &&
		service.StringValue(n.URIDictionaryName) == service.StringValue(other.URIDictionaryName) &&
		service.Int64Value(n.WindowSize) == service.Int64Value(other.WindowSize)
}

var responseAttributeTypes = map[string]attr.Type{
	"content":      types.StringType,
	"content_type": types.StringType,
	"status":       types.Int64Type,
}

func NewResponseObject(content, contentType types.String, status types.Int64) types.Object {
	return types.ObjectValueMust(
		responseAttributeTypes,
		map[string]attr.Value{
			"content":      content,
			"content_type": contentType,
			"status":       status,
		},
	)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A unique human readable name for the rate limiting rule.",
		},
		"action": schema.StringAttribute{
			Required:    true,
			Description: "The action to take when a rate limiter violation is detected. One of `log_only`, `response`, or `response_object`.",
			Validators: []validator.String{
				stringvalidator.OneOfCaseInsensitive("log_only", "response", "response_object"),
			},
			PlanModifiers: []planmodifier.String{
				caseInsensitiveState{},
			},
		},
		"client_key": schema.ListAttribute{
			ElementType: types.StringType,
			Required:    true,
			Description: "VCL variables used to generate a counter key to identify a client. Example: `[\"req.http.Fastly-Client-IP\"]`.",
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
		"feature_revision": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultFeatureRevision),
			Description: "Revision number of the rate limiting feature implementation. Defaults to the most recent revision.",
		},
		"http_methods": schema.ListAttribute{
			ElementType: types.StringType,
			Required:    true,
			Description: "HTTP methods to apply rate limiting to. Each method must be uppercase. Example: `[\"POST\", \"PUT\", \"PATCH\", \"DELETE\"]`.",
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(
					stringvalidator.RegexMatches(uppercaseRe, "each HTTP method must be uppercase (e.g. POST, PUT, PATCH, DELETE)"),
				),
			},
		},
		"logger_type": schema.StringAttribute{
			Optional:    true,
			Description: "Name of the type of logging endpoint to be used when `action` is `log_only`. One of `azureblob`, `bigquery`, `cloudfiles`, `datadog`, `digitalocean`, `elasticsearch`, `ftp`, `gcs`, `googleanalytics`, `heroku`, `honeycomb`, `http`, `https`, `kafka`, `kinesis`, `logentries`, `loggly`, `logshuttle`, `newrelic`, `openstack`, `papertrail`, `pubsub`, `s3`, `scalyr`, `sftp`, `splunk`, `stackdriver`, `sumologic`, `syslog`.",
			Validators: []validator.String{
				stringvalidator.OneOfCaseInsensitive(
					"azureblob", "bigquery", "cloudfiles", "datadog", "digitalocean", "elasticsearch",
					"ftp", "gcs", "googleanalytics", "heroku", "honeycomb", "http", "https", "kafka",
					"kinesis", "logentries", "loggly", "logshuttle", "newrelic", "openstack", "papertrail",
					"pubsub", "s3", "scalyr", "sftp", "splunk", "stackdriver", "sumologic", "syslog",
				),
			},
			PlanModifiers: []planmodifier.String{
				caseInsensitiveState{},
			},
		},
		"penalty_box_duration": schema.Int64Attribute{
			Required:    true,
			Description: "Length of time in minutes that the rate limiter is in effect after the initial violation is detected.",
			Validators: []validator.Int64{
				int64validator.Between(1, 60),
			},
		},
		"rate_limiter_id": schema.StringAttribute{
			Computed:    true,
			Description: "Alphanumeric string identifying the rate limiter.",
		},
		"response": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "Custom response to be sent when the rate limit is exceeded. Required if `action` is `response`.",
			Attributes: map[string]schema.Attribute{
				"content": schema.StringAttribute{
					Required:    true,
					Description: "HTTP response body data.",
				},
				"content_type": schema.StringAttribute{
					Required:    true,
					Description: "HTTP Content-Type (e.g. `application/json`).",
				},
				"status": schema.Int64Attribute{
					Required:    true,
					Description: "HTTP response status code (e.g. `429`).",
				},
			},
		},
		"response_object_name": schema.StringAttribute{
			Optional:    true,
			Description: "Name of existing response object. Required if `action` is `response_object`.",
		},
		"rps_limit": schema.Int64Attribute{
			Required:    true,
			Description: "Upper limit of requests per second allowed by the rate limiter.",
			Validators: []validator.Int64{
				int64validator.Between(10, 10000),
			},
		},
		"uri_dictionary_name": schema.StringAttribute{
			Optional:    true,
			Description: "The name of an Edge Dictionary containing URIs as keys. If not defined or null, all origin URIs will be rate limited.",
		},
		"window_size": schema.Int64Attribute{
			Required:    true,
			Description: "Number of seconds during which the RPS limit must be exceeded in order to trigger a violation. One of `1`, `10`, `60`.",
			Validators: []validator.Int64{
				int64validator.OneOf(1, 10, 60),
			},
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Rate limiters attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

// ops holds the remote ERLs by name produced by the most recent List call within a single
// reconcile run, since cloning a service version assigns new rate limiter IDs and the Fastly
// API only accepts an ID (never name+service+version) for update/delete. reconcile.Run always
// calls List exactly once before any Delete/Update, so Delete/Update can resolve the current ID
// (and, for Update, the currently persisted uri_dictionary_name/response_object_name) from that
// same call instead of re-listing. A fresh ops must be used per Reconcile/ReadForVersion call -
// this cache must not be shared across calls for different services/versions.
type ops struct {
	remoteByName map[string]*fastly.ERL
}

func (o *ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.ERL, error) {
	erls, err := client.ListERLs(ctx, &fastly.ListERLsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return nil, err
	}

	o.remoteByName = make(map[string]*fastly.ERL, len(erls))
	for _, e := range erls {
		o.remoteByName[fastly.ToValue(e.Name)] = e
	}

	return erls, nil
}

func (o ops) GetName(api *fastly.ERL) string {
	return fastly.ToValue(api.Name)
}

func (o *ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	remote := o.remoteByName[name]
	if remote == nil {
		return nil
	}
	return client.DeleteERL(ctx, &fastly.DeleteERLInput{ERLID: fastly.ToValue(remote.RateLimiterID)})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.ERL, error) {
	return client.CreateERL(ctx, &fastly.CreateERLInput{
		ServiceID:          serviceID,
		ServiceVersion:     version,
		Name:               new(service.StringValue(desired.Name)),
		Action:             actionPointer(desired.Action),
		ClientKey:          listToStringSlice(desired.ClientKey),
		FeatureRevision:    new(int(service.Int64Value(desired.FeatureRevision))),
		HTTPMethods:        listToStringSlice(desired.HTTPMethods),
		LoggerType:         loggerTypePointer(desired.LoggerType),
		PenaltyBoxDuration: new(int(service.Int64Value(desired.PenaltyBoxDuration))),
		Response:           responseType(desired.Response),
		ResponseObjectName: optionalStringPointer(desired.ResponseObjectName),
		RpsLimit:           new(int(service.Int64Value(desired.RpsLimit))),
		URIDictionaryName:  optionalStringPointer(desired.URIDictionaryName),
		WindowSize:         windowSizePointer(desired.WindowSize),
	})
}

func (o ops) Equal(desired NestedModel, remote *fastly.ERL) bool {
	return desired.ModelsEqual(o.ToModel(remote))
}

func (o *ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.ERL, error) {
	name := service.StringValue(desired.Name)
	remote := o.remoteByName[name]

	// uri_dictionary_name/response_object_name/response can't be cleared via UpdateERL - see
	// needsRecreate - so clearing any of them goes through a delete+create instead.
	if needsRecreate(desired, remote) {
		if err := o.Delete(ctx, client, serviceID, version, name); err != nil {
			return nil, err
		}
		return o.Create(ctx, client, serviceID, version, desired)
	}

	var id string
	if remote != nil {
		id = fastly.ToValue(remote.RateLimiterID)
	}

	return client.UpdateERL(ctx, &fastly.UpdateERLInput{
		ERLID:              id,
		Name:               new(name),
		Action:             actionPointer(desired.Action),
		ClientKey:          listToStringSlice(desired.ClientKey),
		FeatureRevision:    new(int(service.Int64Value(desired.FeatureRevision))),
		HTTPMethods:        listToStringSlice(desired.HTTPMethods),
		LoggerType:         loggerTypePointer(desired.LoggerType),
		PenaltyBoxDuration: new(int(service.Int64Value(desired.PenaltyBoxDuration))),
		Response:           responseType(desired.Response),
		ResponseObjectName: optionalStringPointer(desired.ResponseObjectName),
		RpsLimit:           new(int(service.Int64Value(desired.RpsLimit))),
		URIDictionaryName:  optionalStringPointer(desired.URIDictionaryName),
		WindowSize:         windowSizePointer(desired.WindowSize),
	})
}

// needsRecreate reports whether applying desired requires deleting and recreating the rate
// limiter rather than updating it in place. optionalStringPointer/responseType omit
// uri_dictionary_name/response_object_name/response entirely when desired clears them, since
// the API rejects an explicit empty value for any of the three - but omitting them on update
// just leaves the previously configured value in place, silently diverging from a plan that
// shows the field cleared (see https://github.com/fastly/terraform-provider-fastly/pull/1408).
// Recreating is the only way to actually clear them, mirroring account_name's handling in
// loggingbigquery.
func needsRecreate(desired NestedModel, remote *fastly.ERL) bool {
	if remote == nil {
		return false
	}

	clearsURIDictionaryName := service.StringValue(desired.URIDictionaryName) == "" && fastly.ToValue(remote.URIDictionaryName) != ""
	clearsResponseObjectName := service.StringValue(desired.ResponseObjectName) == "" && fastly.ToValue(remote.ResponseObjectName) != ""

	// Matches ToModel's criteria for a set response: all three sub-fields present.
	remoteHasResponse := remote.Response != nil && remote.Response.ERLContent != nil &&
		remote.Response.ERLContentType != nil && remote.Response.ERLStatus != nil
	clearsResponse := desired.Response.IsNull() && remoteHasResponse

	return clearsURIDictionaryName || clearsResponseObjectName || clearsResponse
}

func (o ops) ToModel(api *fastly.ERL) NestedModel {
	model := NestedModel{
		Name:               types.StringValue(fastly.ToValue(api.Name)),
		ClientKey:          stringSliceToList(api.ClientKey),
		FeatureRevision:    types.Int64Value(int64(fastly.ToValue(api.FeatureRevision))),
		HTTPMethods:        stringSliceToList(api.HTTPMethods),
		PenaltyBoxDuration: types.Int64Value(int64(fastly.ToValue(api.PenaltyBoxDuration))),
		RateLimiterID:      types.StringValue(fastly.ToValue(api.RateLimiterID)),
		RpsLimit:           types.Int64Value(int64(fastly.ToValue(api.RpsLimit))),
		WindowSize:         types.Int64Value(int64(fastly.ToValue(api.WindowSize))),
	}

	if api.Action != nil && *api.Action != "" {
		model.Action = types.StringValue(string(*api.Action))
	} else {
		model.Action = types.StringNull()
	}
	if api.LoggerType != nil && *api.LoggerType != "" {
		model.LoggerType = types.StringValue(string(*api.LoggerType))
	} else {
		model.LoggerType = types.StringNull()
	}
	if api.ResponseObjectName != nil && *api.ResponseObjectName != "" {
		model.ResponseObjectName = types.StringValue(*api.ResponseObjectName)
	} else {
		model.ResponseObjectName = types.StringNull()
	}
	if api.URIDictionaryName != nil && *api.URIDictionaryName != "" {
		model.URIDictionaryName = types.StringValue(*api.URIDictionaryName)
	} else {
		model.URIDictionaryName = types.StringNull()
	}
	if api.Response != nil && api.Response.ERLContent != nil && api.Response.ERLContentType != nil && api.Response.ERLStatus != nil {
		model.Response = NewResponseObject(
			types.StringValue(*api.Response.ERLContent),
			types.StringValue(*api.Response.ERLContentType),
			types.Int64Value(int64(*api.Response.ERLStatus)),
		)
	} else {
		model.Response = types.ObjectNull(responseAttributeTypes)
	}

	return model
}

// actionPointer lowercases the configured action, since the schema validates it
// case-insensitively but the Fastly API expects the lowercase enum value.
func actionPointer(v types.String) *fastly.ERLAction {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	action := fastly.ERLAction(strings.ToLower(v.ValueString()))
	return &action
}

// optionalStringPointer returns nil for a null/unknown/empty value, so Create/Update omit the
// field entirely rather than sending an explicit empty string. response_object_name and
// uri_dictionary_name are both optional pass-through fields to Fastly's rate limiter VCL
// generation, which produces invalid VCL (e.g. an empty table.contains argument) if sent as ""
// rather than omitted.
func optionalStringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	value := v.ValueString()
	return &value
}

// loggerTypePointer lowercases the configured logger type, since the schema validates it
// case-insensitively but the Fastly API expects the lowercase enum value.
func loggerTypePointer(v types.String) *fastly.ERLLogger {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	logger := fastly.ERLLogger(strings.ToLower(v.ValueString()))
	return &logger
}

func windowSizePointer(v types.Int64) *fastly.ERLWindowSize {
	windowSize := fastly.ERLWindowSize(service.Int64Value(v))
	return &windowSize
}

// responseType converts the optional response object into the wire format the Fastly API
// expects. Returns nil when response is not configured, which is valid when action is
// log_only or response_object (response is only required when action is response).
func responseType(obj types.Object) *fastly.ERLResponseType {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	content, ok := obj.Attributes()["content"].(types.String)
	if !ok {
		return nil
	}
	contentType, ok := obj.Attributes()["content_type"].(types.String)
	if !ok {
		return nil
	}
	status, ok := obj.Attributes()["status"].(types.Int64)
	if !ok {
		return nil
	}

	c := content.ValueString()
	ct := contentType.ValueString()
	s := int(status.ValueInt64())
	return &fastly.ERLResponseType{
		ERLContent:     &c,
		ERLContentType: &ct,
		ERLStatus:      &s,
	}
}

// stringSliceToList converts the Fastly API's []*string into a Terraform list, mapping an
// empty/nil slice to null so config that omits the attribute doesn't drift against state.
func stringSliceToList(s []*string) types.List {
	if len(s) == 0 {
		return types.ListNull(types.StringType)
	}

	elems := make([]attr.Value, 0, len(s))
	for _, v := range s {
		elems = append(elems, types.StringValue(fastly.ToValue(v)))
	}
	return types.ListValueMust(types.StringType, elems)
}

// listToStringSlice converts a Terraform list of strings into the []string pointer the Fastly
// API expects. Both client_key and http_methods are required attributes, so this is only ever
// called with a known, non-null list.
func listToStringSlice(l types.List) *[]string {
	elems := l.Elements()
	parts := make([]string, 0, len(elems))
	for _, e := range elems {
		if s, ok := e.(types.String); ok {
			parts = append(parts, s.ValueString())
		}
	}
	return &parts
}

// newReconciler builds a fresh Resource (and backing ops cache) per call, since ops.idsByName
// is populated per List call and must not be shared across concurrent reconciles of different
// services/versions.
func newReconciler() *reconcile.Resource[NestedModel, fastly.ERL] {
	return &reconcile.Resource[NestedModel, fastly.ERL]{
		Ops: &ops{},
		GetName: func(m NestedModel) string {
			return service.StringValue(m.Name)
		},
		Sortable: true,
	}
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	return newReconciler().ReadForVersion(ctx, client, serviceID, version)
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	return newReconciler().Run(ctx, client, serviceID, version, desired)
}

func Equal(a, b []NestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.Name) }, NestedModel.ModelsEqual, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	return reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
}

// ValidateConfig enforces name uniqueness among a service's rate limiters and that
// action-dependent fields are set, catching misconfigurations at plan time rather than
// deferring to a failed apply. The Fastly API itself does not enforce name uniqueness, and
// reconcile.Run keys remote/desired rate limiters by name, so duplicate names would otherwise
// silently collapse to a single rate limiter instead of failing at plan time.
func ValidateConfig(rateLimiters []NestedModel) error {
	seenNames := make(map[string]struct{}, len(rateLimiters))

	for _, item := range rateLimiters {
		name := service.StringValue(item.Name)

		if !item.Name.IsUnknown() && !item.Name.IsNull() {
			if _, ok := seenNames[name]; ok {
				return fmt.Errorf("multiple rate limiters with the same name %q; names must be unique within a service version", name)
			}
			seenNames[name] = struct{}{}
		}

		if item.Action.IsUnknown() || item.Action.IsNull() {
			continue
		}

		switch strings.ToLower(item.Action.ValueString()) {
		case "response":
			if !item.Response.IsUnknown() && item.Response.IsNull() {
				return fmt.Errorf("rate limiter %q: response is required when action is \"response\"", name)
			}
		case "response_object":
			if !item.ResponseObjectName.IsUnknown() && (item.ResponseObjectName.IsNull() || item.ResponseObjectName.ValueString() == "") {
				return fmt.Errorf("rate limiter %q: response_object_name is required when action is \"response_object\"", name)
			}
		}
	}

	return nil
}

// ValidateDictionaryReferences confirms every configured uri_dictionary_name matches a
// dictionary present in the same service config, catching at plan time the case where a
// dictionary block is renamed or removed but a rate limiter's reference to it is left stale.
// Left unvalidated, that reaches the Fastly API as a version-validation failure instead - the
// generated VCL for the unchanged rate limiter still names the now-deleted dictionary's table,
// and reconcile.Run only reconciles a rate limiter whose own desired fields changed, so removing
// just the dictionary block never updates it.
func ValidateDictionaryReferences(rateLimiters []NestedModel, dictionaries []dictionary.NestedModel) error {
	dictionaryNames := make(map[string]struct{}, len(dictionaries))
	for _, d := range dictionaries {
		if d.Name.IsUnknown() || d.Name.IsNull() {
			continue
		}
		dictionaryNames[service.StringValue(d.Name)] = struct{}{}
	}

	for _, rl := range rateLimiters {
		if rl.URIDictionaryName.IsUnknown() || rl.URIDictionaryName.IsNull() {
			continue
		}

		name := service.StringValue(rl.URIDictionaryName)
		if name == "" {
			continue
		}

		if _, ok := dictionaryNames[name]; !ok {
			return fmt.Errorf("rate limiter %q: uri_dictionary_name %q does not match any configured dictionary", service.StringValue(rl.Name), name)
		}
	}

	return nil
}
