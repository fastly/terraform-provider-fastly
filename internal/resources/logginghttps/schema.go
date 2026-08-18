package logginghttps

import (
	"context"
	"maps"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultFormatVersion     = 2
	DefaultResponseCondition = ""
	DefaultProcessingRegion  = "none"
	// DefaultGzipLevel is a sentinel meaning "gzip_level not configured". A real
	// value is 0-9, so -1 lets the provider distinguish an unset level from an
	// explicit 0 (valid "no compression"). An unset level is never written,
	// because the API rejects requests that set both compression_codec and
	// gzip_level, and it auto-manages the level otherwise (e.g. 3 for gzip).
	DefaultGzipLevel         = -1
	DefaultCompressionCodec  = ""
	DefaultContentType       = ""
	DefaultHeaderName        = ""
	DefaultHeaderValue       = ""
	DefaultJSONFormat        = "0"
	DefaultMessageType       = "blank"
	DefaultMethod            = "POST"
	DefaultPeriod            = 5
	DefaultRequestMaxBytes   = 0
	DefaultRequestMaxEntries = 0
	DefaultTLSHostname       = ""

	// maximumFormatLength is the maximum length the Fastly API accepts for a
	// logging endpoint `format` string. Exceeding it is only rejected by the
	// API at apply time, so it is enforced at plan/validate time instead.
	maximumFormatLength = 12288
)

// commonModel holds the HTTPS logging attributes shared by VCL and Compute
// services. format, format_version, placement, and response_condition only
// affect generated VCL, so they live on NestedModel only — Compute services use
// ComputeNestedModel, which embeds just this common set.
type commonModel struct {
	Name              types.String `tfsdk:"name"`
	URL               types.String `tfsdk:"url"`
	ContentType       types.String `tfsdk:"content_type"`
	CompressionCodec  types.String `tfsdk:"compression_codec"`
	GzipLevel         types.Int64  `tfsdk:"gzip_level"`
	HeaderName        types.String `tfsdk:"header_name"`
	HeaderValue       types.String `tfsdk:"header_value"`
	JSONFormat        types.String `tfsdk:"json_format"`
	MessageType       types.String `tfsdk:"message_type"`
	Method            types.String `tfsdk:"method"`
	Period            types.Int64  `tfsdk:"period"`
	ProcessingRegion  types.String `tfsdk:"processing_region"`
	RequestMaxBytes   types.Int64  `tfsdk:"request_max_bytes"`
	RequestMaxEntries types.Int64  `tfsdk:"request_max_entries"`
	TLS               types.Object `tfsdk:"tls"`
}

// NestedModel is the HTTPS logging model for the standalone
// fastly_service_logging_https resource and the VCL nested block
// (service_cdn_auto.logging_https).
type NestedModel struct {
	commonModel
	Format            types.String `tfsdk:"format"`
	FormatVersion     types.Int64  `tfsdk:"format_version"`
	Placement         types.String `tfsdk:"placement"`
	ResponseCondition types.String `tfsdk:"response_condition"`
}

// ComputeNestedModel is the HTTPS logging model for the Compute nested block
// (service_compute_auto.logging_https). It omits format, format_version,
// placement, and response_condition, which only apply to VCL services.
type ComputeNestedModel struct {
	commonModel
}

var tlsAttributeTypes = map[string]attr.Type{
	"ca_cert":     types.StringType,
	"client_cert": types.StringType,
	"client_key":  types.StringType,
	"hostname":    types.StringType,
}

func NewTLSObject(caCert, clientCert, clientKey, hostname types.String) types.Object {
	return types.ObjectValueMust(
		tlsAttributeTypes,
		map[string]attr.Value{
			"ca_cert":     caCert,
			"client_cert": clientCert,
			"client_key":  clientKey,
			"hostname":    hostname,
		},
	)
}

// defaultTLSObject is the `tls` block's value when the practitioner omits the
// whole block. Unlike loggingsplunk's tls block, none of these fields have an
// environment variable in the live (SDKv2) provider, so a single static
// default suffices — there is no need for a custom defaults.Object
// implementation that resolves each field independently.
var defaultTLSObject = NewTLSObject(
	types.StringValue(""),
	types.StringValue(""),
	types.StringValue(""),
	types.StringValue(DefaultTLSHostname),
)

func objectValue(obj types.Object, name string) types.String {
	if obj.IsNull() || obj.IsUnknown() {
		return types.StringValue("")
	}
	value, ok := obj.Attributes()[name]
	if !ok || value == nil || value.IsNull() || value.IsUnknown() {
		return types.StringValue("")
	}
	stringValue, ok := value.(types.String)
	if !ok {
		return types.StringValue("")
	}
	return stringValue
}

