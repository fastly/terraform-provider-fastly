package productenablement

import (
	"context"
	"fmt"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/go-fastly/v16/fastly/products"
	"github.com/fastly/go-fastly/v16/fastly/products/apidiscovery"
	"github.com/fastly/go-fastly/v16/fastly/products/brotlicompression"
	"github.com/fastly/go-fastly/v16/fastly/products/domaininspector"
	"github.com/fastly/go-fastly/v16/fastly/products/fanout"
	"github.com/fastly/go-fastly/v16/fastly/products/imageoptimizer"
	"github.com/fastly/go-fastly/v16/fastly/products/logexplorerinsights"
	"github.com/fastly/go-fastly/v16/fastly/products/origininspector"
	"github.com/fastly/go-fastly/v16/fastly/products/websockets"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// simpleProductSpec describes a Fastly product whose enablement is a plain
// on/off switch with no further configuration. Each spec backs its own
// resource type: creating that resource enables the product on service_id,
// destroying it disables the product. There is no separate "enabled"
// attribute - the resource's existence is the enablement.
type simpleProductSpec struct {
	// attrName is the resource type name suffix (e.g. "fanout") and the
	// identifier used in error messages.
	attrName string
	// displayName is the human-readable product name used in schema
	// descriptions.
	displayName string
	// restrictTo is the only service type this product is valid for, or ""
	// if it's supported on both CDN and Compute services.
	restrictTo string
	get        func(ctx context.Context, c *fastly.Client, serviceID string) (products.EnableOutput, error)
	enable     func(ctx context.Context, c *fastly.Client, serviceID string) (products.EnableOutput, error)
	disable    func(ctx context.Context, c *fastly.Client, serviceID string) error
}

var (
	fanoutSpec = simpleProductSpec{
		attrName:    "fanout",
		displayName: "Fanout",
		restrictTo:  service.TypeCompute,
		get:         fanout.Get,
		enable:      fanout.Enable,
		disable:     fanout.Disable,
	}
	brotliCompressionSpec = simpleProductSpec{
		attrName:    "brotli_compression",
		displayName: "Brotli Compression",
		restrictTo:  service.TypeVCL,
		get:         brotlicompression.Get,
		enable:      brotlicompression.Enable,
		disable:     brotlicompression.Disable,
	}
	imageOptimizerSpec = simpleProductSpec{
		attrName:    "image_optimizer",
		displayName: "Image Optimizer",
		restrictTo:  service.TypeVCL,
		get:         imageoptimizer.Get,
		enable:      imageoptimizer.Enable,
		disable:     imageoptimizer.Disable,
	}
	originInspectorSpec = simpleProductSpec{
		attrName:    "origin_inspector",
		displayName: "Origin Inspector",
		get:         origininspector.Get,
		enable:      origininspector.Enable,
		disable:     origininspector.Disable,
	}
	domainInspectorSpec = simpleProductSpec{
		attrName:    "domain_inspector",
		displayName: "Domain Inspector",
		get:         domaininspector.Get,
		enable:      domaininspector.Enable,
		disable:     domaininspector.Disable,
	}
	websocketsSpec = simpleProductSpec{
		attrName:    "websockets",
		displayName: "WebSockets",
		get:         websockets.Get,
		enable:      websockets.Enable,
		disable:     websockets.Disable,
	}
	logExplorerInsightsSpec = simpleProductSpec{
		attrName:    "log_explorer_insights",
		displayName: "Log Explorer & Insights",
		get:         logexplorerinsights.Get,
		enable:      logexplorerinsights.Enable,
		disable:     logexplorerinsights.Disable,
	}
	apiDiscoverySpec = simpleProductSpec{
		attrName:    "api_discovery",
		displayName: "API Discovery",
		get:         apidiscovery.Get,
		enable:      apidiscovery.Enable,
		disable:     apidiscovery.Disable,
	}
)

// NewFanoutResource, NewBrotliCompressionResource, ... are registered in
// provider.go's Resources().
func NewFanoutResource() resource.Resource { return &simpleProductResource{spec: fanoutSpec} }
func NewBrotliCompressionResource() resource.Resource {
	return &simpleProductResource{spec: brotliCompressionSpec}
}
func NewImageOptimizerResource() resource.Resource {
	return &simpleProductResource{spec: imageOptimizerSpec}
}
func NewOriginInspectorResource() resource.Resource {
	return &simpleProductResource{spec: originInspectorSpec}
}
func NewDomainInspectorResource() resource.Resource {
	return &simpleProductResource{spec: domainInspectorSpec}
}
func NewWebsocketsResource() resource.Resource { return &simpleProductResource{spec: websocketsSpec} }
func NewLogExplorerInsightsResource() resource.Resource {
	return &simpleProductResource{spec: logExplorerInsightsSpec}
}
func NewAPIDiscoveryResource() resource.Resource {
	return &simpleProductResource{spec: apiDiscoverySpec}
}

var _ resource.Resource = &simpleProductResource{}
var _ resource.ResourceWithImportState = &simpleProductResource{}
var _ resource.ResourceWithModifyPlan = &simpleProductResource{}

type simpleProductModel struct {
	ID        types.String `tfsdk:"id"`
	ServiceID types.String `tfsdk:"service_id"`
}

type simpleProductResource struct {
	spec        simpleProductSpec
	client      *fastly.Client
	typeChecker *service.ServiceTypeChecker
}

func (r *simpleProductResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product_enablement_" + r.spec.attrName
}

func (r *simpleProductResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	desc := fmt.Sprintf("Enables %s on a service. Product Enablement operates on the service directly rather than a specific service version, so this resource is not tied to a `version` and applies immediately.", r.spec.displayName)
	switch r.spec.restrictTo {
	case service.TypeCompute:
		desc += " Only supported for Compute services."
	case service.TypeVCL:
		desc += " Only supported for CDN services."
	}

	resp.Schema = schema.Schema{
		Description: desc,
		Attributes: map[string]schema.Attribute{
			"id":         idAttribute(),
			"service_id": serviceIDAttribute(r.spec.displayName),
		},
	}
}

func (r *simpleProductResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}

	r.client = data.Client
	r.typeChecker = data.ServiceTypeChecker
}

