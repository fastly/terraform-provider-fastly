package servicecompute

import (
	"context"
	"fmt"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Resource struct {
	client *fastly.Client
}

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithIdentity = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

type Model struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Comment      types.String `tfsdk:"comment"`
	ForceDestroy types.Bool   `tfsdk:"force_destroy"`
	Reuse        types.Bool   `tfsdk:"reuse"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_compute"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fastly Compute service resource. Version lifecycle is managed outside normal resource CRUD.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Fastly service ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Fastly service name.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Optional service comment.",
			},
			"force_destroy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Deactivate an active service version before deleting the service. Default `false`.",
			},
			"reuse": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Deactivate an active service version but do not delete the service, allowing it to be reused/imported elsewhere. Default `false`.",
			},
		},
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

	tflog.Debug(ctx, "Creating Fastly Compute service", map[string]any{
		"name":    plan.Name.ValueString(),
		"comment": plan.Comment.ValueString(),
	})

	created, err := r.client.CreateService(ctx, &fastly.CreateServiceInput{
		Name:    fastly.ToPointer(plan.Name.ValueString()),
		Comment: fastly.ToPointer(plan.Comment.ValueString()),
		Type:    fastly.ToPointer(service.TypeCompute),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Fastly Compute service", err.Error())
		return
	}

	if created.ServiceID == nil {
		resp.Diagnostics.AddError("Error creating Fastly Compute service", "Fastly API returned nil service ID.")
		return
	}

	plan.ID = types.StringValue(*created.ServiceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	details, err := r.client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: state.ID.ValueString(),
	})
	if err != nil {
		if httpErr, ok := err.(*fastly.HTTPError); ok && httpErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Fastly Compute service", err.Error())
		return
	}

	serviceType := fastly.ToValue(details.Type)
	if serviceType != service.TypeCompute {
		resp.Diagnostics.AddError(
			"Unexpected Fastly service type",
			fmt.Sprintf("Expected Compute service %q to have type %q, got %q.", state.ID.ValueString(), service.TypeCompute, serviceType),
		)
		return
	}

	if details.ServiceID != nil {
		state.ID = types.StringValue(*details.ServiceID)
	}
	if details.Name != nil {
		state.Name = types.StringValue(*details.Name)
	}
	if details.Comment != nil {
		state.Comment = types.StringValue(*details.Comment)
	} else {
		state.Comment = types.StringValue("")
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

	_, err := r.client.UpdateService(ctx, &fastly.UpdateServiceInput{
		ServiceID: state.ID.ValueString(),
		Name:      fastly.ToPointer(plan.Name.ValueString()),
		Comment:   fastly.ToPointer(plan.Comment.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating Fastly Compute service", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := service.DeleteWithPolicy(ctx, r.client, state.ID.ValueString(), service.BoolValue(state.ForceDestroy), service.BoolValue(state.Reuse)); err != nil {
		resp.Diagnostics.AddError("Error deleting Fastly Compute service", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("service_id"), req, resp)
}

func (r *Resource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"service_id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Fastly service ID.",
			},
		},
	}
}
