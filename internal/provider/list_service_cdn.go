package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var nonIdentifierRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// toIdentifier converts a service name to a valid HCL identifier.
func toIdentifier(name string) string {
	id := nonIdentifierRe.ReplaceAllString(name, "_")
	id = strings.ToLower(id)
	if len(id) > 0 && id[0] >= '0' && id[0] <= '9' {
		id = "_" + id
	}
	return id
}

func toGeneratedResourceName(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}

	id := toIdentifier(strings.Join(nonEmpty, "_"))
	if id == "" || id == "_" {
		return "resource"
	}
	return id
}

var _ list.ListResource = &serviceCDNListResource{}
var _ list.ListResourceWithConfigure = &serviceCDNListResource{}

type serviceCDNListResource struct {
	client *fastly.Client
}

func NewServiceCDNListResource() list.ListResource {
	return &serviceCDNListResource{}
}

func (l *serviceCDNListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_cdn"
}

func (l *serviceCDNListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "List all Fastly CDN services accessible to the API token.",
		Attributes:  map[string]listschema.Attribute{},
	}
}

func (l *serviceCDNListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (l *serviceCDNListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
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
			if svc == nil || svc.Type == nil || *svc.Type != serviceTypeVCL {
				continue
			}
			if req.Limit > 0 && count >= req.Limit {
				return
			}
			count++

			result := req.NewListResult(ctx)
			result.DisplayName = toGeneratedResourceName(fastly.ToValue(svc.Name), fastly.ToValue(svc.ServiceID))

			if svc.ServiceID != nil {
				result.Diagnostics.Append(
					result.Identity.SetAttribute(ctx, path.Root("service_id"), *svc.ServiceID)...,
				)
			}

			if req.IncludeResource {
				result.Diagnostics.Append(setServiceCDNResourceAttrs(ctx, &result, svc)...)
			}

			if !push(result) {
				return
			}
		}
	}
}

func setServiceCDNResourceAttrs(ctx context.Context, result *list.ListResult, svc *fastly.Service) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(result.Resource.SetAttribute(ctx, path.Root("id"), fastly.ToValue(svc.ServiceID))...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("name"), fastly.ToValue(svc.Name))...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("comment"), fastly.ToValue(svc.Comment))...)

	return diags
}
