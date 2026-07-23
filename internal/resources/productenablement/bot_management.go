package productenablement

import (
	"context"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/go-fastly/v16/fastly/products/botmanagement"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &BotManagementResource{}
var _ resource.ResourceWithImportState = &BotManagementResource{}

type BotManagementModel struct {
	ID           types.String `tfsdk:"id"`
	ServiceID    types.String `tfsdk:"service_id"`
	ContentGuard types.String `tfsdk:"contentguard"`
}

type BotManagementResource struct {
	client *fastly.Client
}

func NewBotManagementResource() resource.Resource {
	return &BotManagementResource{}
}

func (r *BotManagementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_product_bot_management"
}

func (r *BotManagementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enables Bot Management on a service. Product Enablement operates on the service directly rather than a specific service version, so this resource is not tied to a `version` and applies immediately.",
		Attributes: map[string]schema.Attribute{
			"id":         idAttribute(),
			"service_id": serviceIDAttribute("Bot Management"),
			"contentguard": schema.StringAttribute{
				Required:    true,
				Description: "ContentGuard status. Can be either `off` or `on`.",
				Validators: []validator.String{
					stringvalidator.OneOf("off", "on"),
				},
			},
		},
	}
}

func (r *BotManagementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}
	r.client = data.Client
}

func (r *BotManagementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BotManagementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	tflog.Debug(ctx, "Creating Fastly Product Enablement (bot_management)", map[string]any{"service_id": serviceID})

	if _, err := botmanagement.Enable(ctx, r.client, serviceID); err != nil {
		resp.Diagnostics.AddError("Error enabling bot_management", err.Error())
		return
	}
	if _, err := botmanagement.UpdateConfiguration(ctx, r.client, serviceID, botmanagement.ConfigureInput{
		ContentGuard: plan.ContentGuard.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Error configuring bot_management", err.Error())
		return
	}

	plan.ID = plan.ServiceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BotManagementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BotManagementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	tflog.Debug(ctx, "Reading Fastly Product Enablement (bot_management)", map[string]any{"service_id": serviceID})

	if _, err := botmanagement.Get(ctx, r.client, serviceID); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	cfg, err := botmanagement.GetConfiguration(ctx, r.client, serviceID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading bot_management configuration", err.Error())
		return
	}
	state.ContentGuard = types.StringPointerValue(cfg.Configuration.ContentGuard)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BotManagementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BotManagementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	tflog.Debug(ctx, "Updating Fastly Product Enablement (bot_management)", map[string]any{"service_id": serviceID})

	if _, err := botmanagement.UpdateConfiguration(ctx, r.client, serviceID, botmanagement.ConfigureInput{
		ContentGuard: plan.ContentGuard.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Error configuring bot_management", err.Error())
		return
	}

	plan.ID = plan.ServiceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BotManagementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BotManagementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	tflog.Debug(ctx, "Deleting Fastly Product Enablement (bot_management)", map[string]any{"service_id": serviceID})

	if err := botmanagement.Disable(ctx, r.client, serviceID); err != nil && !isEntitlementError(err) && !errors.IsNotFound(err) {
		resp.Diagnostics.AddError("Error disabling bot_management", err.Error())
	}
}

func (r *BotManagementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("service_id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
