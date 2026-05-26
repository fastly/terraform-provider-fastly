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

var _ resource.Resource = &serviceDomainResource{}
var _ resource.ResourceWithImportState = &serviceDomainResource{}
var _ resource.ResourceWithIdentity = &serviceDomainResource{}

type serviceDomainResource struct {
	providerData *providerData
}

func NewServiceDomainResource() resource.Resource {
	return &serviceDomainResource{}
}

type serviceDomainModel struct {
	ID      types.String `tfsdk:"id"`
	Service types.String `tfsdk:"service_id"`
	Version types.Int64  `tfsdk:"version"`
	Name    types.String `tfsdk:"name"`
	Comment types.String `tfsdk:"comment"`
}

func (r *serviceDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_domain"
}

func (r *serviceDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: domainResourceAttributes(),
	}
}

func (r *serviceDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		return
	}
	r.providerData = providerData
}

func (r *serviceDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := ensureServiceTypeSupported(ctx, r.providerData.client, plan.Service.ValueString(), "fastly_service_domain", serviceTypeVCL, serviceTypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	opts := expandServiceDomainCreate(plan)
	tflog.Debug(ctx, "Creating Fastly service domain", map[string]any{
		"service_id": opts.ServiceID,
		"version":    opts.ServiceVersion,
		"name":       *opts.Name,
		"comment":    opts.Comment,
	})

	resp.Diagnostics.Append(r.providerData.ensureVersionMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.providerData.client.CreateDomain(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating explicit service domain", err.Error())
		return
	}

	flattenServiceDomain(ctx, d, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Fastly service domain from API", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	d, err := r.providerData.client.GetDomain(ctx, &fastly.GetDomainInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if fastlyErr, ok := err.(*fastly.HTTPError); ok && fastlyErr.StatusCode == 404 {
			tflog.Warn(ctx, "Service domain not found, removing from state", map[string]any{
				"service_id": state.Service.ValueString(),
				"name":       state.Name.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading explicit service domain", err.Error())
		return
	}

	flattenServiceDomain(ctx, d, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := ensureServiceTypeSupported(ctx, r.providerData.client, plan.Service.ValueString(), "fastly_service_domain", serviceTypeVCL, serviceTypeCompute); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	opts := expandServiceDomainUpdate(plan)
	tflog.Debug(ctx, "Updating Fastly service domain", map[string]any{
		"service_id": opts.ServiceID,
		"version":    opts.ServiceVersion,
		"name":       opts.Name,
		"comment":    opts.Comment,
	})

	resp.Diagnostics.Append(r.providerData.ensureVersionMutable(ctx, plan.Service.ValueString(), int(plan.Version.ValueInt64()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.providerData.client.UpdateDomain(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating explicit service domain", err.Error())
		return
	}

	flattenServiceDomain(ctx, d, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Fastly service domain", map[string]any{
		"service_id": state.Service.ValueString(),
		"version":    state.Version.ValueInt64(),
		"name":       state.Name.ValueString(),
	})

	if err := ensureServiceTypeSupported(ctx, r.providerData.client, state.Service.ValueString(), "fastly_service_domain", serviceTypeVCL, serviceTypeCompute); err != nil {
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

	err := r.providerData.client.DeleteDomain(ctx, &fastly.DeleteDomainInput{
		ServiceID:      state.Service.ValueString(),
		ServiceVersion: int(state.Version.ValueInt64()),
		Name:           state.Name.ValueString(),
	})
	if err != nil {
		if isFastlyNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting explicit service domain", err.Error())
	}
}

func (r *serviceDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *serviceDomainResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
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
				Description:       "Domain name.",
			},
		},
	}
}

func expandServiceDomainCreate(m serviceDomainModel) *fastly.CreateDomainInput {
	opts := &fastly.CreateDomainInput{
		ServiceID:      m.Service.ValueString(),
		ServiceVersion: int(m.Version.ValueInt64()),
		Name:           fastly.ToPointer(m.Name.ValueString()),
	}
	if !m.Comment.IsNull() && m.Comment.ValueString() != "" {
		opts.Comment = fastly.ToPointer(m.Comment.ValueString())
	}
	return opts
}

func expandServiceDomainUpdate(m serviceDomainModel) *fastly.UpdateDomainInput {
	opts := &fastly.UpdateDomainInput{
		ServiceID:      m.Service.ValueString(),
		ServiceVersion: int(m.Version.ValueInt64()),
		Name:           m.Name.ValueString(),
	}
	if !m.Comment.IsNull() && m.Comment.ValueString() != "" {
		opts.Comment = fastly.ToPointer(m.Comment.ValueString())
	}
	return opts
}

func flattenServiceDomain(ctx context.Context, d *fastly.Domain, m *serviceDomainModel) {
	if d == nil {
		tflog.Warn(ctx, "flattenServiceDomain called with nil domain")
		return
	}

	id := *d.ServiceID + "-" + strconv.Itoa(*d.ServiceVersion) + "-" + *d.Name
	m.ID = types.StringValue(id)
	m.Service = types.StringValue(*d.ServiceID)
	m.Version = types.Int64Value(int64(*d.ServiceVersion))
	m.Name = types.StringValue(*d.Name)

	if d.Comment != nil && *d.Comment != "" {
		m.Comment = types.StringValue(*d.Comment)
	} else {
		m.Comment = types.StringNull()
	}

	tflog.Debug(ctx, "Flattened service domain state", map[string]any{
		"id":      id,
		"service": *d.ServiceID,
		"version": *d.ServiceVersion,
		"name":    *d.Name,
		"comment": d.Comment,
	})
}
