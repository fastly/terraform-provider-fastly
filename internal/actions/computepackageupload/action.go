package computepackageupload

import (
	"context"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/computepackage"
	"github.com/fastly/terraform-provider-fastly/internal/service"
	"github.com/fastly/terraform-provider-fastly/internal/validation"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Action struct {
	providerData *fastlyclient.Data
}

var _ action.Action = &Action{}

func NewAction() action.Action {
	return &Action{}
}

type Model struct {
	ServiceID types.String `tfsdk:"service_id"`
	Version   types.Int64  `tfsdk:"version"`
	Content   types.String `tfsdk:"content"`
	Filename  types.String `tfsdk:"filename"`
}

func (a *Action) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_compute_package_upload"
}

func (a *Action) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Uploads or replaces a Compute package on a specific Fastly service version. Intended for explicit/default workflows; not for use with automatic versioned service resources.",
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
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("filename"),
					),
				},
			},
			"filename": schema.StringAttribute{
				Optional:    true,
				Description: "The path to the Compute deployment package on the local filesystem. Conflicts with `content`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("content"),
					),
				},
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

	a.providerData = data
}

func (a *Action) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := cfg.ServiceID.ValueString()
	version := int(cfg.Version.ValueInt64())

	pkg := []computepackage.Model{{
		Content:  cfg.Content,
		Filename: cfg.Filename,
	}}

	if err := computepackage.ValidateInput(pkg); err != nil {
		resp.Diagnostics.AddError("Invalid Fastly Compute package configuration", err.Error())
		return
	}

	if err := validation.EnsureServiceTypeSupported(ctx, a.providerData.ServiceTypeChecker, serviceID, "fastly_service_compute_package_upload", service.TypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	resp.Diagnostics.Append(a.providerData.VersionChecker.EnsureMutable(ctx, serviceID, version)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Uploading Fastly Compute package", map[string]any{
		"service_id": serviceID,
		"version":    version,
	})

	if err := computepackage.Update(ctx, a.providerData.Client, serviceID, version, pkg); err != nil {
		resp.Diagnostics.AddError("Failed to upload Fastly Compute package", err.Error())
		return
	}
}
