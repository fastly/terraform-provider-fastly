package provider

import (
	"context"
	"fmt"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ list.ListResource = &serviceDomainListResource{}
var _ list.ListResourceWithConfigure = &serviceDomainListResource{}

type serviceDomainListResource struct {
	client *fastly.Client
}

func NewServiceDomainListResource() list.ListResource {
	return &serviceDomainListResource{}
}

func (l *serviceDomainListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_domain"
}

func (l *serviceDomainListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "List all domains across all Fastly CDN and Compute services at their active version, or latest version when no active version exists.",
		Attributes:  map[string]listschema.Attribute{},
	}
}

func (l *serviceDomainListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	l.client = providerData.client
}

func (l *serviceDomainListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	tflog.Debug(ctx, "Listing Fastly service domains ")

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
			if svc == nil || svc.Type == nil || !serviceTypeSupported(*svc.Type, serviceTypeVCL, serviceTypeCompute) {
				continue
			}
			serviceID := fastly.ToValue(svc.ServiceID)
			if serviceID == "" {
				continue
			}

			version, _, err := selectServiceReadVersionFromServiceSummary(ctx, l.client, svc)
			if err != nil {
				tflog.Warn(ctx, "Error selecting service version for query", map[string]any{
					"service_id": serviceID,
					"error":      err.Error(),
				})
				continue
			}

			domains, err := l.client.ListDomains(ctx, &fastly.ListDomainsInput{
				ServiceID:      serviceID,
				ServiceVersion: version,
			})
			if err != nil {
				tflog.Warn(ctx, "Error listing domains for service", map[string]any{
					"service_id": serviceID,
					"error":      err.Error(),
				})
				continue
			}

			for _, d := range domains {
				if d == nil || d.Name == nil {
					continue
				}
				if req.Limit > 0 && count >= req.Limit {
					return
				}
				count++

				result := req.NewListResult(ctx)
				result.DisplayName = toGeneratedResourceName(fastly.ToValue(svc.Name), serviceID, *d.Name)

				result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
				result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("version"), int64(version))...)
				result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("name"), *d.Name)...)

				if req.IncludeResource {
					result.Diagnostics.Append(setServiceDomainResourceAttrs(ctx, &result, d, serviceID, version)...)
				}

				if !push(result) {
					return
				}
			}
		}
	}
}

func setServiceDomainResourceAttrs(ctx context.Context, result *list.ListResult, d *fastly.Domain, serviceID string, version int) diag.Diagnostics {
	var diags diag.Diagnostics

	id := serviceID + "-" + fmt.Sprintf("%d", version) + "-" + fastly.ToValue(d.Name)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("id"), id)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("version"), int64(version))...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("name"), fastly.ToValue(d.Name))...)

	if d.Comment != nil && *d.Comment != "" {
		diags.Append(result.Resource.SetAttribute(ctx, path.Root("comment"), *d.Comment)...)
	} else {
		diags.Append(result.Resource.SetAttribute(ctx, path.Root("comment"), types.StringNull())...)
	}

	return diags
}
