package provider

import (
	"context"
	"fmt"

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

type serviceCDNResource struct {
	client *fastly.Client
}

var _ resource.Resource = &serviceCDNResource{}
var _ resource.ResourceWithImportState = &serviceCDNResource{}
var _ resource.ResourceWithIdentity = &serviceCDNResource{}

func NewServiceCDNResource() resource.Resource {
	return &serviceCDNResource{}
}

type serviceCDNModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Comment      types.String `tfsdk:"comment"`
	ForceDestroy types.Bool   `tfsdk:"force_destroy"`
	Reuse        types.Bool   `tfsdk:"reuse"`
}

func (r *serviceCDNResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_cdn"
}

func (r *serviceCDNResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *serviceCDNResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected ProviderData type",
			fmt.Sprintf("Expected *providerData, got: %T", req.ProviderData),
		)
		return
	}

	r.client = providerData.client
}

func (r *serviceCDNResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceCDNModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Fastly CDN service", map[string]any{
		"name":    plan.Name.ValueString(),
		"comment": plan.Comment.ValueString(),
	})

	service, err := r.client.CreateService(ctx, &fastly.CreateServiceInput{
		Name:    fastly.ToPointer(plan.Name.ValueString()),
		Comment: fastly.ToPointer(plan.Comment.ValueString()),
		Type:    fastly.ToPointer(serviceTypeVCL),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Fastly CDN service", err.Error())
		return
	}

	if service.ServiceID == nil {
		resp.Diagnostics.AddError("Error creating Fastly CDN service", "Fastly API returned nil service ID.")
		return
	}

	plan.ID = types.StringValue(*service.ServiceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceCDNResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceCDNModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := r.client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: state.ID.ValueString(),
	})
	if err != nil {
		if httpErr, ok := err.(*fastly.HTTPError); ok && httpErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Fastly CDN service", err.Error())
		return
	}

	serviceType := fastly.ToValue(service.Type)
	if serviceType != serviceTypeVCL {
		resp.Diagnostics.AddError(
			"Unexpected Fastly service type",
			fmt.Sprintf("Expected VCL service %q to have type %q, got %q.", state.ID.ValueString(), serviceTypeVCL, serviceType),
		)
		return
	}

	if service.ServiceID != nil {
		state.ID = types.StringValue(*service.ServiceID)
	}
	if service.Name != nil {
		state.Name = types.StringValue(*service.Name)
	}
	if service.Comment != nil {
		state.Comment = types.StringValue(*service.Comment)
	} else {
		state.Comment = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceCDNResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceCDNModel
	var state serviceCDNModel

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
		resp.Diagnostics.AddError("Error updating Fastly CDN service", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceCDNResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceCDNModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := deleteServiceWithPolicy(ctx, r.client, state.ID.ValueString(), boolValue(state.ForceDestroy), boolValue(state.Reuse)); err != nil {
		resp.Diagnostics.AddError("Error deleting Fastly CDN service", err.Error())
	}
}

func (r *serviceCDNResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("service_id"), req, resp)
}

func (r *serviceCDNResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"service_id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Fastly service ID.",
			},
		},
	}
}
