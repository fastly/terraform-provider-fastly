package aclentriescdn

import (
	"context"
	"fmt"
	"strings"

	"github.com/fastly/go-fastly/v16/fastly"
	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/service"
	"github.com/fastly/terraform-provider-fastly/internal/validation"
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
	providerData *fastlyclient.Data
}

func NewResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_cdn_acl_entries"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ACL entries for a Fastly service ACL. Provides batch operations for creating, updating, and deleting ACL entries.",
		Attributes:  ResourceAttributes(),
		Blocks:      ResourceBlocks(),
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

	serviceID := plan.ServiceID.ValueString()
	aclID := plan.ACLID.ValueString()

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, serviceID, "fastly_service_cdn_acl_entries", service.TypeVCL); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Creating Fastly service ACL entries", map[string]any{
		"service_id": serviceID,
		"acl_id":     aclID,
	})

	entries := expandEntries(ctx, plan.Entry, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var batchEntries []*fastly.BatchACLEntry
	for _, entry := range entries {
		batchEntry := buildBatchACLEntry(ctx, entry, fastly.CreateBatchOperation)
		batchEntries = append(batchEntries, batchEntry)
	}

	err := executeBatchACLOperations(ctx, r.providerData.Client, serviceID, aclID, batchEntries)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ACL entries",
			fmt.Sprintf("service %s, ACL %s: %s", serviceID, aclID, err),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", serviceID, aclID))

	if len(entries) == 0 && !plan.ManageEntries.ValueBool() {
		tflog.Debug(ctx, "Skipping ACL entries refresh after create: manage_entries is false and no entries were configured")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	refreshedEntries, err := r.refreshEntries(ctx, serviceID, aclID)
	if err != nil {
		resp.Diagnostics.AddError("Error refreshing ACL entries", err.Error())
		return
	}

	plan.Entry = flattenEntries(ctx, refreshedEntries, plan.Entry, &resp.Diagnostics)
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

	serviceID := state.ServiceID.ValueString()
	aclID := state.ACLID.ValueString()

	tflog.Debug(ctx, "Refreshing ACL entries configuration", map[string]any{
		"service_id": serviceID,
		"acl_id":     aclID,
	})

	remoteState, err := r.refreshEntries(ctx, serviceID, aclID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ACL entries", err.Error())
		return
	}

	for _, entry := range remoteState {
		tflog.Debug(ctx, "Read: Remote entry from API", map[string]any{
			"ip":      entry.IP,
			"negated": entry.Negated,
			"subnet":  entry.Subnet,
		})
	}

	state.Entry = flattenEntries(ctx, remoteState, state.Entry, &resp.Diagnostics)
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

	serviceID := plan.ServiceID.ValueString()
	aclID := plan.ACLID.ValueString()

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, serviceID, "fastly_service_cdn_acl_entries", service.TypeVCL); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Updating Fastly service ACL entries", map[string]any{
		"service_id": serviceID,
		"acl_id":     aclID,
	})

	var batchEntries []*fastly.BatchACLEntry

	oldEntries := expandEntries(ctx, state.Entry, &resp.Diagnostics)
	newEntries := expandEntries(ctx, plan.Entry, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Entries are matched across old/new by (ip, subnet), which is what Fastly's ACL
	// enforces uniqueness on -- not the full content key. Matching on full content
	// (including negated/comment) would treat a comment-only change as a delete of
	// the old entry plus a create of a new one at the same ip/subnet, and the batch
	// API rejects that create as a duplicate before the delete takes effect.
	oldByIdentity := make(map[string]EntryModel)
	for _, e := range oldEntries {
		if !e.ID.IsNull() && !e.ID.IsUnknown() {
			oldByIdentity[entryIdentityKey(e)] = e
		}
	}

	newByIdentity := make(map[string]EntryModel)
	for _, e := range newEntries {
		newByIdentity[entryIdentityKey(e)] = e
	}

	for identityKey, oldEntry := range oldByIdentity {
		if _, exists := newByIdentity[identityKey]; !exists {
			batchEntries = append(batchEntries, &fastly.BatchACLEntry{
				Operation: new(fastly.BatchOperation),
				EntryID:   oldEntry.ID.ValueStringPointer(),
			})
			*batchEntries[len(batchEntries)-1].Operation = fastly.DeleteBatchOperation
			tflog.Debug(ctx, "Deleting entry", map[string]any{"ip": oldEntry.IP.ValueString()})
		}
	}

	for _, newEntry := range newEntries {
		identityKey := entryIdentityKey(newEntry)
		oldEntry, existsInOld := oldByIdentity[identityKey]

		tflog.Debug(ctx, "Processing entry in update", map[string]any{
			"ip":            newEntry.IP.ValueString(),
			"exists_in_old": existsInOld,
		})

		if !existsInOld {
			batchEntry := buildBatchACLEntry(ctx, newEntry, fastly.CreateBatchOperation)
			batchEntries = append(batchEntries, batchEntry)
			tflog.Debug(ctx, "Creating new entry", map[string]any{"ip": newEntry.IP.ValueString()})
			continue
		}

		if !entriesEqual(oldEntry, newEntry) {
			updatedEntry := newEntry
			updatedEntry.ID = oldEntry.ID
			batchEntry := buildBatchACLEntry(ctx, updatedEntry, fastly.UpdateBatchOperation)
			batchEntries = append(batchEntries, batchEntry)
			tflog.Debug(ctx, "Updating existing entry", map[string]any{"ip": newEntry.IP.ValueString()})
		} else {
			tflog.Debug(ctx, "Entry unchanged", map[string]any{"ip": newEntry.IP.ValueString()})
		}
	}

	tflog.Debug(ctx, "Batch operations to execute", map[string]any{"count": len(batchEntries)})

	err := executeBatchACLOperations(ctx, r.providerData.Client, serviceID, aclID, batchEntries)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ACL entries",
			fmt.Sprintf("service %s, ACL %s: %s", serviceID, aclID, err),
		)
		return
	}

	remoteState, err := r.refreshEntries(ctx, serviceID, aclID)
	if err != nil {
		resp.Diagnostics.AddError("Error refreshing ACL entries", err.Error())
		return
	}

	plan.Entry = flattenEntries(ctx, remoteState, plan.Entry, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	aclID := state.ACLID.ValueString()

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, serviceID, "fastly_service_cdn_acl_entries", service.TypeVCL); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Deleting Fastly service ACL entries", map[string]any{
		"service_id": serviceID,
		"acl_id":     aclID,
	})

	entries := expandEntries(ctx, state.Entry, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var batchEntries []*fastly.BatchACLEntry
	for _, entry := range entries {
		if !entry.ID.IsNull() && !entry.ID.IsUnknown() {
			batchEntries = append(batchEntries, &fastly.BatchACLEntry{
				Operation: new(fastly.BatchOperation),
				EntryID:   entry.ID.ValueStringPointer(),
			})
			*batchEntries[len(batchEntries)-1].Operation = fastly.DeleteBatchOperation
		}
	}

	err := executeBatchACLOperations(ctx, r.providerData.Client, serviceID, aclID, batchEntries)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ACL entries",
			fmt.Sprintf("service %s, ACL %s: %s", serviceID, aclID, err),
		)
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	split := strings.Split(req.ID, "/")

	if len(split) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Invalid id: %s. The ID should be in the format [service_id]/[acl_id]", req.ID),
		)
		return
	}

	serviceID := split[0]
	aclID := split[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("acl_id"), aclID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("manage_entries"), true)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *Resource) refreshEntries(ctx context.Context, serviceID, aclID string) ([]*fastly.ACLEntry, error) {
	paginator := r.providerData.Client.GetACLEntries(ctx, &fastly.GetACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})

	var entries []*fastly.ACLEntry
	for paginator.HasNext() {
		results, err := paginator.GetNext()
		if err != nil {
			return nil, err
		}
		entries = append(entries, results...)
	}
	return entries, nil
}

