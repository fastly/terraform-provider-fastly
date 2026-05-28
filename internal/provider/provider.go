package provider

import (
	"context"
	"os"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-fastly-dual-model-poc/internal/actions/computepackageupload"
	"terraform-provider-fastly-dual-model-poc/internal/actions/versionactivate"
	"terraform-provider-fastly-dual-model-poc/internal/actions/versionclone"
	"terraform-provider-fastly-dual-model-poc/internal/actions/versionstage"
	fastlyclient "terraform-provider-fastly-dual-model-poc/internal/client"
	"terraform-provider-fastly-dual-model-poc/internal/datasources/serviceversion"
	"terraform-provider-fastly-dual-model-poc/internal/resources/backend"
	"terraform-provider-fastly-dual-model-poc/internal/resources/domain"
	"terraform-provider-fastly-dual-model-poc/internal/resources/servicecdn"
	"terraform-provider-fastly-dual-model-poc/internal/resources/servicecdnauto"
	"terraform-provider-fastly-dual-model-poc/internal/resources/servicecompute"
	"terraform-provider-fastly-dual-model-poc/internal/resources/servicecomputeauto"
)

type fastlyProvider struct{}

type fastlyProviderModel struct {
	APIToken types.String `tfsdk:"api_token"`
}

func New() provider.Provider {
	return &fastlyProvider{}
}

func (p *fastlyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fastly"
}

func (p *fastlyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Fastly API token. Can also be set via the FASTLY_API_TOKEN environment variable.",
			},
		},
	}
}

func (p *fastlyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config fastlyProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiToken := os.Getenv("FASTLY_API_TOKEN")
	if !config.APIToken.IsNull() && config.APIToken.ValueString() != "" {
		apiToken = config.APIToken.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"An API token must be provided via the `api_token` provider configuration or FASTLY_API_TOKEN environment variable.",
		)
		return
	}

	client, err := fastly.NewClient(apiToken)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Fastly client", err.Error())
		return
	}

	data := fastlyclient.NewData(client)

	resp.ResourceData = data
	resp.DataSourceData = data
	resp.ActionData = data
	resp.ListResourceData = data
}

func (p *fastlyProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		backend.NewResource,
		domain.NewResource,
		servicecompute.NewResource,
		servicecomputeauto.NewResource,
		servicecdn.NewResource,
		servicecdnauto.NewResource,
	}
}

func (p *fastlyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		serviceversion.NewDataSource,
	}
}

func (p *fastlyProvider) ListResources(_ context.Context) []func() list.ListResource {
	return []func() list.ListResource{
		backend.NewListResource,
		domain.NewListResource,
		servicecompute.NewListResource,
		servicecdn.NewListResource,
	}
}

func (p *fastlyProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
		versionclone.NewAction,
		versionactivate.NewAction,
		versionstage.NewAction,
		computepackageupload.NewAction,
	}
}
