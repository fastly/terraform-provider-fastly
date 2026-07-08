package cdnaclentries

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
	resp.TypeName = req.ProviderTypeName + "_service_cdn_acl_entries"
}

func (l *ListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "List ACL entries for all ACLs across all Fastly CDN services at their active version, or latest version when no active version exists.",
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
	tflog.Debug(ctx, "Listing Fastly service CDN ACL entries")

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
			if svc == nil || svc.Type == nil || !service.TypeSupported(*svc.Type, service.TypeVCL) {
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

			acls, err := l.client.ListACLs(ctx, &fastly.ListACLsInput{
				ServiceID:      serviceID,
				ServiceVersion: version,
			})
			if err != nil {
				tflog.Warn(ctx, "Error listing ACLs for service", map[string]any{
					"service_id": serviceID,
					"error":      err.Error(),
				})
				continue
			}

			for _, a := range acls {
				if a == nil || a.ACLID == nil || a.Name == nil {
					continue
				}
				if req.Limit > 0 && count >= req.Limit {
					return
				}
				count++

				aclID := fastly.ToValue(a.ACLID)

				result := listidentity.NewResult(ctx, req)
				result.DisplayName = service.ToGeneratedResourceName(fastly.ToValue(svc.Name), serviceID, *a.Name)

				if req.IncludeResource {
					result.Diagnostics.Append(setListResourceAttrs(ctx, l.client, &result, serviceID, aclID)...)
				}

				if !push(result) {
					return
				}
			}
		}
	}
}

func setListResourceAttrs(ctx context.Context, client *fastly.Client, result *list.ListResult, serviceID, aclID string) diag.Diagnostics {
	var diags diag.Diagnostics

	paginator := client.GetACLEntries(ctx, &fastly.GetACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})

	var remoteEntries []*fastly.ACLEntry
	for paginator.HasNext() {
		page, err := paginator.GetNext()
		if err != nil {
			diags.AddError("Error reading ACL entries", fmt.Sprintf("service %s, ACL %s: %s", serviceID, aclID, err))
			return diags
		}
		remoteEntries = append(remoteEntries, page...)
	}

	entrySet := flattenEntries(ctx, remoteEntries, types.SetNull(types.ObjectType{AttrTypes: entryAttrTypes}), &diags)
	if diags.HasError() {
		return diags
	}

	model := Model{
		ID:            types.StringValue(fmt.Sprintf("%s/%s", serviceID, aclID)),
		ServiceID:     types.StringValue(serviceID),
		ACLID:         types.StringValue(aclID),
		Entry:         entrySet,
		ManageEntries: types.BoolValue(false),
	}
	diags.Append(result.Resource.Set(ctx, &model)...)
	return diags
}
