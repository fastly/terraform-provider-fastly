package backend

import (
	"context"
	"strings"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/importutil"
	"github.com/fastly/terraform-provider-fastly/internal/service"
	"github.com/fastly/terraform-provider-fastly/internal/validation"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithIdentity = &Resource{}

type Resource struct {
	providerData *fastlyclient.Data
}

func NewResource() resource.Resource {
	return &Resource{}
}

type Model struct {
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

type BackendIdentityModel struct {
	ServiceID types.String `tfsdk:"service_id"`
	Name      types.String `tfsdk:"name"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_backend"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fastly service backend resource. Writes directly to the specified writable service version.",
		Attributes:  ResourceAttributes(),
	}
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}

	r.providerData = data
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, plan.Service.ValueString(), "fastly_service_backend", service.TypeVCL, service.TypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Creating Fastly service backend", map[string]any{
		"service_id": plan.Service.ValueString(),
		"version":    plan.Version.ValueInt64(),
		"name":       service.StringValue(plan.Name),
	})

	resp.Diagnostics.Append(r.providerData.VersionChecker.EnsureMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := BuildCreateInput(plan.Service.ValueString(), int(plan.Version.ValueInt64()), ModelToNested(plan))

	b, err := r.providerData.Client.CreateBackend(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating explicit service backend", err.Error())
		return
	}

	flatten(ctx, b, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, &BackendIdentityModel{
			ServiceID: plan.Service,
			Name:      plan.Name,
		})...)
	}
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Fastly service backend from API", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	b, err := r.providerData.Client.GetBackend(ctx, &fastly.GetBackendInput{
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

	flatten(ctx, b, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, &BackendIdentityModel{
			ServiceID: state.Service,
			Name:      state.Name,
		})...)
	}
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model
	var state Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, plan.Service.ValueString(), "fastly_service_backend", service.TypeVCL, service.TypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	resp.Diagnostics.Append(r.providerData.VersionChecker.EnsureMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	forceAll := plan.Version.ValueInt64() != state.Version.ValueInt64()
	stateBackend := ModelToNested(state)
	opts := BuildUpdateInput(
		plan.Service.ValueString(),
		int(plan.Version.ValueInt64()),
		ModelToNested(plan),
		&stateBackend,
		forceAll,
	)

	tflog.Debug(ctx, "Updating Fastly service backend", map[string]any{
		"service_id": opts.ServiceID,
		"version":    opts.ServiceVersion,
		"name":       opts.Name,
		"force_all":  forceAll,
	})

	b, err := r.providerData.Client.UpdateBackend(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating explicit service backend", err.Error())
		return
	}

	flatten(ctx, b, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	// Update identity to reflect any changes
	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, &BackendIdentityModel{
			ServiceID: plan.Service,
			Name:      plan.Name,
		})...)
	}
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Fastly service backend", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	if err := validation.EnsureServiceTypeSupported(ctx, r.providerData.ServiceTypeChecker, state.Service.ValueString(), "fastly_service_backend", service.TypeVCL, service.TypeCompute); err != nil {
		if fastlyclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	notFound, diags := r.providerData.VersionChecker.EnsureMutableForDelete(ctx, state.Service.ValueString(), int(state.Version.ValueInt64()))
	resp.Diagnostics.Append(diags...)
	if notFound || resp.Diagnostics.HasError() {
		return
	}

	err := r.providerData.Client.DeleteBackend(ctx, &fastly.DeleteBackendInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if fastlyclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting explicit service backend", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Support legacy composite ID format: service_id/version/name
	if req.ID != "" && strings.Contains(req.ID, "/") {
		serviceID, version, name, err := importutil.ParseCompositeID(req.ID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				"Expected import ID in format: service_id/version/name\n"+
					"For example: service123/3/origin\n\n"+
					"Error: "+err.Error(),
			)
			return
		}

		tflog.Debug(ctx, "Importing backend with legacy composite ID", map[string]any{
			"service_id": serviceID,
			"version":    version,
			"name":       name,
		})

		// Use the API to read the full backend configuration
		b, err := r.providerData.Client.GetBackend(ctx, &fastly.GetBackendInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Name:           name,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error importing backend", err.Error())
			return
		}

		// Populate state with the full backend data
		var state Model
		state.Service = types.StringValue(serviceID)
		state.Version = types.Int64Value(int64(version))
		flatten(ctx, b, &state)

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Set identity using stable identity schema (service_id + name only)
		if resp.Identity != nil {
			resp.Diagnostics.Append(resp.Identity.Set(ctx, &BackendIdentityModel{
				ServiceID: types.StringValue(serviceID),
				Name:      types.StringValue(name),
			})...)
		}
		return
	}

	// Support new identity-based import
	if req.ID != "" {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}

	var identity BackendIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &BackendIdentityModel{
		ServiceID: identity.ServiceID,
		Name:      identity.Name,
	})...)
}

func (r *Resource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"service_id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Fastly service ID.",
			},
			"name": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "Backend name.",
			},
		},
	}
}
