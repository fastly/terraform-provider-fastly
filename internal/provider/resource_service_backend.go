package provider

import (
	"context"
	"strconv"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &serviceBackendResource{}
var _ resource.ResourceWithImportState = &serviceBackendResource{}
var _ resource.ResourceWithIdentity = &serviceBackendResource{}

type serviceBackendResource struct {
	providerData *providerData
}

func NewServiceBackendResource() resource.Resource {
	return &serviceBackendResource{}
}

type serviceBackendModel struct {
	ID                  types.String `tfsdk:"id"`
	Service             types.String `tfsdk:"service_id"`
	Version             types.Int64  `tfsdk:"version"`
	Name                types.String `tfsdk:"name"`
	Address             types.String `tfsdk:"address"`
	Port                types.Int64  `tfsdk:"port"`
	Comment             types.String `tfsdk:"comment"`
	AutoLoadbalance     types.Bool   `tfsdk:"auto_loadbalance"`
	BetweenBytesTimeout types.Int64  `tfsdk:"between_bytes_timeout"`
	ConnectTimeout      types.Int64  `tfsdk:"connect_timeout"`
	ErrorThreshold      types.Int64  `tfsdk:"error_threshold"`
	FirstByteTimeout    types.Int64  `tfsdk:"first_byte_timeout"`
	HealthCheck         types.String `tfsdk:"healthcheck"`
	KeepaliveTime       types.Int64  `tfsdk:"keepalive_time"`
	MaxConn             types.Int64  `tfsdk:"max_conn"`
	MaxLifetime         types.Int64  `tfsdk:"max_lifetime"`
	MaxTLSVersion       types.String `tfsdk:"max_tls_version"`
	MaxUse              types.Int64  `tfsdk:"max_use"`
	MinTLSVersion       types.String `tfsdk:"min_tls_version"`
	OverrideHost        types.String `tfsdk:"override_host"`
	PreferIPv6          types.Bool   `tfsdk:"prefer_ipv6"`
	RequestCondition    types.String `tfsdk:"request_condition"`
	ShareKey            types.String `tfsdk:"share_key"`
	Shield              types.String `tfsdk:"shield"`
	SSLCACert           types.String `tfsdk:"ssl_ca_cert"`
	SSLCertHostname     types.String `tfsdk:"ssl_cert_hostname"`
	SSLCheckCert        types.Bool   `tfsdk:"ssl_check_cert"`
	SSLCiphers          types.String `tfsdk:"ssl_ciphers"`
	SSLClientCert       types.String `tfsdk:"ssl_client_cert"`
	SSLClientKey        types.String `tfsdk:"ssl_client_key"`
	SSLSNIHostname      types.String `tfsdk:"ssl_sni_hostname"`
	UseSSL              types.Bool   `tfsdk:"use_ssl"`
	Weight              types.Int64  `tfsdk:"weight"`
}

func (r *serviceBackendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_backend"
}

func (r *serviceBackendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fastly service backend resource. Writes directly to the specified writable service version.",
		Attributes:  backendResourceAttributes(),
	}
}

func (r *serviceBackendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected ProviderData type", "Expected *providerData.")
		return
	}
	r.providerData = providerData
}

