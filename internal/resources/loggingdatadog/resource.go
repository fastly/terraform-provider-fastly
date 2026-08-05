package loggingdatadog

import (
	"context"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/importutil"
	"github.com/fastly/terraform-provider-fastly/internal/service"
	"github.com/fastly/terraform-provider-fastly/internal/validation"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

type Resource struct {
	providerData *fastlyclient.Data
}

func NewResource() resource.Resource {
	return &Resource{}
}

type Model struct {
	NestedModel
	ID      types.String `tfsdk:"id"`
	Service types.String `tfsdk:"service_id"`
	Version types.Int64  `tfsdk:"version"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_logging_datadog"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fastly service Datadog logging endpoint resource. Writes directly to the specified writable service version.",
		Attributes:  ResourceAttributes(),
	}
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}

	r.providerData = data
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Fastly Datadog logging endpoint", map[string]any{
		"service_id": plan.Service.ValueString(),
		"version":    plan.Version.ValueInt64(),
		"name":       service.StringValue(plan.Name),
	})

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, plan.Service.ValueString(), "fastly_service_logging_datadog", service.TypeVCL, service.TypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported service type", err.Error())
		return
	}

	serviceType, err := r.providerData.ServiceTypeChecker.GetType(ctx, plan.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error determining service type", err.Error())
		return
	}
	if serviceType == service.TypeCompute {
		resp.Diagnostics.Append(ValidateNoVCLOnlyAttributesForCompute(ctx, req.Config)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(r.providerData.VersionChecker.EnsureMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := BuildCreateInput(plan.Service.ValueString(), int(plan.Version.ValueInt64()), plan.NestedModel)
	if serviceType == service.TypeCompute {
		ClearVCLOnlyCreateFields(input)
	}

	d, err := r.providerData.Client.CreateDatadog(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Datadog logging endpoint", err.Error())
		return
	}

	flatten(ctx, d, &plan)
	if serviceType == service.TypeCompute {
		ResetVCLOnlyToDefaults(&plan.NestedModel)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Fastly Datadog logging endpoint from API", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	d, err := r.providerData.Client.GetDatadog(ctx, &fastly.GetDatadogInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			tflog.Warn(ctx, "Datadog logging endpoint not found, removing from state", map[string]any{
				"service_id": state.Service.ValueString(),
				"name":       state.Name.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Datadog logging endpoint", err.Error())
		return
	}

	// The endpoint Get above succeeded, so the service exists and this is a cached
	// lookup in all but the first call per service.
	serviceType, err := r.providerData.ServiceTypeChecker.GetType(ctx, state.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error determining service type", err.Error())
		return
	}

	flatten(ctx, d, &state)
	if serviceType == service.TypeCompute {
		ResetVCLOnlyToDefaults(&state.NestedModel)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model
	var state Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, plan.Service.ValueString(), "fastly_service_logging_datadog", service.TypeVCL, service.TypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported service type", err.Error())
		return
	}

	serviceType, err := r.providerData.ServiceTypeChecker.GetType(ctx, plan.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error determining service type", err.Error())
		return
	}
	if serviceType == service.TypeCompute {
		resp.Diagnostics.Append(ValidateNoVCLOnlyAttributesForCompute(ctx, req.Config)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(r.providerData.VersionChecker.EnsureMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := BuildUpdateInput(plan.Service.ValueString(), int(plan.Version.ValueInt64()), plan.NestedModel)
	if serviceType == service.TypeCompute {
		ClearVCLOnlyUpdateFields(opts)
	}

	tflog.Debug(ctx, "Updating Fastly Datadog logging endpoint", map[string]any{
		"service_id": opts.ServiceID,
		"version":    opts.ServiceVersion,
		"name":       opts.Name,
	})

	d, err := r.providerData.Client.UpdateDatadog(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Datadog logging endpoint", err.Error())
		return
	}

	flatten(ctx, d, &plan)
	if serviceType == service.TypeCompute {
		ResetVCLOnlyToDefaults(&plan.NestedModel)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Fastly Datadog logging endpoint", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, state.Service.ValueString(), "fastly_service_logging_datadog", service.TypeVCL, service.TypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported service type", err.Error())
		return
	}

	notFound, diags := r.providerData.VersionChecker.EnsureMutableForDelete(ctx, state.Service.ValueString(), int(state.Version.ValueInt64()))
	resp.Diagnostics.Append(diags...)
	if notFound || resp.Diagnostics.HasError() {
		return
	}

	err := r.providerData.Client.DeleteDatadog(ctx, &fastly.DeleteDatadogInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting Datadog logging endpoint", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serviceID, version, name, err := importutil.ParseCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: service_id/version/name\n"+
				"For example: service123/3/my-datadog-logger\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Importing Datadog logging endpoint", map[string]any{
		"service_id": serviceID,
		"version":    version,
		"name":       name,
	})

	d, err := r.providerData.Client.GetDatadog(ctx, &fastly.GetDatadogInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error importing Datadog logging endpoint", err.Error())
		return
	}

	serviceType, err := r.providerData.ServiceTypeChecker.GetType(ctx, serviceID)
	if err != nil {
		resp.Diagnostics.AddError("Error determining service type", err.Error())
		return
	}

	var state Model
	state.Service = types.StringValue(serviceID)
	state.Version = types.Int64Value(int64(version))
	flatten(ctx, d, &state)
	if serviceType == service.TypeCompute {
		ResetVCLOnlyToDefaults(&state.NestedModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
