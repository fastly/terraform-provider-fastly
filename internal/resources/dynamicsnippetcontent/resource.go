package dynamicsnippetcontent

import (
	"context"
	"fmt"
	"strings"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/service"
	"github.com/fastly/terraform-provider-fastly/internal/validation"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

type Resource struct {
	providerData *fastlyclient.Data
}

func NewResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_dynamic_snippet_content"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fastly dynamic VCL snippet content resource. Updates versionless dynamic snippet code directly by service ID and snippet ID.",
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

func (r *Resource) ensureServiceTypeSupported(ctx context.Context, serviceID string) error {
	return validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, serviceID, "fastly_service_dynamic_snippet_content", service.TypeVCL)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.ensureServiceTypeSupported(ctx, plan.Service.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Creating Fastly dynamic VCL snippet content", map[string]any{
		"service_id": plan.Service.ValueString(),
		"snippet_id": plan.SnippetID.ValueString(),
	})

	if err := r.updateContent(ctx, plan.Service.ValueString(), plan.SnippetID.ValueString(), plan.Content.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating dynamic VCL snippet content", err.Error())
		return
	}

	plan.ID = types.StringValue(ID(plan.Service.ValueString(), plan.SnippetID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.ensureServiceTypeSupported(ctx, state.Service.ValueString()); err != nil {
		if errors.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Reading Fastly dynamic VCL snippet content", map[string]any{
		"service_id": state.Service.ValueString(),
		"snippet_id": state.SnippetID.ValueString(),
	})

	dynamicSnippet, err := r.providerData.Client.GetDynamicSnippet(ctx, &fastly.GetDynamicSnippetInput{
		ServiceID: state.Service.ValueString(),
		SnippetID: state.SnippetID.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading dynamic VCL snippet content", err.Error())
		return
	}

	state.ID = types.StringValue(ID(state.Service.ValueString(), state.SnippetID.ValueString()))
	if state.ManageSnippet.ValueBool() {
		state.Content = types.StringValue(fastly.ToValue(dynamicSnippet.Content))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.ensureServiceTypeSupported(ctx, plan.Service.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Updating Fastly dynamic VCL snippet content", map[string]any{
		"service_id": plan.Service.ValueString(),
		"snippet_id": plan.SnippetID.ValueString(),
	})

	if err := r.updateContent(ctx, plan.Service.ValueString(), plan.SnippetID.ValueString(), plan.Content.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating dynamic VCL snippet content", err.Error())
		return
	}

	plan.ID = types.StringValue(ID(plan.Service.ValueString(), plan.SnippetID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.ensureServiceTypeSupported(ctx, state.Service.ValueString()); err != nil {
		if errors.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	if state.ManageSnippet.ValueBool() {
		tflog.Debug(ctx, "Clearing Fastly dynamic VCL snippet content", map[string]any{
			"service_id": state.Service.ValueString(),
			"snippet_id": state.SnippetID.ValueString(),
		})

		if err := r.updateContent(ctx, state.Service.ValueString(), state.SnippetID.ValueString(), ""); err != nil && !errors.IsNotFound(err) {
			resp.Diagnostics.AddError("Error deleting dynamic VCL snippet content", err.Error())
			return
		}
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serviceID, snippetID, err := parseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: service_id/snippet_id\n"+
				"For example: service123/abc123\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if err := r.ensureServiceTypeSupported(ctx, serviceID); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	dynamicSnippet, err := r.providerData.Client.GetDynamicSnippet(ctx, &fastly.GetDynamicSnippetInput{
		ServiceID: serviceID,
		SnippetID: snippetID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error importing dynamic VCL snippet content", err.Error())
		return
	}

	state := Model{
		ID:            types.StringValue(ID(serviceID, snippetID)),
		Service:       types.StringValue(serviceID),
		SnippetID:     types.StringValue(snippetID),
		Content:       types.StringValue(fastly.ToValue(dynamicSnippet.Content)),
		ManageSnippet: types.BoolValue(false),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) updateContent(ctx context.Context, serviceID string, snippetID string, content string) error {
	_, err := r.providerData.Client.UpdateDynamicSnippet(ctx, &fastly.UpdateDynamicSnippetInput{
		ServiceID: serviceID,
		SnippetID: snippetID,
		Content:   &content,
	})
	return err
}

func parseImportID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid composite import ID format: expected service_id/snippet_id, got %q", id)
	}

	if parts[0] == "" {
		return "", "", fmt.Errorf("service_id cannot be empty in import ID %q", id)
	}

	if parts[1] == "" {
		return "", "", fmt.Errorf("snippet_id cannot be empty in import ID %q", id)
	}

	return parts[0], parts[1], nil
}
