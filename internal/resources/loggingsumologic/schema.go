package loggingsumologic

import (
	"context"
	"maps"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultFormatVersion     = 2
	DefaultMessageType       = "blank"
	DefaultResponseCondition = ""
	DefaultProcessingRegion  = "none"

	// maximumFormatLength is the maximum length the Fastly API accepts for a
	// logging endpoint `format` string. Exceeding it is only rejected by the
	// API at apply time, so it is enforced at plan/validate time instead.
	maximumFormatLength = 12288
)

// commonModel holds the Sumo Logic logging attributes shared by VCL and
// Compute services. format, format_version, placement, and response_condition
// only affect generated VCL, so they live on NestedModel only — Compute
// services use ComputeNestedModel, which embeds just this common set.
type commonModel struct {
	Name             types.String `tfsdk:"name"`
	URL              types.String `tfsdk:"url"`
	MessageType      types.String `tfsdk:"message_type"`
	ProcessingRegion types.String `tfsdk:"processing_region"`
}

// NestedModel is the Sumo Logic logging model for the standalone
// fastly_service_logging_sumologic resource and the VCL nested block
// (service_cdn_auto.logging_sumologic).
type NestedModel struct {
	commonModel
	Format            types.String `tfsdk:"format"`
	FormatVersion     types.Int64  `tfsdk:"format_version"`
	Placement         types.String `tfsdk:"placement"`
	ResponseCondition types.String `tfsdk:"response_condition"`
}

// ComputeNestedModel is the Sumo Logic logging model for the Compute nested
// block (service_compute_auto.logging_sumologic). It omits format,
// format_version, placement, and response_condition, which only apply to VCL
// services.
type ComputeNestedModel struct {
	commonModel
}

func (n commonModel) equal(other commonModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		service.StringValue(n.URL) == service.StringValue(other.URL) &&
		service.StringValue(n.MessageType) == service.StringValue(other.MessageType) &&
		service.StringValue(n.ProcessingRegion) == service.StringValue(other.ProcessingRegion)
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

// CommonAttributes returns the full Sumo Logic logging attribute set — the
// shared attributes plus the VCL-only ones (format, format_version,
// placement, response_condition). Used by the standalone
// fastly_service_logging_sumologic resource (which can attach to either
// service type) and the VCL nested block (NestedBlockSchema). Compute
// services use ComputeAttributes instead.
func CommonAttributes() map[string]schema.Attribute {
	attrs := sharedAttributes()
	maps.Copy(attrs, vclOnlyAttributes())
	return attrs
}

// ComputeAttributes returns the Sumo Logic logging attribute set for Compute
// services, omitting the VCL-only attributes exposed by CommonAttributes.
func ComputeAttributes() map[string]schema.Attribute {
	return sharedAttributes()
}

// sharedAttributes returns the Sumo Logic logging attributes common to both
// VCL and Compute services.
func sharedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		// Required
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name for the real-time logging configuration. Must be unique within the service.",
		},
		"url": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				isURL{},
			},
			Description: "The URL to post logs to.",
		},
		// Optional
		"message_type": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(DefaultMessageType),
			Validators: []validator.String{
				stringvalidator.OneOf("classic", "loggly", "logplex", "blank"),
			},
			Description: "How the message should be formatted. Valid values are `classic`, `loggly`, `logplex`, and `blank`. Default `blank`.",
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
	}
}

// vclOnlyAttributes returns the Sumo Logic logging attributes that only
// affect generated VCL and have no meaning for Compute services.
func vclOnlyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"format": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(constants.LoggingSumologicDefaultFormat),
			Validators: []validator.String{
				stringvalidator.LengthAtMost(maximumFormatLength),
			},
			Description: "A Fastly [log format string](https://www.fastly.com/documentation/guides/integrations/streaming-logs/custom-log-formats/). Must produce valid content for the configured `message_type`.",
		},
		"format_version": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(DefaultFormatVersion),
			Validators: []validator.Int64{
				int64validator.Between(1, 2),
			},
			Description: "The version of the custom logging format used for the configured endpoint. The logging call gets placed by default in `vcl_log` if format_version is set to `2` and in `vcl_deliver` if `format_version` is set to `1`.",
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

