package cdnacl

import (
	"context"
	"fmt"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/importutil"
	"github.com/fastly/terraform-provider-fastly/internal/service"
	"github.com/fastly/terraform-provider-fastly/internal/validation"

	"github.com/fastly/go-fastly/v16/fastly"
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
	resp.TypeName = req.ProviderTypeName + "_service_cdn_acl"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fastly CDN service ACL resource. Writes directly to the specified writable service version.",
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

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, plan.Service.ValueString(), "fastly_service_cdn_acl", service.TypeVCL); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Creating Fastly service ACL", map[string]any{
		"service_id": plan.Service.ValueString(),
		"version":    plan.Version.ValueInt64(),
		"name":       service.StringValue(plan.Name),
	})

	resp.Diagnostics.Append(r.providerData.VersionChecker.EnsureMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := BuildCreateInput(plan.Service.ValueString(), int(plan.Version.ValueInt64()), plan.NestedModel)

	a, err := r.providerData.Client.CreateACL(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating explicit service ACL", err.Error())
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

	tflog.Debug(ctx, "Reading Fastly service ACL from API", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	a, err := r.providerData.Client.GetACL(ctx, &fastly.GetACLInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			tflog.Warn(ctx, "Service ACL not found, removing from state", map[string]any{
				"service_id": state.Service.ValueString(),
				"version":    state.Version.ValueInt64(),
				"name":       state.Name.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading explicit service ACL", err.Error())
		return
	}

	flatten(ctx, a, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update only runs when the ACL's version changes in place: service_id and name
// both force replacement, so this is never reached for a rename or a move to a
// different service. The target version must already contain an ACL with this
// name (e.g. because it was cloned from the prior version), since the ACL API has
// no update operation to create or rename one here. Fetch that ACL and flatten it
// into state so id/acl_id reflect the new version rather than the old one.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Fastly service ACL for new version", map[string]any{
		"service_id": plan.Service.ValueString(),
		"version":    plan.Version.ValueInt64(),
		"name":       service.StringValue(plan.Name),
	})

	a, err := r.providerData.Client.GetACL(ctx, &fastly.GetACLInput{
		ServiceID:      plan.Service.ValueString(),
		ServiceVersion: int(plan.Version.ValueInt64()),
		Name:           plan.Name.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"ACL not found in target version",
				fmt.Sprintf(
					"Service %q version %d has no ACL named %q. Clone a version that already contains this ACL before switching to it.",
					plan.Service.ValueString(),
					plan.Version.ValueInt64(),
					plan.Name.ValueString(),
				),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading explicit service ACL for new version", err.Error())
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

	tflog.Debug(ctx, "Deleting Fastly service ACL", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, state.Service.ValueString(), "fastly_service_cdn_acl", service.TypeVCL); err != nil {
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

	if !service.BoolValue(state.ForceDestroy) {
		mayDelete, err := isACLEmpty(ctx, state.Service.ValueString(), state.ACLID.ValueString(), r.providerData.Client)
		if err != nil {
			resp.Diagnostics.AddError("Error checking if ACL is empty", err.Error())
			return
		}

		if !mayDelete {
			resp.Diagnostics.AddError(
				"Cannot delete non-empty ACL",
				fmt.Sprintf("Cannot delete ACL %q because it contains entries. Either delete the entries first, or set force_destroy to true and apply it before making this change.", state.ACLID.ValueString()),
			)
			return
		}
	}

	err := r.providerData.Client.DeleteACL(ctx, &fastly.DeleteACLInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting explicit service ACL", err.Error())
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

	tflog.Debug(ctx, "Importing ACL", map[string]any{
		"service_id": serviceID,
		"version":    version,
		"name":       name,
	})

	a, err := r.providerData.Client.GetACL(ctx, &fastly.GetACLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error importing ACL", err.Error())
		return
	}

	var state Model
	state.Service = types.StringValue(serviceID)
	state.Version = types.Int64Value(int64(version))
	flatten(ctx, a, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func isACLEmpty(ctx context.Context, serviceID, aclID string, client *fastly.Client) (bool, error) {
	entries, err := client.ListACLEntries(ctx, &fastly.ListACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}
