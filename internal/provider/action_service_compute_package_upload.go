package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type serviceComputePackageUploadAction struct {
	providerData *providerData
}

var _ action.Action = &serviceComputePackageUploadAction{}

func NewServiceComputePackageUploadAction() action.Action {
	return &serviceComputePackageUploadAction{}
}

type serviceComputePackageUploadModel struct {
	ServiceID types.String `tfsdk:"service_id"`
	Version   types.Int64  `tfsdk:"version"`
	Content   types.String `tfsdk:"content"`
	Filename  types.String `tfsdk:"filename"`
}

func (a *serviceComputePackageUploadAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_compute_package_upload"
}

func (a *serviceComputePackageUploadAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Uploads or replaces a Compute package on a specific Fastly service version. Intended for explicit lifecycle workflows.",
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "Fastly Compute service ID.",
			},
			"version": schema.Int64Attribute{
				Required:    true,
				Description: "Writable Fastly service version to modify.",
			},
			"content": schema.StringAttribute{
				Optional:    true,
				Description: "The contents of the Compute deployment package as a base64-encoded string. Conflicts with `filename`.",
			},
			"filename": schema.StringAttribute{
				Optional:    true,
				Description: "The path to the Compute deployment package on the local filesystem. Conflicts with `content`.",
			},
		},
	}
}

func (a *serviceComputePackageUploadAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	a.providerData = providerData
}

func (a *serviceComputePackageUploadAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg serviceComputePackageUploadModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := cfg.ServiceID.ValueString()
	version := int(cfg.Version.ValueInt64())

	if err := ensureServiceTypeSupported(ctx, a.providerData.client, serviceID, "fastly_service_compute_package_upload", serviceTypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	resp.Diagnostics.Append(a.providerData.ensureVersionMutable(ctx, serviceID, version)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pkg := []serviceComputePackageModel{{
		Content:  cfg.Content,
		Filename: cfg.Filename,
	}}

	tflog.Info(ctx, "Uploading Fastly Compute package", map[string]any{
		"service_id": serviceID,
		"version":    version,
	})

	if err := updateComputePackage(ctx, a.providerData.client, serviceID, version, pkg); err != nil {
		resp.Diagnostics.AddError("Failed to upload Fastly Compute package", err.Error())
		return
	}
}
