package versionactivate

import (
	"context"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Action struct {
	client *fastly.Client
}

var _ action.Action = &Action{}

func NewAction() action.Action {
	return &Action{}
}

type Model struct {
	ServiceID types.String `tfsdk:"service_id"`
	Version   types.Int64  `tfsdk:"version"`
}

func (a *Action) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_version_activate"
}

func (a *Action) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Activates a Fastly service version in production. Implemented as a Terraform Action (non-CRUD).",
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "Fastly service ID.",
			},
			"version": schema.Int64Attribute{
				Required:    true,
				Description: "Service version number to activate in production.",
			},
		},
	}
}

func (a *Action) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}

	a.client = data.Client
}

func (a *Action) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := cfg.ServiceID.ValueString()
	version := int(cfg.Version.ValueInt64())

	tflog.Info(ctx, "Activating Fastly service version in production", map[string]any{
		"service_id": serviceID,
		"version":    version,
	})

	_, err := a.client.ActivateVersion(ctx, &fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to activate Fastly service version", err.Error())
		return
	}

	tflog.Info(ctx, "Activated Fastly service version successfully", map[string]any{
		"service_id": serviceID,
		"version":    version,
	})
}
