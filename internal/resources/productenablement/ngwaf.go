package productenablement

import (
	"context"
	"strconv"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	"github.com/fastly/go-fastly/v16/fastly"
	ngwafproduct "github.com/fastly/go-fastly/v16/fastly/products/ngwaf"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &NGWAFResource{}
var _ resource.ResourceWithImportState = &NGWAFResource{}
var _ resource.ResourceWithModifyPlan = &NGWAFResource{}

type NGWAFModel struct {
	ID          types.String `tfsdk:"id"`
	ServiceID   types.String `tfsdk:"service_id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	TrafficRamp types.Int64  `tfsdk:"traffic_ramp"`
}

type NGWAFResource struct {
	client      *fastly.Client
	typeChecker *service.ServiceTypeChecker
}

func NewNGWAFResource() resource.Resource {
	return &NGWAFResource{}
}

func (r *NGWAFResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_product_ngwaf"
}

func (r *NGWAFResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enables Next-Gen WAF on a service. Product Enablement operates on the service directly rather than a specific service version, so this resource is not tied to a `version` and applies immediately.",
		Attributes: map[string]schema.Attribute{
			"id":         idAttribute(),
			"service_id": serviceIDAttribute("Next-Gen WAF"),
			"workspace_id": schema.StringAttribute{
				Required:    true,
				Description: "The Next-Gen WAF workspace to link.",
			},
			"traffic_ramp": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Description: "The percentage of traffic to inspect. Only supported for CDN services; defaults to 100.",
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
		},
	}
}

func (r *NGWAFResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}
	r.client = data.Client
	r.typeChecker = data.ServiceTypeChecker
}

func validateNGWAFTrafficRamp(plan *NGWAFModel, serviceType string) diag.Diagnostics {
	var diags diag.Diagnostics
	if serviceType != service.TypeVCL && !plan.TrafficRamp.IsNull() && plan.TrafficRamp.ValueInt64() != 100 {
		diags.AddError("Invalid Attribute Combination", `"traffic_ramp" is only supported for CDN services.`)
	}
	return diags
}

// ModifyPlan surfaces a non-default traffic_ramp on a Compute service as a
// `terraform plan` error rather than waiting for `apply`, whenever
// service_id is already known. It can't run during `terraform validate`
// (no service_id is available then), and falls back to the same check in
// Create/Update for the rare case where service_id is still unknown at
// plan time (e.g. it comes from a service being created in the same apply).
func (r *NGWAFResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan NGWAFModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ServiceID.IsUnknown() || plan.ServiceID.IsNull() {
		return
	}

	serviceType, err := r.typeChecker.GetType(ctx, plan.ServiceID.ValueString())
	if err != nil {
		return
	}

	resp.Diagnostics.Append(validateNGWAFTrafficRamp(&plan, serviceType)...)
}

func (r *NGWAFResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NGWAFModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	tflog.Debug(ctx, "Creating Fastly Product Enablement (ngwaf)", map[string]any{"service_id": serviceID})

	serviceType, err := r.typeChecker.GetType(ctx, serviceID)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up service type", err.Error())
		return
	}

	resp.Diagnostics.Append(validateNGWAFTrafficRamp(&plan, serviceType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := ngwafproduct.Enable(ctx, r.client, serviceID, ngwafproduct.EnableInput{
		WorkspaceID: plan.WorkspaceID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Error enabling ngwaf", err.Error())
		return
	}

	if serviceType == service.TypeVCL && !plan.TrafficRamp.IsNull() && plan.TrafficRamp.ValueInt64() != 100 {
		if _, err := ngwafproduct.UpdateConfiguration(ctx, r.client, serviceID, ngwafproduct.ConfigureInput{
			WorkspaceID: plan.WorkspaceID.ValueString(),
			TrafficRamp: strconv.FormatInt(plan.TrafficRamp.ValueInt64(), 10),
		}); err != nil {
			resp.Diagnostics.AddError("Error configuring ngwaf", err.Error())
			return
		}
	}

	plan.ID = plan.ServiceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NGWAFResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NGWAFModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	tflog.Debug(ctx, "Reading Fastly Product Enablement (ngwaf)", map[string]any{"service_id": serviceID})

	if _, err := ngwafproduct.Get(ctx, r.client, serviceID); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	serviceType, err := r.typeChecker.GetType(ctx, serviceID)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up service type", err.Error())
		return
	}

	cfg, err := ngwafproduct.GetConfiguration(ctx, r.client, serviceID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ngwaf configuration", err.Error())
		return
	}

	if cfg.Configuration.WorkspaceID != nil {
		state.WorkspaceID = types.StringValue(*cfg.Configuration.WorkspaceID)
	}

	// traffic_ramp is schema-defaulted to 100 regardless of service type
	// (Terraform Core applies the default before Create/Update ever runs),
	// but it only has an effect for CDN services. Pinning it to 100 for
	// Compute rather than leaving it null keeps refresh consistent with
	// that plan-time default and avoids perpetual drift; ModifyPlan already
	// rejects any other value on a Compute service.
	switch {
	case serviceType == service.TypeVCL && cfg.Configuration.TrafficRamp != nil:
		tr, err := strconv.ParseInt(*cfg.Configuration.TrafficRamp, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError("Error parsing ngwaf traffic_ramp", err.Error())
			return
		}
		state.TrafficRamp = types.Int64Value(tr)
	case serviceType != service.TypeVCL:
		state.TrafficRamp = types.Int64Value(100)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NGWAFResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NGWAFModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	tflog.Debug(ctx, "Updating Fastly Product Enablement (ngwaf)", map[string]any{"service_id": serviceID})

	serviceType, err := r.typeChecker.GetType(ctx, serviceID)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up service type", err.Error())
		return
	}

	resp.Diagnostics.Append(validateNGWAFTrafficRamp(&plan, serviceType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trafficRamp := "100"
	if serviceType == service.TypeVCL && !plan.TrafficRamp.IsNull() {
		trafficRamp = strconv.FormatInt(plan.TrafficRamp.ValueInt64(), 10)
	}

	if _, err := ngwafproduct.UpdateConfiguration(ctx, r.client, serviceID, ngwafproduct.ConfigureInput{
		WorkspaceID: plan.WorkspaceID.ValueString(),
		TrafficRamp: trafficRamp,
	}); err != nil {
		resp.Diagnostics.AddError("Error configuring ngwaf", err.Error())
		return
	}

	plan.ID = plan.ServiceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NGWAFResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NGWAFModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	tflog.Debug(ctx, "Deleting Fastly Product Enablement (ngwaf)", map[string]any{"service_id": serviceID})

	if err := ngwafproduct.Disable(ctx, r.client, serviceID); err != nil && !isEntitlementError(err) && !errors.IsNotFound(err) {
		resp.Diagnostics.AddError("Error disabling ngwaf", err.Error())
	}
}

func (r *NGWAFResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("service_id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
