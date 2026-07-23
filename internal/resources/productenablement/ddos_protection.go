package productenablement

import (
	"context"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/errors"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/go-fastly/v16/fastly/products/ddosprotection"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DDoSProtectionResource{}
var _ resource.ResourceWithImportState = &DDoSProtectionResource{}

type DDoSProtectionModel struct {
	ID        types.String `tfsdk:"id"`
	ServiceID types.String `tfsdk:"service_id"`
	Mode      types.String `tfsdk:"mode"`
}

type DDoSProtectionResource struct {
	client *fastly.Client
}

func NewDDoSProtectionResource() resource.Resource {
	return &DDoSProtectionResource{}
}

func (r *DDoSProtectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_product_ddos_protection"
}

func (r *DDoSProtectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enables DDoS Protection on a service. Product Enablement operates on the service directly rather than a specific service version, so this resource is not tied to a `version` and applies immediately.",
		Attributes: map[string]schema.Attribute{
			"id":         idAttribute(),
			"service_id": serviceIDAttribute("DDoS Protection"),
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "Operation mode. Can be `off`, `log`, or `block`.",
				Validators: []validator.String{
					stringvalidator.OneOf("off", "log", "block"),
				},
			},
		},
	}
}

func (r *DDoSProtectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}
	r.client = data.Client
}

func (r *DDoSProtectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DDoSProtectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	tflog.Debug(ctx, "Creating Fastly Product Enablement (ddos_protection)", map[string]any{"service_id": serviceID})

	if _, err := ddosprotection.Enable(ctx, r.client, serviceID, ddosprotection.EnableInput{
		Mode: plan.Mode.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Error enabling ddos_protection", err.Error())
		return
	}

	plan.ID = plan.ServiceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DDoSProtectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DDoSProtectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	tflog.Debug(ctx, "Reading Fastly Product Enablement (ddos_protection)", map[string]any{"service_id": serviceID})

	if _, err := ddosprotection.Get(ctx, r.client, serviceID); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	cfg, err := ddosprotection.GetConfiguration(ctx, r.client, serviceID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ddos_protection configuration", err.Error())
		return
	}
	state.Mode = types.StringPointerValue(cfg.Configuration.Mode)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DDoSProtectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DDoSProtectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	tflog.Debug(ctx, "Updating Fastly Product Enablement (ddos_protection)", map[string]any{"service_id": serviceID})

	if _, err := ddosprotection.UpdateConfiguration(ctx, r.client, serviceID, ddosprotection.ConfigureInput{
		Mode: plan.Mode.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Error configuring ddos_protection", err.Error())
		return
	}

	plan.ID = plan.ServiceID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DDoSProtectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DDoSProtectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	tflog.Debug(ctx, "Deleting Fastly Product Enablement (ddos_protection)", map[string]any{"service_id": serviceID})

	if err := ddosprotection.Disable(ctx, r.client, serviceID); err != nil && !isEntitlementError(err) && !errors.IsNotFound(err) {
		resp.Diagnostics.AddError("Error disabling ddos_protection", err.Error())
	}
}

func (r *DDoSProtectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("service_id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
