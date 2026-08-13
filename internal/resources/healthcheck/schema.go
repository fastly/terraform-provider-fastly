package healthcheck

import (
	"context"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultCheckInterval    = 5000
	DefaultExpectedResponse = 200
	DefaultHTTPVersion      = "1.1"
	DefaultInitial          = 3
	DefaultMethod           = "HEAD"
	DefaultThreshold        = 3
	DefaultTimeout          = 5000
	DefaultWindow           = 5
)

type NestedModel struct {
	Name             types.String `tfsdk:"name"`
	Host             types.String `tfsdk:"host"`
	Path             types.String `tfsdk:"path"`
	CheckInterval    types.Int64  `tfsdk:"check_interval"`
	ExpectedResponse types.Int64  `tfsdk:"expected_response"`
	Headers          types.Set    `tfsdk:"headers"`
	HTTPVersion      types.String `tfsdk:"http_version"`
	Initial          types.Int64  `tfsdk:"initial"`
	Method           types.String `tfsdk:"method"`
	Threshold        types.Int64  `tfsdk:"threshold"`
	Timeout          types.Int64  `tfsdk:"timeout"`
	Window           types.Int64  `tfsdk:"window"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.Host) == service.StringValue(other.Host) &&
		service.StringValue(n.Path) == service.StringValue(other.Path) &&
		service.Int64Value(n.CheckInterval) == service.Int64Value(other.CheckInterval) &&
		service.Int64Value(n.ExpectedResponse) == service.Int64Value(other.ExpectedResponse) &&
		headersEqual(n.Headers, other.Headers) &&
		service.StringValue(n.HTTPVersion) == service.StringValue(other.HTTPVersion) &&
		service.Int64Value(n.Initial) == service.Int64Value(other.Initial) &&
		service.StringValue(n.Method) == service.StringValue(other.Method) &&
		service.Int64Value(n.Threshold) == service.Int64Value(other.Threshold) &&
		service.Int64Value(n.Timeout) == service.Int64Value(other.Timeout) &&
		service.Int64Value(n.Window) == service.Int64Value(other.Window)
}

// headersEqual treats a null, unknown, or empty set as equivalent to any other empty
// representation - including a zero-value types.Set, which a bare NestedModel{} literal
// (as used in tests and reconcile.MatchOrder's name-only lookups) produces, and which
// types.Set.Equal does not otherwise consider equal to itself across two independently
// zero-valued instances.
func headersEqual(a, b types.Set) bool {
	if headersUnset(a) && headersUnset(b) {
		return true
	}
	return a.Equal(b)
}

func headersUnset(s types.Set) bool {
	return s.IsNull() || s.IsUnknown() || len(s.Elements()) == 0
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A unique name to identify this health check. Changing this attribute will delete and recreate the resource.",
		},
		"host": schema.StringAttribute{
			Required:    true,
			Description: "The host to check.",
		},
		"path": schema.StringAttribute{
			Required:    true,
			Description: "The path to check.",
		},
		"check_interval": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultCheckInterval),
			Description: "How often to run the health check in milliseconds. Default `5000`.",
		},
		"expected_response": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultExpectedResponse),
			Description: "The status code expected from the host. Default `200`.",
		},
		"headers": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Custom health check HTTP headers (e.g. if your health check requires an API key to be provided).",
		},
		"http_version": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultHTTPVersion),
			Description: "Whether to use version `1.0` or `1.1` HTTP. Default `1.1`.",
		},
		"initial": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultInitial),
			Description: "When loading a config, the initial number of probes to be seen as OK. Default `3`.",
		},
		"method": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultMethod),
			Description: "Which HTTP method to use. Default `HEAD`.",
		},
		"threshold": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultThreshold),
			Description: "How many health checks must succeed to be considered healthy. Default `3`.",
		},
		"timeout": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultTimeout),
			Description: "Timeout in milliseconds. Default `5000`.",
		},
		"window": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultWindow),
			Description: "The number of most recent health check queries to keep for this health check. Default `5`.",
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Health checks attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.HealthCheck, error) {
	return client.ListHealthChecks(ctx, &fastly.ListHealthChecksInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.HealthCheck) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteHealthCheck(ctx, &fastly.DeleteHealthCheckInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.HealthCheck, error) {
	return client.CreateHealthCheck(ctx, &fastly.CreateHealthCheckInput{
		ServiceID:        serviceID,
		ServiceVersion:   version,
		Name:             new(service.StringValue(desired.Name)),
		Host:             new(service.StringValue(desired.Host)),
		Path:             new(service.StringValue(desired.Path)),
		CheckInterval:    new(int(service.Int64Value(desired.CheckInterval))),
		ExpectedResponse: new(int(service.Int64Value(desired.ExpectedResponse))),
		Headers:          headersToPointer(desired.Headers),
		HTTPVersion:      new(service.StringValue(desired.HTTPVersion)),
		Initial:          new(int(service.Int64Value(desired.Initial))),
		Method:           new(service.StringValue(desired.Method)),
		Threshold:        new(int(service.Int64Value(desired.Threshold))),
		Timeout:          new(int(service.Int64Value(desired.Timeout))),
		Window:           new(int(service.Int64Value(desired.Window))),
	})
}

func (o ops) Equal(desired NestedModel, remote *fastly.HealthCheck) bool {
	return desired.ModelsEqual(o.ToModel(remote))
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.HealthCheck, error) {
	return client.UpdateHealthCheck(ctx, &fastly.UpdateHealthCheckInput{
		ServiceID:        serviceID,
		ServiceVersion:   version,
		Name:             service.StringValue(desired.Name),
		Host:             new(service.StringValue(desired.Host)),
		Path:             new(service.StringValue(desired.Path)),
		CheckInterval:    new(int(service.Int64Value(desired.CheckInterval))),
		ExpectedResponse: new(int(service.Int64Value(desired.ExpectedResponse))),
		Headers:          headersToPointer(desired.Headers),
		HTTPVersion:      new(service.StringValue(desired.HTTPVersion)),
		Initial:          new(int(service.Int64Value(desired.Initial))),
		Method:           new(service.StringValue(desired.Method)),
		Threshold:        new(int(service.Int64Value(desired.Threshold))),
		Timeout:          new(int(service.Int64Value(desired.Timeout))),
		Window:           new(int(service.Int64Value(desired.Window))),
	})
}

func (o ops) ToModel(api *fastly.HealthCheck) NestedModel {
	return NestedModel{
		Name:             types.StringValue(fastly.ToValue(api.Name)),
		Host:             types.StringValue(fastly.ToValue(api.Host)),
		Path:             types.StringValue(fastly.ToValue(api.Path)),
		CheckInterval:    service.Int64PointerOrDefault(api.CheckInterval, DefaultCheckInterval),
		ExpectedResponse: service.Int64PointerOrDefault(api.ExpectedResponse, DefaultExpectedResponse),
		Headers:          headersFromSlice(api.Headers),
		HTTPVersion:      service.StringPointerOrDefault(api.HTTPVersion, DefaultHTTPVersion),
		Initial:          service.Int64PointerOrDefault(api.Initial, DefaultInitial),
		Method:           service.StringPointerOrDefault(api.Method, DefaultMethod),
		Threshold:        service.Int64PointerOrDefault(api.Threshold, DefaultThreshold),
		Timeout:          service.Int64PointerOrDefault(api.Timeout, DefaultTimeout),
		Window:           service.Int64PointerOrDefault(api.Window, DefaultWindow),
	}
}

// headersToPointer converts headers into the *[]string shape CreateHealthCheckInput and
// UpdateHealthCheckInput expect, returning nil (omitting the field entirely) when no headers
// are configured rather than sending an empty array.
func headersToPointer(s types.Set) *[]string {
	headers := headersToSlice(s)
	if headers == nil {
		return nil
	}
	return &headers
}

func headersToSlice(s types.Set) []string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}

	elems := s.Elements()
	if len(elems) == 0 {
		return nil
	}

	headers := make([]string, 0, len(elems))
	for _, e := range elems {
		if v, ok := e.(types.String); ok {
			headers = append(headers, v.ValueString())
		}
	}
	return headers
}

// headersFromSlice converts the API's headers slice into a Set rather than a List: the Fastly
// API always returns headers pre-sorted, so a List would show a perpetual diff against
// differently-ordered configuration (see the equivalent TypeSet usage in the old SDKv2
// provider's block_fastly_service_healthcheck.go).
func headersFromSlice(headers []string) types.Set {
	if len(headers) == 0 {
		return types.SetNull(types.StringType)
	}

	elems := make([]attr.Value, len(headers))
	for i, h := range headers {
		elems[i] = types.StringValue(h)
	}
	return types.SetValueMust(types.StringType, elems)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.HealthCheck]{
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
