package provider

import (
	"context"
	"fmt"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type serviceVersionActivateAction struct {
	client *fastly.Client
}

var _ action.Action = &serviceVersionActivateAction{}

func NewServiceVersionActivateAction() action.Action {
	return &serviceVersionActivateAction{}
}

type serviceVersionActivateModel struct {
	ServiceID types.String `tfsdk:"service_id"`
	Version   types.Int64  `tfsdk:"version"`
	Staging   types.Bool   `tfsdk:"staging"`
}

func (a *serviceVersionActivateAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_version_activate"
}

func (a *serviceVersionActivateAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Activates a Fastly service version. Implemented as a Terraform Action (non-CRUD).",
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "Fastly service ID.",
			},
			"version": schema.Int64Attribute{
				Required:    true,
				Description: "Service version number to activate.",
			},
			"staging": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, activate on staging. If false/omitted, activate on production.",
			},
		},
	}
}

func (a *serviceVersionActivateAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	a.client = providerData.client
}

func (a *serviceVersionActivateAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg serviceVersionActivateModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := cfg.ServiceID.ValueString()
	version := int(cfg.Version.ValueInt64())

	staging := false
	if !cfg.Staging.IsNull() {
		staging = cfg.Staging.ValueBool()
	}

	env := ""
	if staging {
		env = "staging"
	}

	tflog.Info(ctx, "Activating Fastly service version", map[string]any{
		"service_id":   serviceID,
		"version":      version,
		"environment":  env,
		"staging_bool": staging,
	})

	_, err := a.client.ActivateVersion(ctx, &fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Environment:    env, // "" for prod, "staging" for staging
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to activate Fastly service version", err.Error())
		return
	}

	tflog.Info(ctx, "Activated Fastly service version successfully", map[string]any{
		"service_id":   serviceID,
		"version":      version,
		"environment":  env,
		"staging_bool": staging,
	})
}
