package servicecdn

import (
	"context"

	fastlyclient "terraform-provider-fastly-dual-model-poc/internal/client"
	"terraform-provider-fastly-dual-model-poc/internal/service"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ list.ListResource = &ListResource{}
var _ list.ListResourceWithConfigure = &ListResource{}

type ListResource struct {
	client *fastly.Client
}

func NewListResource() list.ListResource {
	return &ListResource{}
}

func (l *ListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_cdn"
}

func (l *ListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "List all Fastly CDN services accessible to the API token.",
		Attributes:  map[string]listschema.Attribute{},
	}
}

func (l *ListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}

	l.client = data.Client
}

func (l *ListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	tflog.Debug(ctx, "Listing Fastly CDN services")

	services, err := l.client.ListServices(ctx, &fastly.ListServicesInput{})
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Error listing Fastly services", err.Error()),
		})
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		var count int64
		for _, svc := range services {
			if svc == nil || svc.Type == nil || *svc.Type != service.TypeVCL {
				continue
			}
			if req.Limit > 0 && count >= req.Limit {
				return
			}
			count++

			result := req.NewListResult(ctx)
			result.DisplayName = service.ToGeneratedResourceName(fastly.ToValue(svc.Name), fastly.ToValue(svc.ServiceID))

			if svc.ServiceID != nil {
				result.Diagnostics.Append(
					result.Identity.SetAttribute(ctx, path.Root("service_id"), *svc.ServiceID)...,
				)
			}

			if req.IncludeResource {
				result.Diagnostics.Append(setResourceAttrs(ctx, &result, svc)...)
			}

			if !push(result) {
				return
			}
		}
	}
}

func setResourceAttrs(ctx context.Context, result *list.ListResult, svc *fastly.Service) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(result.Resource.SetAttribute(ctx, path.Root("id"), fastly.ToValue(svc.ServiceID))...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("name"), fastly.ToValue(svc.Name))...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("comment"), fastly.ToValue(svc.Comment))...)

	return diags
}
