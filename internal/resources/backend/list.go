package backend

import (
	"context"
	"fmt"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/service"

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
	resp.TypeName = req.ProviderTypeName + "_service_backend"
}

func (l *ListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "List all backends across all Fastly CDN and Compute services at their active version, or latest version when no active version exists.",
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
	tflog.Debug(ctx, "Listing Fastly service backends ")

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

			backends, err := l.client.ListBackends(ctx, &fastly.ListBackendsInput{
				ServiceID:      serviceID,
				ServiceVersion: version,
			})
			if err != nil {
				tflog.Warn(ctx, "Error listing backends for service", map[string]any{
					"service_id": serviceID,
					"error":      err.Error(),
				})
				continue
			}

			for _, b := range backends {
				if b == nil || b.Name == nil {
					continue
				}
				if req.Limit > 0 && count >= req.Limit {
					return
				}
				count++

				result := req.NewListResult(ctx)
				result.DisplayName = service.ToGeneratedResourceName(fastly.ToValue(svc.Name), serviceID, *b.Name)

				result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
				result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("version"), int64(version))...)
				result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("name"), *b.Name)...)

				if req.IncludeResource {
					result.Diagnostics.Append(setResourceAttrs(ctx, &result, b, serviceID, version)...)
				}

				if !push(result) {
					return
				}
			}
		}
	}
}

func setResourceAttrs(ctx context.Context, result *list.ListResult, b *fastly.Backend, serviceID string, version int) diag.Diagnostics {
	var diags diag.Diagnostics

	id := serviceID + "-" + fmt.Sprintf("%d", version) + "-" + fastly.ToValue(b.Name)
	model := FlattenToNestedModel(b)

	diags.Append(result.Resource.SetAttribute(ctx, path.Root("id"), id)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("version"), int64(version))...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("name"), model.Name)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("address"), model.Address)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("port"), model.Port)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("comment"), model.Comment)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("auto_loadbalance"), model.AutoLoadbalance)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("between_bytes_timeout"), model.BetweenBytesTimeout)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("connect_timeout"), model.ConnectTimeout)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("error_threshold"), model.ErrorThreshold)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("first_byte_timeout"), model.FirstByteTimeout)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("healthcheck"), model.HealthCheck)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("keepalive_time"), model.KeepaliveTime)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("max_conn"), model.MaxConn)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("max_lifetime"), model.MaxLifetime)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("max_tls_version"), model.MaxTLSVersion)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("max_use"), model.MaxUse)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("min_tls_version"), model.MinTLSVersion)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("override_host"), model.OverrideHost)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("prefer_ipv6"), model.PreferIPv6)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("request_condition"), model.RequestCondition)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("share_key"), model.ShareKey)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("shield"), model.Shield)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("ssl_ca_cert"), model.SSLCACert)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("ssl_cert_hostname"), model.SSLCertHostname)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("ssl_check_cert"), model.SSLCheckCert)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("ssl_ciphers"), model.SSLCiphers)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("ssl_client_cert"), model.SSLClientCert)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("ssl_client_key"), model.SSLClientKey)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("ssl_sni_hostname"), model.SSLSNIHostname)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("use_ssl"), model.UseSSL)...)
	diags.Append(result.Resource.SetAttribute(ctx, path.Root("weight"), model.Weight)...)

	return diags
}