func executeBatchACLOperations(ctx context.Context, client *fastly.Client, serviceID, aclID string, batchACLEntries []*fastly.BatchACLEntry) error {
	batchSize := fastly.BatchModifyMaximumOperations

	tflog.Debug(ctx, "executeBatchACLOperations", map[string]any{
		"total_entries": len(batchACLEntries),
		"batch_size":    batchSize,
	})

	for i, entry := range batchACLEntries {
		negatedStr := "nil"
		if entry.Negated != nil {
			negatedStr = fmt.Sprintf("%v (type: %T)", *entry.Negated, *entry.Negated)
		}
		tflog.Debug(ctx, "Batch entry details", map[string]any{
			"index":   i,
			"ip":      entry.IP,
			"negated": negatedStr,
			"op":      entry.Operation,
		})
	}

	for i := 0; i < len(batchACLEntries); i += batchSize {
		j := min(i+batchSize, len(batchACLEntries))

		tflog.Debug(ctx, "Calling BatchModifyACLEntries", map[string]any{
			"batch_start": i,
			"batch_end":   j,
			"count":       j - i,
		})

		err := client.BatchModifyACLEntries(ctx, &fastly.BatchModifyACLEntriesInput{
			ServiceID: serviceID,
			ACLID:     aclID,
			Entries:   batchACLEntries[i:j],
		})
		if err != nil {
			tflog.Error(ctx, "BatchModifyACLEntries failed", map[string]any{"error": err.Error()})
			return err
		}
		tflog.Debug(ctx, "BatchModifyACLEntries succeeded")
	}

	return nil
}

func entriesEqual(a, b EntryModel) bool {
	return a.IP.Equal(b.IP) &&
		a.Subnet.Equal(b.Subnet) &&
		a.Negated.Equal(b.Negated) &&
		a.Comment.Equal(b.Comment)
}

func plannedEntryContentKey(e EntryModel) string {
	ip := ""
	subnet := int64(0)
	negated := false
	comment := ""

	if !e.IP.IsNull() && !e.IP.IsUnknown() {
		ip = e.IP.ValueString()
	}
	if !e.Subnet.IsNull() && !e.Subnet.IsUnknown() {
		subnet = e.Subnet.ValueInt64()
	}
	if !e.Negated.IsNull() && !e.Negated.IsUnknown() {
		negated = e.Negated.ValueBool()
	}
	if !e.Comment.IsNull() && !e.Comment.IsUnknown() {
		comment = e.Comment.ValueString()
	}

	return fmt.Sprintf("%s|%d|%t|%s", ip, subnet, negated, comment)
}

// entryIdentityKey returns the (ip, subnet) pair Fastly's ACL enforces uniqueness on.
// Used to match entries across old/new state when deciding create vs. update vs. delete,
// as opposed to plannedEntryContentKey which also factors in negated/comment.
func entryIdentityKey(e EntryModel) string {
	ip := ""
	subnet := int64(0)

	if !e.IP.IsNull() && !e.IP.IsUnknown() {
		ip = e.IP.ValueString()
	}
	if !e.Subnet.IsNull() && !e.Subnet.IsUnknown() {
		subnet = e.Subnet.ValueInt64()
	}

	return fmt.Sprintf("%s|%d", ip, subnet)
}

func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.ManageEntries.ValueBool() && !req.State.Raw.IsNull() {
		var state Model
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if plan.ACLID.Equal(state.ACLID) {
			plan.Entry = state.Entry
			resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		}
	}
}
