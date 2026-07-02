package serviceversion

import (
	"context"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DataSource{}

type DataSource struct {
	client *fastly.Client
}

type DataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ServiceID     types.String `tfsdk:"service_id"`
	LatestVersion types.Int64  `tfsdk:"latest_version"`
	ActiveVersion types.Int64  `tfsdk:"active_version"`
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_version"
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read-only view of the latest and active Fastly service versions for a service. This data source is for observability only and never creates, clones, activates, or mutates versions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform data source identifier. Mirrors service_id.",
			},
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "Fastly service ID.",
			},
			"latest_version": schema.Int64Attribute{
				Computed:    true,
				Description: "Highest version number that exists for the service.",
			},
			"active_version": schema.Int64Attribute{
				Computed:    true,
				Description: "Currently active production version, if any.",
			},
		},
	}
}

func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}

	d.client = data.Client
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()

	tflog.Debug(ctx, "Reading Fastly service version summary", map[string]any{
		"service_id": serviceID,
	})

	latestVersion, err := d.client.LatestVersion(ctx, &fastly.LatestVersionInput{
		ServiceID: serviceID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading latest Fastly service version", err.Error())
		return
	}

	activeServiceDetails, err := d.client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: serviceID,
		Filters: []fastly.ServiceDetailsFilter{
			{Key: "versions.active", Value: true},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading active Fastly service version", err.Error())
		return
	}

	state.ID = types.StringValue(serviceID)
	state.LatestVersion = versionNumberValue(latestVersion)
	state.ActiveVersion = types.Int64Null()

	if activeServiceDetails.ActiveVersion != nil {
		state.ActiveVersion = versionNumberValue(activeServiceDetails.ActiveVersion)
	} else if activeServiceDetails.Version != nil {
		state.ActiveVersion = versionNumberValue(activeServiceDetails.Version)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func versionNumberValue(version *fastly.Version) types.Int64 {
	if version == nil || version.Number == nil {
		return types.Int64Null()
	}

	return types.Int64Value(int64(*version.Number))
}
