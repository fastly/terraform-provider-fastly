package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/fastly/terraform-provider-fastly/internal/actions/computepackageupload"
	"github.com/fastly/terraform-provider-fastly/internal/actions/versionactivate"
	"github.com/fastly/terraform-provider-fastly/internal/actions/versionclone"
	"github.com/fastly/terraform-provider-fastly/internal/actions/versionstage"
	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/datasources/acls"
	"github.com/fastly/terraform-provider-fastly/internal/datasources/kvstores"
	"github.com/fastly/terraform-provider-fastly/internal/datasources/serviceversion"
	"github.com/fastly/terraform-provider-fastly/internal/resources/acl"
	"github.com/fastly/terraform-provider-fastly/internal/resources/aclentries"
	"github.com/fastly/terraform-provider-fastly/internal/resources/backend"
	"github.com/fastly/terraform-provider-fastly/internal/resources/cdnacl"
	"github.com/fastly/terraform-provider-fastly/internal/resources/cdnaclentries"
	"github.com/fastly/terraform-provider-fastly/internal/resources/domain"
	"github.com/fastly/terraform-provider-fastly/internal/resources/kvstore"
	"github.com/fastly/terraform-provider-fastly/internal/resources/loggings3"
	"github.com/fastly/terraform-provider-fastly/internal/resources/productenablement"
	"github.com/fastly/terraform-provider-fastly/internal/resources/resourcelink"
	"github.com/fastly/terraform-provider-fastly/internal/resources/servicecdn"
	"github.com/fastly/terraform-provider-fastly/internal/resources/servicecdnauto"
	"github.com/fastly/terraform-provider-fastly/internal/resources/servicecompute"
	"github.com/fastly/terraform-provider-fastly/internal/resources/servicecomputeauto"
	"github.com/fastly/terraform-provider-fastly/internal/version"
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

	userAgentPrefix := fmt.Sprintf("terraform-provider-fastly/%s", version.Version)
	data := fastlyclient.NewData(client, userAgentPrefix)

	resp.ResourceData = data
	resp.DataSourceData = data
	resp.ActionData = data
	resp.ListResourceData = data
}

func (p *fastlyProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		acl.NewResource,
		aclentries.NewResource,
		backend.NewResource,
		cdnacl.NewResource,
		cdnaclentries.NewResource,
		domain.NewResource,
		loggings3.NewResource,
		kvstore.NewResource,
		productenablement.NewFanoutResource,
		productenablement.NewBrotliCompressionResource,
		productenablement.NewImageOptimizerResource,
		productenablement.NewOriginInspectorResource,
		productenablement.NewDomainInspectorResource,
		productenablement.NewWebsocketsResource,
		productenablement.NewLogExplorerInsightsResource,
		productenablement.NewAPIDiscoveryResource,
		productenablement.NewBotManagementResource,
		productenablement.NewDDoSProtectionResource,
		productenablement.NewNGWAFResource,
		resourcelink.NewResource,
		servicecdn.NewResource,
		servicecdnauto.NewResource,
		servicecompute.NewResource,
		servicecomputeauto.NewResource,
	}
}

func (p *fastlyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		acls.NewDataSource,
		kvstores.NewDataSource,
		serviceversion.NewDataSource,
	}
}

func (p *fastlyProvider) ListResources(_ context.Context) []func() list.ListResource {
	return []func() list.ListResource{
		backend.NewListResource,
		cdnacl.NewListResource,
		cdnaclentries.NewListResource,
		domain.NewListResource,
		loggings3.NewListResource,
		servicecdn.NewListResource,
		servicecompute.NewListResource,
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
