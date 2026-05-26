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

type serviceVersionCloneAction struct {
	client *fastly.Client
}

var _ action.Action = &serviceVersionCloneAction{}

func NewServiceVersionCloneAction() action.Action {
	return &serviceVersionCloneAction{}
}

type serviceVersionCloneModel struct {
	ServiceID types.String `tfsdk:"service_id"`
	Version   types.Int64  `tfsdk:"version"`
}

func (a *serviceVersionCloneAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_version_clone"
}

func (a *serviceVersionCloneAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Clones a Fastly service version. Implemented as a Terraform Action (non-CRUD).",
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "Fastly service ID.",
			},
			"version": schema.Int64Attribute{
				Required:    true,
				Description: "Service version number to clone.",
			},
		},
	}
}

func (a *serviceVersionCloneAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *serviceVersionCloneAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg serviceVersionCloneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := cfg.ServiceID.ValueString()
	version := int(cfg.Version.ValueInt64())

	tflog.Info(ctx, "Cloning Fastly service version", map[string]any{
		"service_id": serviceID,
		"version":    version,
	})

	v, err := a.client.CloneVersion(ctx, &fastly.CloneVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to clone Fastly service version", err.Error())
		return
	}

	newVersion := 0
	if v != nil && v.Number != nil {
		newVersion = *v.Number
	}

	tflog.Info(ctx, "Cloned Fastly service version successfully", map[string]any{
		"service_id":   serviceID,
		"from_version": version,
		"new_version":  newVersion,
	})

	// NOTE: Actions currently don't feed results into Terraform graph/state in a durable way.
	// Users should copy the logged new version into terraform.tfvars (or their pipeline vars).
}