func (n commonModel) TLSCACert() types.String {
	return objectValue(n.TLS, "ca_cert")
}

func (n commonModel) TLSClientCert() types.String {
	return objectValue(n.TLS, "client_cert")
}

func (n commonModel) TLSClientKey() types.String {
	return objectValue(n.TLS, "client_key")
}

func (n commonModel) TLSHostname() types.String {
	return objectValue(n.TLS, "hostname")
}

func (n commonModel) equal(other commonModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.URL) == service.StringValue(other.URL) &&
		service.StringValue(n.ContentType) == service.StringValue(other.ContentType) &&
		service.StringValue(n.CompressionCodec) == service.StringValue(other.CompressionCodec) &&
		service.Int64Value(n.GzipLevel) == service.Int64Value(other.GzipLevel) &&
		service.StringValue(n.HeaderName) == service.StringValue(other.HeaderName) &&
		service.StringValue(n.HeaderValue) == service.StringValue(other.HeaderValue) &&
		service.StringValue(n.JSONFormat) == service.StringValue(other.JSONFormat) &&
		service.StringValue(n.MessageType) == service.StringValue(other.MessageType) &&
		service.StringValue(n.Method) == service.StringValue(other.Method) &&
		service.Int64Value(n.Period) == service.Int64Value(other.Period) &&
		service.StringValue(n.ProcessingRegion) == service.StringValue(other.ProcessingRegion) &&
		service.Int64Value(n.RequestMaxBytes) == service.Int64Value(other.RequestMaxBytes) &&
		service.Int64Value(n.RequestMaxEntries) == service.Int64Value(other.RequestMaxEntries) &&
		service.StringValue(n.TLSCACert()) == service.StringValue(other.TLSCACert()) &&
		service.StringValue(n.TLSClientCert()) == service.StringValue(other.TLSClientCert()) &&
		service.StringValue(n.TLSClientKey()) == service.StringValue(other.TLSClientKey()) &&
		service.StringValue(n.TLSHostname()) == service.StringValue(other.TLSHostname())
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return n.commonModel.equal(other.commonModel) &&
		service.StringValue(n.Format) == service.StringValue(other.Format) &&
		service.Int64Value(n.FormatVersion) == service.Int64Value(other.FormatVersion) &&
		service.StringValue(n.Placement) == service.StringValue(other.Placement) &&
		service.StringValue(n.ResponseCondition) == service.StringValue(other.ResponseCondition)
}

func (c ComputeNestedModel) ModelsEqual(other ComputeNestedModel) bool {
	return c.commonModel.equal(other.commonModel)
}

// CommonAttributes returns the full HTTPS logging attribute set — the shared
// attributes plus the VCL-only ones (format, format_version, placement,
// response_condition). Used by the standalone fastly_service_logging_https
// resource (which can attach to either service type) and the VCL nested block
// (NestedBlockSchema). Compute services use ComputeAttributes instead.
func CommonAttributes() map[string]schema.Attribute {
	attrs := sharedAttributes()
	maps.Copy(attrs, vclOnlyAttributes())
	return attrs
}

// ComputeAttributes returns the HTTPS logging attribute set for Compute
// services, omitting the VCL-only attributes exposed by CommonAttributes.
func ComputeAttributes() map[string]schema.Attribute {
	return sharedAttributes()
}

