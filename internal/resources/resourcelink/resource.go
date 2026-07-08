package resourcelink

import (
	"context"
	"fmt"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/importutil"
	"github.com/fastly/terraform-provider-fastly/internal/service"
	"github.com/fastly/terraform-provider-fastly/internal/validation"

	fastly "github.com/fastly/go-fastly/v16/fastly"
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
	resp.TypeName = req.ProviderTypeName + "_service_resource_link"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Links a shared resource (such as a KV Store or Config Store) to a Fastly service version, making it accessible from Compute code. Writes directly to the specified writable service version.",
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

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, plan.Service.ValueString(), "fastly_service_resource_link", service.TypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Creating Fastly resource link", map[string]any{
		"service_id":  plan.Service.ValueString(),
		"version":     plan.Version.ValueInt64(),
		"name":        service.StringValue(plan.Name),
		"resource_id": service.StringValue(plan.ResourceID),
	})

	resp.Diagnostics.Append(r.providerData.VersionChecker.EnsureMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	a, err := r.providerData.Client.CreateResource(ctx, BuildCreateInput(plan.Service.ValueString(), int(plan.Version.ValueInt64()), plan.NestedModel))
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource link", err.Error())
		return
	}

	flatten(ctx, a, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Fastly resource link from API", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"link_id":    state.LinkID.ValueString(),
	})

	a, err := r.providerData.Client.GetResource(ctx, &fastly.GetResourceInput{
		ResourceID:     state.LinkID.ValueString(),
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			tflog.Warn(ctx, "Resource link not found, removing from state", map[string]any{
				"service_id": state.Service.ValueString(),
				"version":    state.Version.ValueInt64(),
				"link_id":    state.LinkID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading resource link", err.Error())
		return
	}

	flatten(ctx, a, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update handles two independent, non-replacing changes: a rename (name changed, same
// version) applied via UpdateResource, and a move to a different writable version (version
// changed) where the link must already exist in the target version (e.g. because it was
// cloned from the prior one) since there's no API to create-or-rename across versions.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.providerData.VersionChecker.EnsureMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	linkID := state.LinkID.ValueString()
	if plan.Version.ValueInt64() != state.Version.ValueInt64() {
		tflog.Debug(ctx, "Reading Fastly resource link for new version", map[string]any{
			"service_id":  plan.Service.ValueString(),
			"version":     plan.Version.ValueInt64(),
			"resource_id": service.StringValue(plan.ResourceID),
		})

		found, err := findLinkID(ctx, r.providerData.Client, plan.Service.ValueString(), int(plan.Version.ValueInt64()), service.StringValue(plan.ResourceID))
		if err != nil {
			resp.Diagnostics.AddError("Error reading resource link for new version", err.Error())
			return
		}
		if found == "" {
			resp.Diagnostics.AddError(
				"Resource link not found in target version",
				fmt.Sprintf(
					"Service %q version %d has no resource link for resource_id %q. Clone a version that already contains this link before switching to it.",
					plan.Service.ValueString(), plan.Version.ValueInt64(), service.StringValue(plan.ResourceID),
				),
			)
			return
		}
		linkID = found
	}

	a, err := r.providerData.Client.UpdateResource(ctx, &fastly.UpdateResourceInput{
		ResourceID:     linkID,
		Name:           plan.Name.ValueStringPointer(),
		ServiceID:      plan.Service.ValueString(),
		ServiceVersion: int(plan.Version.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating resource link", err.Error())
		return
	}

	flatten(ctx, a, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Fastly resource link", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"link_id":    state.LinkID.ValueString(),
	})

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, state.Service.ValueString(), "fastly_service_resource_link", service.TypeCompute); err != nil {
		if errors.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	notFound, diags := r.providerData.VersionChecker.EnsureMutableForDelete(ctx, state.Service.ValueString(), int(state.Version.ValueInt64()))
	resp.Diagnostics.Append(diags...)
	if notFound || resp.Diagnostics.HasError() {
		return
	}

	err := r.providerData.Client.DeleteResource(ctx, &fastly.DeleteResourceInput{
		ResourceID:     state.LinkID.ValueString(),
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
	})
	if err != nil && !errors.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting resource link", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serviceID, version, name, err := importutil.ParseCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: service_id/version/name\n"+
				"For example: service123/3/my-acl\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Importing resource link", map[string]any{
		"service_id": serviceID,
		"version":    version,
		"name":       name,
	})

	items, err := ops{}.List(ctx, r.providerData.Client, serviceID, version)
	if err != nil {
		resp.Diagnostics.AddError("Error importing resource link", err.Error())
		return
	}

	var found *fastly.Resource
	for _, item := range items {
		if fastly.ToValue(item.Name) == name {
			found = item
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError(
			"Resource link not found",
			fmt.Sprintf("Service %q version %d has no resource link named %q.", serviceID, version, name),
		)
		return
	}

	var state Model
	flatten(ctx, found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
