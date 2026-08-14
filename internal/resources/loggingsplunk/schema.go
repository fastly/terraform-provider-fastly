package loggingsplunk

import (
	"context"
	"maps"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/defaults"
	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	fwdefaults "github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
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
	DefaultUseTLS            = false
	DefaultRequestMaxBytes   = 0
	DefaultRequestMaxEntries = 0
	DefaultTLSHostname       = ""

	// maximumFormatLength is the maximum length the Fastly API accepts for a
	// logging endpoint `format` string. Exceeding it is only rejected by the
	// API at apply time, so it is enforced at plan/validate time instead.
	maximumFormatLength = 12288

	splunkTokenEnvVar         = "FASTLY_SPLUNK_TOKEN"
	splunkTLSCACertEnvVar     = "FASTLY_SPLUNK_CA_CERT"
	splunkTLSClientCertEnvVar = "FASTLY_SPLUNK_CLIENT_CERT"
	splunkTLSClientKeyEnvVar  = "FASTLY_SPLUNK_CLIENT_KEY"
)

// commonModel holds the Splunk logging attributes shared by VCL and Compute
// services. format, format_version, placement, and response_condition only
// affect generated VCL, so they live on NestedModel only — Compute services use
// ComputeNestedModel, which embeds just this common set.
type commonModel struct {
	Name              types.String `tfsdk:"name"`
	URL               types.String `tfsdk:"url"`
	Authentication    types.Object `tfsdk:"authentication"`
	TLS               types.Object `tfsdk:"tls"`
	UseTLS            types.Bool   `tfsdk:"use_tls"`
	ProcessingRegion  types.String `tfsdk:"processing_region"`
	RequestMaxBytes   types.Int64  `tfsdk:"request_max_bytes"`
	RequestMaxEntries types.Int64  `tfsdk:"request_max_entries"`
}

// NestedModel is the Splunk logging model for the standalone
// fastly_service_logging_splunk resource and the VCL nested block
// (service_cdn_auto.logging_splunk).
type NestedModel struct {
	commonModel
	Format            types.String `tfsdk:"format"`
	FormatVersion     types.Int64  `tfsdk:"format_version"`
	Placement         types.String `tfsdk:"placement"`
	ResponseCondition types.String `tfsdk:"response_condition"`
}

// ComputeNestedModel is the Splunk logging model for the Compute nested block
// (service_compute_auto.logging_splunk). It omits format, format_version,
// placement, and response_condition, which only apply to VCL services.
type ComputeNestedModel struct {
	commonModel
}

var authenticationAttributeTypes = map[string]attr.Type{
	"token": types.StringType,
}