// NestedBlockSchema returns the Sumo Logic logging nested block schema for
// VCL services (service_cdn_auto.logging_sumologic), including the VCL-only
// attributes.
func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Sumo Logic logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

// ComputeNestedBlockSchema returns the Sumo Logic logging nested block schema
// for Compute services (service_compute_auto.logging_sumologic), omitting the
// VCL-only attributes.
func ComputeNestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Sumo Logic logging endpoints attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: ComputeAttributes(),
		},
	}
}

// ValidateNoVCLOnlyAttributesForCompute returns an error diagnostic if format,
// format_version, placement, or response_condition are explicitly configured
// on a Compute service. The standalone fastly_service_logging_sumologic
// resource has one schema shared by both service types — unlike the nested
// blocks, which have distinct VCL (NestedBlockSchema) and Compute
// (ComputeNestedBlockSchema) schemas — so this is the only way to catch the
// mistake before it silently sends unsupported VCL-only attributes to a
// Compute service.
func ValidateNoVCLOnlyAttributesForCompute(ctx context.Context, cfg tfsdk.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	var format, placement, responseCondition types.String
	var formatVersion types.Int64

	diags.Append(cfg.GetAttribute(ctx, path.Root("format"), &format)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root("format_version"), &formatVersion)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root("placement"), &placement)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root("response_condition"), &responseCondition)...)
	if diags.HasError() {
		return diags
	}

	var configured []string
	if !format.IsNull() {
		configured = append(configured, "format")
	}
	if !formatVersion.IsNull() {
		configured = append(configured, "format_version")
	}
	if !placement.IsNull() {
		configured = append(configured, "placement")
	}
	if !responseCondition.IsNull() {
		configured = append(configured, "response_condition")
	}

	if len(configured) > 0 {
		diags.AddError(
			"VCL-only attributes not supported on Compute services",
			"The following attributes only affect generated VCL and are not supported when `service_id` refers to a Compute service: "+
				strings.Join(configured, ", ")+". Remove them from this configuration.",
		)
	}

	return diags
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Sumologic, error) {
	return client.ListSumologics(ctx, &fastly.ListSumologicsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o ops) GetName(api *fastly.Sumologic) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteSumologic(ctx, &fastly.DeleteSumologicInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Sumologic, error) {
	input := BuildCreateInput(serviceID, version, desired)
	return client.CreateSumologic(ctx, input)
}

func (o ops) Equal(desired NestedModel, remote *fastly.Sumologic) bool {
	return desired.ModelsEqual(FlattenToNestedModel(remote))
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Sumologic, error) {
	input := BuildUpdateInput(serviceID, version, desired)
	return client.UpdateSumologic(ctx, input)
}

func (o ops) ToModel(api *fastly.Sumologic) NestedModel {
	return FlattenToNestedModel(api)
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Sumologic]{
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

func (o computeOps) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Sumologic, error) {
	return client.ListSumologics(ctx, &fastly.ListSumologicsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
}

func (o computeOps) GetName(api *fastly.Sumologic) string {
	return fastly.ToValue(api.Name)
}

func (o computeOps) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteSumologic(ctx, &fastly.DeleteSumologicInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o computeOps) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.Sumologic, error) {
	input := BuildComputeCreateInput(serviceID, version, desired)
	return client.CreateSumologic(ctx, input)
}

func (o computeOps) Equal(desired ComputeNestedModel, remote *fastly.Sumologic) bool {
	return desired.ModelsEqual(FlattenToComputeNestedModel(remote))
}

func (o computeOps) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired ComputeNestedModel) (*fastly.Sumologic, error) {
	input := BuildComputeUpdateInput(serviceID, version, desired)
	return client.UpdateSumologic(ctx, input)
}

func (o computeOps) ToModel(api *fastly.Sumologic) ComputeNestedModel {
	return FlattenToComputeNestedModel(api)
}

var computeReconciler = &reconcile.Resource[ComputeNestedModel, fastly.Sumologic]{
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
