package kvstore

import (
	"context"
	"fmt"
	"sort"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

type Resource struct {
	client *fastly.Client
}

func NewResource() resource.Resource {
	return &Resource{}
}

type Model struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Location     types.String `tfsdk:"location"`
	ForceDestroy types.Bool   `tfsdk:"force_destroy"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kvstore"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a KV Store, a low-latency, high-throughput key-value data store that is accessible to Compute services during request processing.",
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

	tflog.Debug(ctx, "Creating Fastly KV Store", map[string]any{
		"name": plan.Name.ValueString(),
	})

	store, err := r.client.CreateKVStore(ctx, &fastly.CreateKVStoreInput{
		Name:     plan.Name.ValueString(),
		Location: plan.Location.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating KV Store", err.Error())
		return
	}

	flatten(&plan, store)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Fastly KV Store", map[string]any{
		"id": state.ID.ValueString(),
	})

	store, err := r.client.GetKVStore(ctx, &fastly.GetKVStoreInput{
		StoreID: state.ID.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			tflog.Warn(ctx, "KV Store not found, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading KV Store", err.Error())
		return
	}

	// force_destroy has no server-side representation, so on import (where state starts
	// with only the ID populated) it must be explicitly defaulted here to match the
	// schema default; otherwise it would be left null instead of false.
	if state.ForceDestroy.IsNull() {
		state.ForceDestroy = types.BoolValue(false)
	}

	// The API does not return the store's location, so it is left untouched
	// here and only ever reflects what was set at creation time.
	flatten(&state, store)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update only ever handles force_destroy: name and location both force
// replacement, since there is no API endpoint to modify either in place.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storeID := state.ID.ValueString()

	tflog.Debug(ctx, "Deleting Fastly KV Store", map[string]any{
		"id": storeID,
	})

	if !state.ForceDestroy.ValueBool() {
		empty, err := r.isEmpty(ctx, storeID)
		if err != nil {
			resp.Diagnostics.AddError("Error checking KV Store contents", err.Error())
			return
		}
		if !empty {
			resp.Diagnostics.AddError(
				"KV Store is not empty",
				fmt.Sprintf("cannot delete KV Store (%s), it is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", storeID),
			)
			return
		}
	}

	if err := r.deleteAllKeys(ctx, storeID); err != nil {
		resp.Diagnostics.AddError("Error deleting KV Store keys", err.Error())
		return
	}

	err := r.client.DeleteKVStore(ctx, &fastly.DeleteKVStoreInput{
		StoreID: storeID,
	})
	if err != nil && !errors.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting KV Store", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) isEmpty(ctx context.Context, storeID string) (bool, error) {
	keys, err := r.client.ListKVStoreKeys(ctx, &fastly.ListKVStoreKeysInput{
		StoreID: storeID,
	})
	if err != nil {
		return false, err
	}
	return len(keys.Data) == 0, nil
}

// deleteAllKeys removes every key from the store; the store itself cannot
// be deleted by the API while it still contains entries.
func (r *Resource) deleteAllKeys(ctx context.Context, storeID string) error {
	p := r.client.NewListKVStoreKeysPaginator(ctx, &fastly.ListKVStoreKeysInput{
		StoreID: storeID,
	})
	for p.Next() {
		keys := p.Keys()
		sort.Strings(keys)
		for _, key := range keys {
			err := r.client.DeleteKVStoreKey(ctx, &fastly.DeleteKVStoreKeyInput{
				StoreID: storeID,
				Key:     key,
			})
			if err != nil {
				return fmt.Errorf("error during KV Store key cleanup: %w", err)
			}
		}
	}
	if err := p.Err(); err != nil {
		return fmt.Errorf("error during KV Store cleanup pagination: %w", err)
	}
	return nil
}

func flatten(m *Model, store *fastly.KVStore) {
	if store == nil {
		return
	}

	m.ID = types.StringValue(store.StoreID)
	m.Name = types.StringValue(store.Name)
}
