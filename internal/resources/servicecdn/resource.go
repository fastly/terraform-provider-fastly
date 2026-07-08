package servicecdn

import (
	"context"
	"fmt"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Resource struct {
	providerData *fastlyclient.Data
}

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

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
	resp.TypeName = req.ProviderTypeName + "_service_cdn"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fastly CDN service resource. Version lifecycle is managed outside normal resource CRUD.",
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

	r.providerData = data
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Fastly CDN service", map[string]any{
		"name":    plan.Name.ValueString(),
		"comment": plan.Comment.ValueString(),
	})

	created, err := r.providerData.Client.CreateService(ctx, &fastly.CreateServiceInput{
		Name:    new(plan.Name.ValueString()),
		Comment: new(plan.Comment.ValueString()),
		Type:    new(service.TypeVCL),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Fastly CDN service", err.Error())
		return
	}

	if created.ServiceID == nil {
		resp.Diagnostics.AddError("Error creating Fastly CDN service", "Fastly API returned nil service ID.")
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

	details, err := r.providerData.Client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: state.ID.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Fastly CDN service", err.Error())
		return
	}

	serviceType := fastly.ToValue(details.Type)
	if serviceType != service.TypeVCL {
		resp.Diagnostics.AddError(
			"Unexpected Fastly service type",
			fmt.Sprintf("Expected VCL service %q to have type %q, got %q.", state.ID.ValueString(), service.TypeVCL, serviceType),
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

	_, err := r.providerData.Client.UpdateService(ctx, &fastly.UpdateServiceInput{
		ServiceID: state.ID.ValueString(),
		Name:      new(plan.Name.ValueString()),
		Comment:   new(plan.Comment.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating Fastly CDN service", err.Error())
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

	if err := service.DeleteWithPolicy(ctx, r.providerData.Client, state.ID.ValueString(), service.BoolValue(state.ForceDestroy), service.BoolValue(state.Reuse)); err != nil {
		resp.Diagnostics.AddError("Error deleting Fastly CDN service", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