// sharedAttributes returns the HTTPS logging attributes common to both VCL and
// Compute services.
func sharedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		// Required
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name for the real-time logging configuration. Must be unique within the service.",
		},
		"url": schema.StringAttribute{
			Required:    true,
			Description: "URL that log data will be sent to. Must use the HTTPS protocol.",
			Validators: []validator.String{
				httpsURL{},
			},
		},
		// Optional
		"compression_codec": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultCompressionCodec),
			Validators: []validator.String{
				stringvalidator.OneOf("zstd", "snappy", "gzip"),
			},
			Description: "The codec used for compressing your logs. Valid values are `zstd`, `snappy`, and `gzip`. If the codec is `gzip`, `gzip_level` defaults to `3`; to use a different level, leave `compression_codec` unset and set `gzip_level` instead. Conflicts with `gzip_level`: setting both in the same request will result in an error.",
		},
		"content_type": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultContentType),
			Description: "Value of the `Content-Type` header sent with the request.",
		},
		"gzip_level": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultGzipLevel),
			// compression_codec and gzip_level are mutually exclusive; the API
			// rejects a request that sets both. Validation runs against config,
			// where an unset gzip_level is null rather than the -1 default, so
			// this correctly fires only when both are set. int64validator.Between
			// likewise skips a null (omitted) config value, so it only rejects an
			// explicitly configured value outside 0-9 — including the internal -1
			// sentinel itself, which a user should never configure directly (omit
			// the attribute instead to get the same "unset" behavior).
			Validators: []validator.Int64{
				gzipLevelCodecConflict{},
				int64validator.Between(0, 9),
			},
			Description: "The level of gzip encoding when sending logs. Valid values are `0` (no compression) through `9`. To compress at a specific gzip level, leave `compression_codec` unset and set this. Conflicts with `compression_codec`: setting both in the same request will result in an error.",
		},
		"header_name": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultHeaderName),
			Description: "Custom header sent with the request.",
		},
		"header_value": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultHeaderValue),
			Description: "Value of the custom header sent with the request.",
		},
		"json_format": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultJSONFormat),
			Validators: []validator.String{
				stringvalidator.OneOf("0", "1", "2"),
			},
			Description: "Enforces valid JSON formatting for log entries. Can be either disabled (`0`), array of JSON (`1`), or newline delimited JSON (`2`). Default `0`.",
		},
		"message_type": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultMessageType),
			Validators: []validator.String{
				stringvalidator.OneOf("classic", "loggly", "logplex", "blank"),
			},
			Description: "How the message should be formatted. Valid values are `classic`, `loggly`, `logplex`, and `blank`. Default `blank`.",
		},
		"method": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultMethod),
			Validators: []validator.String{
				stringvalidator.OneOf("POST", "PUT"),
			},
			Description: "HTTP method used for request. Can be either `POST` or `PUT`. Default `POST`.",
		},
		"period": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultPeriod),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			Description: "How frequently, in seconds, batches of log data are sent to the HTTPS endpoint. A value of `0` sends logs at the same interval as the default, which is `5` seconds.",
		},
		"processing_region": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultProcessingRegion),
			Validators: []validator.String{
				stringvalidator.OneOf("none", "us", "eu"),
			},
			Description: "The geographic region where the logs will be processed before streaming. Valid values are `us`, `eu`, and `none` for global. Default: `none`.",
		},
		"request_max_bytes": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultRequestMaxBytes),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			Description: "The maximum number of bytes sent in one request. Default `0` for unbounded (100MB).",
		},
		"request_max_entries": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultRequestMaxEntries),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			Description: "The maximum number of logs sent in one request. Default `0` for unbounded (10k).",
		},
		// Grouped under `tls` since client_key is credential material used to
		// authenticate this endpoint to the HTTPS server via mutual TLS, and
		// ca_cert/client_cert/hostname are used alongside it to configure that
		// same connection.
		"tls": schema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Default:     objectdefault.StaticValue(defaultTLSObject),
			Description: "TLS configuration used to authenticate the HTTPS server, and optionally this endpoint via mutual TLS.",
			Attributes: map[string]schema.Attribute{
				"ca_cert": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
					Description: "A secure certificate to authenticate the server with. Must be in PEM format.",
				},
				"client_cert": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(""),
					Description: "The client certificate used to make authenticated requests. Must be in PEM format.",
				},
				"client_key": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Sensitive:   true,
					Default:     stringdefault.StaticString(""),
					Description: "The client private key used to make authenticated requests. Must be in PEM format.",
				},
				"hostname": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(DefaultTLSHostname),
					Description: "The hostname used to verify the server's certificate. This should be one of the Subject Alternative Name (SAN) fields for the certificate. Common Names (CN) are not supported.",
				},
			},
		},
	}
}

// vclOnlyAttributes returns the HTTPS logging attributes that only affect
// generated VCL and have no meaning for Compute services.
func vclOnlyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"format": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(constants.LoggingHTTPSDefaultFormat),
			Validators: []validator.String{
				stringvalidator.LengthAtMost(maximumFormatLength),
			},
			Description: "A Fastly [log format string](https://www.fastly.com/documentation/guides/integrations/streaming-logs/custom-log-formats/).",
		},
		"format_version": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultFormatVersion),
			Validators: []validator.Int64{
				int64validator.Between(1, 2),
			},
			Description: "The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if `format_version` is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.",
		},
		"placement": schema.StringAttribute{
			// Not Computed: unset and explicitly "none" are distinct states on the
			// API (unset lets the API auto-place the logging call; "none" suppresses
			// it entirely). Keeping it plain Optional means removing placement from
			// config is a real, visible diff from Terraform core's own plan proposal,
			// with no plan modifier required to force it.
			//
			// On a VCL service the API resolves this to exactly what was configured.
			// Compute (wasm) services are the exception — the API forces "none" there
			// regardless of what was sent, which can never match the planned null, so
			// ResetVCLOnlyToDefaults discards it after apply for those services.
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf("none"),
			},
			Description: "Where in the generated VCL the logging call should be placed. If not set, endpoints with `format_version` of `2` are placed in `vcl_log` and those with `format_version` of `1` are placed in `vcl_deliver`. Valid value is `none`.",
		},
		"response_condition": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultResponseCondition),
			Description: "The name of an existing condition in the configured endpoint, or leave blank to always execute.",
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
	// For the standalone resource, service_id + name locate the endpoint in the
	// API, so a change to either cannot be an in-place update. version is not
	// replacement-forcing: the explicit clone workflow copies the endpoint into
	// the new version, so an in-place update there succeeds. Applied to name
	// here (not in CommonAttributes) so the nested block, where name is only a
	// list key, is unaffected.
	nameAttr := attrs["name"].(schema.StringAttribute)
	nameAttr.PlanModifiers = []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
	attrs["name"] = nameAttr
	return attrs
}