func NewAuthenticationObject(token types.String) types.Object {
	return types.ObjectValueMust(
		authenticationAttributeTypes,
		map[string]attr.Value{
			"token": token,
		},
	)
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

// authenticationEnvDefault populates the authentication object from the
// FASTLY_SPLUNK_TOKEN environment variable when the practitioner omits the
// whole `authentication` block. The framework only walks into an object
// attribute's per-field Default handlers once the object itself already
// resolves to a known value; a Computed object attribute with no Default of
// its own is instead marked wholesale unknown, and its children's Defaults
// are never evaluated. Setting this Default on the parent gives the object a
// known value up front so the per-field default still runs.
type authenticationEnvDefault struct{}

func (authenticationEnvDefault) Description(_ context.Context) string {
	return "value defaults to the FASTLY_SPLUNK_TOKEN environment variable"
}

func (d authenticationEnvDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (authenticationEnvDefault) DefaultObject(ctx context.Context, _ fwdefaults.ObjectRequest, resp *fwdefaults.ObjectResponse) {
	resp.PlanValue = NewAuthenticationObject(envStringDefault(ctx, splunkTokenEnvVar))
}

// tlsEnvDefault is authenticationEnvDefault for the `tls` object: it defaults
// ca_cert, client_cert, and client_key from their environment variables.
// hostname has no environment variable in the live (SDKv2) provider, so it
// always defaults to "".
type tlsEnvDefault struct{}

func (tlsEnvDefault) Description(_ context.Context) string {
	return "ca_cert, client_cert, and client_key default to the FASTLY_SPLUNK_CA_CERT, FASTLY_SPLUNK_CLIENT_CERT, and FASTLY_SPLUNK_CLIENT_KEY environment variables"
}

func (d tlsEnvDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (tlsEnvDefault) DefaultObject(ctx context.Context, _ fwdefaults.ObjectRequest, resp *fwdefaults.ObjectResponse) {
	resp.PlanValue = NewTLSObject(
		envStringDefault(ctx, splunkTLSCACertEnvVar),
		envStringDefault(ctx, splunkTLSClientCertEnvVar),
		envStringDefault(ctx, splunkTLSClientKeyEnvVar),
		types.StringValue(DefaultTLSHostname),
	)
}

func envStringDefault(ctx context.Context, envVar string) types.String {
	var resp fwdefaults.StringResponse
	defaults.EnvString(envVar, "").DefaultString(ctx, fwdefaults.StringRequest{}, &resp)
	return resp.PlanValue
}

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

func (n commonModel) Token() types.String {
	return objectValue(n.Authentication, "token")
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
		service.StringValue(n.Token()) == service.StringValue(other.Token()) &&
		service.StringValue(n.TLSCACert()) == service.StringValue(other.TLSCACert()) &&
		service.StringValue(n.TLSClientCert()) == service.StringValue(other.TLSClientCert()) &&
		service.StringValue(n.TLSClientKey()) == service.StringValue(other.TLSClientKey()) &&
		service.StringValue(n.TLSHostname()) == service.StringValue(other.TLSHostname()) &&
		service.BoolValue(n.UseTLS) == service.BoolValue(other.UseTLS) &&
		service.StringValue(n.ProcessingRegion) == service.StringValue(other.ProcessingRegion) &&
		service.Int64Value(n.RequestMaxBytes) == service.Int64Value(other.RequestMaxBytes) &&
		service.Int64Value(n.RequestMaxEntries) == service.Int64Value(other.RequestMaxEntries)
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

// CommonAttributes returns the full Splunk logging attribute set — the shared
// attributes plus the VCL-only ones (format, format_version, placement,
// response_condition). Used by the standalone fastly_service_logging_splunk
// resource (which can attach to either service type) and the VCL nested block
// (NestedBlockSchema). Compute services use ComputeAttributes instead.
func CommonAttributes() map[string]schema.Attribute {
	attrs := sharedAttributes()
	maps.Copy(attrs, vclOnlyAttributes())
	return attrs
}

// ComputeAttributes returns the Splunk logging attribute set for Compute
// services, omitting the VCL-only attributes exposed by CommonAttributes.
func ComputeAttributes() map[string]schema.Attribute {
	return sharedAttributes()
}

// sharedAttributes returns the Splunk logging attributes common to both VCL and
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
			Description: "The URL to post logs to.",
		},
		// Grouped under `authentication` to match the other logging endpoints, even
		// though Splunk has a single credential. Optional+Computed rather than
		// Required like loggingdatadog's: the FASTLY_SPLUNK_TOKEN environment
		// variable is existing (SDKv2 provider) behavior that must be preserved.
		// authenticationRequired enforces that the token itself stays required.
		"authentication": schema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Default:     authenticationEnvDefault{},
			Description: "Splunk authentication credentials. When this block is omitted entirely, defaults to the `FASTLY_SPLUNK_TOKEN` environment variable.",
			Validators: []validator.Object{
				authenticationRequired{},
			},
			Attributes: map[string]schema.Attribute{
				"token": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Sensitive:   true,
					Default:     defaults.EnvString(splunkTokenEnvVar, ""),
					Description: "A Splunk token for use in posting logs over HTTP to your collector. Can be set via the `FASTLY_SPLUNK_TOKEN` environment variable.",
				},
			},
		},
		// Grouped under `tls` since client_key is credential material used to
		// authenticate this endpoint to the Splunk collector via mutual TLS, and
		// ca_cert/client_cert/hostname are used alongside it to configure that same
		// connection.
		"tls": schema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Default:     tlsEnvDefault{},
			Description: "TLS configuration used when `use_tls` is enabled. When this block is omitted entirely, `ca_cert`, `client_cert`, and `client_key` default to the `FASTLY_SPLUNK_CA_CERT`, `FASTLY_SPLUNK_CLIENT_CERT`, and `FASTLY_SPLUNK_CLIENT_KEY` environment variables.",
			Attributes: map[string]schema.Attribute{
				"ca_cert": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     defaults.EnvString(splunkTLSCACertEnvVar, ""),
					Description: "A secure certificate to authenticate the server with. Must be in PEM format. Can be set via the `FASTLY_SPLUNK_CA_CERT` environment variable.",
				},
				"client_cert": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     defaults.EnvString(splunkTLSClientCertEnvVar, ""),
					Description: "The client certificate used to make authenticated requests. Must be in PEM format. Can be set via the `FASTLY_SPLUNK_CLIENT_CERT` environment variable.",
				},
				"client_key": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Sensitive:   true,
					Default:     defaults.EnvString(splunkTLSClientKeyEnvVar, ""),
					Description: "The client private key used to make authenticated requests. Must be in PEM format. Can be set via the `FASTLY_SPLUNK_CLIENT_KEY` environment variable.",
				},
				"hostname": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString(DefaultTLSHostname),
					Description: "The hostname used to verify the server's certificate. This should be one of the Subject Alternative Name (SAN) fields for the certificate. Common Names (CN) are not supported.",
				},
			},
		},
		"use_tls": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultUseTLS),
			Description: "Whether to use TLS for secure logging. Default: `false`.",
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
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultRequestMaxBytes),
			Description: "The maximum number of bytes sent in one request. Default `0` for unbounded.",
		},
		"request_max_entries": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultRequestMaxEntries),
			Description: "The maximum number of logs sent in one request. Default `0` for unbounded.",
		},
	}
}