func (r *serviceBackendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceBackendModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := ensureServiceTypeSupported(ctx, r.providerData.client, plan.Service.ValueString(), "fastly_service_backend", serviceTypeVCL, serviceTypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	resp.Diagnostics.Append(r.providerData.ensureVersionMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := buildCreateBackendInput(plan.Service.ValueString(), int(plan.Version.ValueInt64()), explicitBackendToVCLModel(plan))
	tflog.Debug(ctx, "Creating Fastly service backend", map[string]any{
		"service_id": opts.ServiceID,
		"version":    opts.ServiceVersion,
		"name":       stringValue(plan.Name),
	})

	b, err := r.providerData.client.CreateBackend(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating explicit service backend", err.Error())
		return
	}

	flattenServiceBackend(ctx, b, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceBackendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceBackendModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Fastly service backend from API", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	b, err := r.providerData.client.GetBackend(ctx, &fastly.GetBackendInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if fastlyErr, ok := err.(*fastly.HTTPError); ok && fastlyErr.StatusCode == 404 {
			tflog.Warn(ctx, "Service backend not found, removing from state", map[string]any{
				"service_id": state.Service.ValueString(),
				"name":       state.Name.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading explicit service backend", err.Error())
		return
	}

	flattenServiceBackend(ctx, b, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceBackendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceBackendModel
	var state serviceBackendModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := ensureServiceTypeSupported(ctx, r.providerData.client, plan.Service.ValueString(), "fastly_service_backend", serviceTypeVCL, serviceTypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	resp.Diagnostics.Append(r.providerData.ensureVersionMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	forceAll := plan.Version.ValueInt64() != state.Version.ValueInt64()
	stateBackend := explicitBackendToVCLModel(state)
	opts := buildUpdateBackendInput(
		plan.Service.ValueString(),
		int(plan.Version.ValueInt64()),
		explicitBackendToVCLModel(plan),
		&stateBackend,
		forceAll,
	)

	tflog.Debug(ctx, "Updating Fastly service backend", map[string]any{
		"service_id": opts.ServiceID,
		"version":    opts.ServiceVersion,
		"name":       opts.Name,
		"force_all":  forceAll,
	})

	b, err := r.providerData.client.UpdateBackend(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating explicit service backend", err.Error())
		return
	}

	flattenServiceBackend(ctx, b, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceBackendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceBackendModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Fastly service backend", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	if err := ensureServiceTypeSupported(ctx, r.providerData.client, state.Service.ValueString(), "fastly_service_backend", serviceTypeVCL, serviceTypeCompute); err != nil {
		if isFastlyNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	notFound, diags := r.providerData.ensureVersionMutableForDelete(ctx, state.Service.ValueString(), int(state.Version.ValueInt64()))
	resp.Diagnostics.Append(diags...)
	if notFound || resp.Diagnostics.HasError() {
		return
	}

	err := r.providerData.client.DeleteBackend(ctx, &fastly.DeleteBackendInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if isFastlyNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting explicit service backend", err.Error())
	}
}

func (r *serviceBackendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}
	var serviceID types.String
	var version types.Int64
	var name types.String
	resp.Diagnostics.Append(req.Identity.GetAttribute(ctx, path.Root("service_id"), &serviceID)...)
	resp.Diagnostics.Append(req.Identity.GetAttribute(ctx, path.Root("version"), &version)...)
	resp.Diagnostics.Append(req.Identity.GetAttribute(ctx, path.Root("name"), &name)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

func (r *serviceBackendResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"service_id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Fastly service ID.",
			},
			"version": identityschema.Int64Attribute{
				RequiredForImport: true,
				Description:       "Fastly service version.",
			},
			"name": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Backend name.",
			},
		},
	}
}

func flattenServiceBackend(ctx context.Context, b *fastly.Backend, m *serviceBackendModel) {
	if b == nil {
		tflog.Warn(ctx, "flattenServiceBackend called with nil backend")
		return
	}

	id := fastly.ToValue(b.ServiceID) + "-" + strconv.Itoa(fastly.ToValue(b.ServiceVersion)) + "-" + fastly.ToValue(b.Name)
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(fastly.ToValue(b.ServiceID))
	m.Version = types.Int64Value(int64(fastly.ToValue(b.ServiceVersion)))

	backendModel := flattenBackendToVCLModel(b)
	applyVCLModelToExplicitBackend(backendModel, m)

	tflog.Debug(ctx, "Flattened service backend state", map[string]any{
		"id":      id,
		"service": m.Service.ValueString(),
		"version": m.Version.ValueInt64(),
		"name":    m.Name.ValueString(),
	})
}