func (r *simpleProductResource) validateServiceType(serviceType string) diag.Diagnostics {
	var diags diag.Diagnostics
	if r.spec.restrictTo != "" && r.spec.restrictTo != serviceType {
		only := "CDN"
		if r.spec.restrictTo == service.TypeCompute {
			only = "Compute"
		}
		diags.AddError("Invalid Attribute Combination", fmt.Sprintf("%q is only supported for %s services.", r.spec.attrName, only))
	}
	return diags
}

// ModifyPlan surfaces a service-type mismatch (e.g. enabling fanout on a
// CDN service) as a `terraform plan` error rather than waiting for `apply`,
// whenever service_id is already known. It can't run during `terraform
// validate` (no service_id is available then), and falls back to the same
// check in Create for the rare case where service_id is still unknown at
// plan time (e.g. it comes from a service being created in the same apply).
// A no-op for products with no service-type restriction.
func (r *simpleProductResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if r.spec.restrictTo == "" || req.Plan.Raw.IsNull() {
		return
	}

	var plan simpleProductModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ServiceID.IsUnknown() || plan.ServiceID.IsNull() {
		return
	}

	serviceType, err := r.typeChecker.GetType(ctx, plan.ServiceID.ValueString())
	if err != nil {
		// Don't fail planning over a lookup error (e.g. transient API
		// error, or the service doesn't exist yet); Create will surface a
		// clearer error at apply time.
		return
	}

	resp.Diagnostics.Append(r.validateServiceType(serviceType)...)
}

func (r *simpleProductResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan simpleProductModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Creating Fastly Product Enablement (%s)", r.spec.attrName), map[string]any{
		"service_id": serviceID,
	})

	if r.spec.restrictTo != "" {
		serviceType, err := r.typeChecker.GetType(ctx, serviceID)
		if err != nil {
			resp.Diagnostics.AddError("Error looking up service type", err.Error())
			return
		}
		resp.Diagnostics.Append(r.validateServiceType(serviceType)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if _, err := r.spec.enable(ctx, r.client, serviceID); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error enabling %s", r.spec.attrName), err.Error())
		return
	}

	plan.ID = plan.ServiceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read treats any error from the product's Get endpoint - whether because
// the product was disabled outside Terraform or because the service itself
// is gone - as "no longer enabled" and removes the resource from state.
func (r *simpleProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state simpleProductModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Reading Fastly Product Enablement (%s)", r.spec.attrName), map[string]any{
		"service_id": serviceID,
	})

	if _, err := r.spec.get(ctx, r.client, serviceID); err != nil {
		tflog.Warn(ctx, fmt.Sprintf("%s no longer enabled, removing from state", r.spec.attrName), map[string]any{
			"service_id": serviceID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update never runs in practice: service_id is the only attribute besides
// the computed id, and changing it forces replacement. Implemented only to
// satisfy the resource.Resource interface.
func (r *simpleProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan simpleProductModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete disables the product. Entitlement-related errors (accounts that
// can't self-service disable a product) and a since-deleted service are
// both treated as success so that `terraform destroy` can still complete
// and leave a clean state.
func (r *simpleProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state simpleProductModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Deleting Fastly Product Enablement (%s)", r.spec.attrName), map[string]any{
		"service_id": serviceID,
	})

	if err := r.spec.disable(ctx, r.client, serviceID); err != nil && !isEntitlementError(err) && !errors.IsNotFound(err) {
		resp.Diagnostics.AddError(fmt.Sprintf("Error disabling %s", r.spec.attrName), err.Error())
	}
}

func (r *simpleProductResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("service_id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