// vclOnlyAttributes returns the Splunk logging attributes that only affect
// generated VCL and have no meaning for Compute services.
func vclOnlyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"format": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(constants.LoggingSplunkDefaultFormat),
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

// NestedBlockSchema returns the Splunk logging nested block schema for VCL
// services (service_cdn_auto.logging_splunk), including the VCL-only
// attributes.
func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Splunk logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

// ComputeNestedBlockSchema returns the Splunk logging nested block schema for
// Compute services (service_compute_auto.logging_splunk), omitting the
// VCL-only attributes.
func ComputeNestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Splunk logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: ComputeAttributes(),
		},
	}
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Splunk, error) {
	return client.ListSplunks(ctx, &fastly.ListSplunksInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.Splunk) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteSplunk(ctx, &fastly.DeleteSplunkInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Splunk, error) {
	input := BuildCreateInput(serviceID, version, desired)
	return client.CreateSplunk(ctx, input)
}

func (o ops) Equal(desired NestedModel, remote *fastly.Splunk) bool {
	return desired.ModelsEqual(FlattenToNestedModel(remote))
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Splunk, error) {
	input := BuildUpdateInput(serviceID, version, desired)
	return client.UpdateSplunk(ctx, input)
}

func (o ops) ToModel(api *fastly.Splunk) NestedModel {
	return FlattenToNestedModel(api)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Splunk]{
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

type computeOps struct{}

func (o computeOps) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Splunk, error) {
	return client.ListSplunks(ctx, &fastly.ListSplunksInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o computeOps) GetName(api *fastly.Splunk) string {
	return fastly.ToValue(api.Name)
}

func (o computeOps) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteSplunk(ctx, &fastly.DeleteSplunkInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o computeOps) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.Splunk, error) {
	input := BuildComputeCreateInput(serviceID, version, desired)
	return client.CreateSplunk(ctx, input)
}

func (o computeOps) Equal(desired ComputeNestedModel, remote *fastly.Splunk) bool {
	return desired.ModelsEqual(FlattenToComputeNestedModel(remote))
}

func (o computeOps) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.Splunk, error) {
	input := BuildComputeUpdateInput(serviceID, version, desired)
	return client.UpdateSplunk(ctx, input)
}

func (o computeOps) ToModel(api *fastly.Splunk) ComputeNestedModel {
	return FlattenToComputeNestedModel(api)
}

var computeReconciler = &reconcile.Resource[ComputeNestedModel, fastly.Splunk]{
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
	return reconcile.MatchOrder(items, order, func(m ComputeNestedModel) string { return service.StringValue(m.Name) })
}
