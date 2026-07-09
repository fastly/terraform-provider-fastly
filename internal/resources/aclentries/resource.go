package aclentries

import (
	"context"
	"fmt"
	"strings"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/go-fastly/v16/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithModifyPlan = &Resource{}

type Resource struct {
	client *fastly.Client
}

func NewResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_entries"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages CIDR-based allow/block entries within a Fastly ACL.",
		Attributes:  ResourceAttributes(),
	}
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}

	r.client = data.Client
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aclID := plan.ACLID.ValueString()

	tflog.Debug(ctx, "Creating Fastly ACL entries", map[string]any{
		"acl_id": aclID,
	})

	newEntries := expandEntries(ctx, plan.Entries, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	batch := buildBatchEntries(nil, newEntries, plan.ManageEntries.ValueBool())
	if err := r.updateEntries(ctx, aclID, batch); err != nil {
		resp.Diagnostics.AddError("Error creating ACL entries", fmt.Sprintf("ACL %s: %s", aclID, err))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/entries", aclID))

	if len(newEntries) == 0 && !plan.ManageEntries.ValueBool() {
		tflog.Debug(ctx, "Skipping ACL entries refresh after create: manage_entries is false and no entries were configured")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	remote, err := r.listEntries(ctx, aclID)
	if err != nil {
		resp.Diagnostics.AddError("Error refreshing ACL entries", err.Error())
		return
	}

	plan.Entries = flattenEntries(remote, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.ManageEntries.ValueBool() {
		tflog.Debug(ctx, "Skipping ACL entries refresh: manage_entries is false")
		return
	}

	aclID := state.ACLID.ValueString()

	tflog.Debug(ctx, "Reading Fastly ACL entries", map[string]any{
		"acl_id": aclID,
	})

	remote, err := r.listEntries(ctx, aclID)
	if err != nil {
		if errors.IsNotFound(err) {
			tflog.Warn(ctx, "ACL not found, removing entries from state", map[string]any{
				"acl_id": aclID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading ACL entries", err.Error())
		return
	}

	state.Entries = flattenEntries(remote, &resp.Diagnostics)
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

	aclID := plan.ACLID.ValueString()

	tflog.Debug(ctx, "Updating Fastly ACL entries", map[string]any{
		"acl_id": aclID,
	})

	oldEntries := expandEntries(ctx, state.Entries, &resp.Diagnostics)
	newEntries := expandEntries(ctx, plan.Entries, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	batch := buildBatchEntries(oldEntries, newEntries, plan.ManageEntries.ValueBool())
	if err := r.updateEntries(ctx, aclID, batch); err != nil {
		resp.Diagnostics.AddError("Error updating ACL entries", fmt.Sprintf("ACL %s: %s", aclID, err))
		return
	}

	remote, err := r.listEntries(ctx, aclID)
	if err != nil {
		resp.Diagnostics.AddError("Error refreshing ACL entries", err.Error())
		return
	}

	plan.Entries = flattenEntries(remote, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aclID := state.ACLID.ValueString()

	tflog.Debug(ctx, "Deleting Fastly ACL entries", map[string]any{
		"acl_id": aclID,
	})

	entries := expandEntries(ctx, state.Entries, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	batch := buildBatchEntries(entries, nil, true)
	if err := r.updateEntries(ctx, aclID, batch); err != nil {
		if errors.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting ACL entries", fmt.Sprintf("ACL %s: %s", aclID, err))
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	aclID, suffix, ok := strings.Cut(req.ID, "/")
	if !ok || suffix != "entries" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Invalid id: %s. The ID should be in the format <acl_id>/entries", req.ID),
		)
		return
	}

	tflog.Debug(ctx, "Importing ACL entries", map[string]any{
		"acl_id": aclID,
	})

	remote, err := r.listEntries(ctx, aclID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ACL entries", err.Error())
		return
	}
	entries := flattenEntries(remote, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%s/entries", aclID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("acl_id"), aclID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entries"), entries)...)
}

// ModifyPlan preserves the prior state's entries when manage_entries is
// false, so unmanaged drift in the config doesn't surface as a plan diff.
func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ManageEntries.ValueBool() {
		return
	}

	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.ACLID.Equal(state.ACLID) {
		return
	}

	plan.Entries = state.Entries
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *Resource) listEntries(ctx context.Context, aclID string) ([]computeacls.ComputeACLEntry, error) {
	var entries []computeacls.ComputeACLEntry
	var cursor *string

	for {
		page, err := computeacls.ListEntries(ctx, r.client, &computeacls.ListEntriesInput{
			ComputeACLID: &aclID,
			Cursor:       cursor,
		})
		if err != nil {
			return nil, err
		}

		entries = append(entries, page.Entries...)

		if page.Meta.NextCursor == "" {
			break
		}
		cursor = new(page.Meta.NextCursor)
	}

	return entries, nil
}

func (r *Resource) updateEntries(ctx context.Context, aclID string, batch []*computeacls.BatchComputeACLEntry) error {
	if len(batch) == 0 {
		return nil
	}

	return computeacls.Update(ctx, r.client, &computeacls.UpdateInput{
		ComputeACLID: &aclID,
		Entries:      batch,
	})
}
