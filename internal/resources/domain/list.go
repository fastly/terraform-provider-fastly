package domain

import (
	"context"
	"fmt"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/listidentity"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	resp.TypeName = req.ProviderTypeName + "_service_domain"
}

func (l *ListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "List all domains across all Fastly CDN and Compute services at their active version, or latest version when no active version exists.",
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
			if svc == nil || svc.Type == nil || !service.TypeSupported(*svc.Type, service.TypeVCL, service.TypeCompute) {
				continue
			}
			serviceID := fastly.ToValue(svc.ServiceID)
			if serviceID == "" {
				continue
			}

			version, _, err := service.SelectReadVersionFromServiceSummary(ctx, l.client, svc)
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

				result := listidentity.NewResult(ctx, req)
				result.DisplayName = service.ToGeneratedResourceName(fastly.ToValue(svc.Name), serviceID, *d.Name)

				if req.IncludeResource {
					result.Diagnostics.Append(setResourceAttrs(ctx, &result, d, serviceID, version)...)
				}

				if !push(result) {
					return
				}
			}
		}
	}
}

func setResourceAttrs(ctx context.Context, result *list.ListResult, d *fastly.Domain, serviceID string, version int) diag.Diagnostics {
	var diags diag.Diagnostics

	id := serviceID + "-" + fmt.Sprintf("%d", version) + "-" + fastly.ToValue(d.Name)

	model := Model{
		ID:      types.StringValue(id),
		Service: types.StringValue(serviceID),
		Version: types.Int64Value(int64(version)),
		Name:    types.StringValue(fastly.ToValue(d.Name)),
	}

	if d.Comment != nil && *d.Comment != "" {
		model.Comment = types.StringValue(*d.Comment)
	} else {
		model.Comment = types.StringNull()
	}

	diags.Append(result.Resource.Set(ctx, &model)...)
	return diags
}