// NestedBlockSchema returns the HTTPS logging nested block schema for VCL
// services (service_cdn_auto.logging_https), including the VCL-only
// attributes.
func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "HTTPS logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

// ComputeNestedBlockSchema returns the HTTPS logging nested block schema for
// Compute services (service_compute_auto.logging_https), omitting the
// VCL-only attributes.
func ComputeNestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "HTTPS logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: ComputeAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.HTTPS, error) {
	return client.ListHTTPS(ctx, &fastly.ListHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.HTTPS) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteHTTPS(ctx, &fastly.DeleteHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.HTTPS, error) {
	input := BuildCreateInput(serviceID, version, desired)
	return client.CreateHTTPS(ctx, input)
}

func (o ops) Equal(desired NestedModel, remote *fastly.HTTPS) bool {
	remoteModel := FlattenToNestedModel(remote)
	preserveGzipSentinel(&remoteModel, desired)
	return desired.ModelsEqual(remoteModel)
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.HTTPS, error) {
	input := BuildUpdateInput(serviceID, version, desired)
	return client.UpdateHTTPS(ctx, input)
}

func (o ops) ToModel(api *fastly.HTTPS) NestedModel {
	return FlattenToNestedModel(api)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.HTTPS]{
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
	result := reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
	// order carries the configured/prior models, so it holds the gzip_level
	// sentinel for endpoints the user left unset; preserve it on the read-back.
	preserveGzipSentinelList(result, order)
	return result
}

type computeOps struct{}

func (o computeOps) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.HTTPS, error) {
	return client.ListHTTPS(ctx, &fastly.ListHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o computeOps) GetName(api *fastly.HTTPS) string {
	return fastly.ToValue(api.Name)
}

func (o computeOps) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteHTTPS(ctx, &fastly.DeleteHTTPSInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o computeOps) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.HTTPS, error) {
	input := BuildComputeCreateInput(serviceID, version, desired)
	return client.CreateHTTPS(ctx, input)
}

func (o computeOps) Equal(desired ComputeNestedModel, remote *fastly.HTTPS) bool {
	remoteModel := FlattenToComputeNestedModel(remote)
	preserveGzipSentinelCompute(&remoteModel, desired)
	return desired.ModelsEqual(remoteModel)
}

func (o computeOps) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.HTTPS, error) {
	input := BuildComputeUpdateInput(serviceID, version, desired)
	return client.UpdateHTTPS(ctx, input)
}

func (o computeOps) ToModel(api *fastly.HTTPS) ComputeNestedModel {
	return FlattenToComputeNestedModel(api)
}

var computeReconciler = &reconcile.Resource[ComputeNestedModel, fastly.HTTPS]{
	Ops: computeOps{},
	GetName: func(m ComputeNestedModel) string {
		return service.StringValue(m.Name)
	},
	Sortable: true,
}

func ComputeReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]ComputeNestedModel, error) {
	return computeReconciler.ReadForVersion(ctx, client, serviceID, version)
}

func ComputeReconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []ComputeNestedModel) error {
	return computeReconciler.Run(ctx, client, serviceID, version, desired)
}

func ComputeEqual(a, b []ComputeNestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m ComputeNestedModel) string { return service.StringValue(m.Name) }, ComputeNestedModel.ModelsEqual, true)
}

func ComputeMatchOrder(items, order []ComputeNestedModel) []ComputeNestedModel {
	result := reconcile.MatchOrder(items, order, func(m ComputeNestedModel) string { return service.StringValue(m.Name) })
	// order carries the configured/prior models, so it holds the gzip_level
	// sentinel for endpoints the user left unset; preserve it on the read-back.
	preserveGzipSentinelListCompute(result, order)
	return result
}
