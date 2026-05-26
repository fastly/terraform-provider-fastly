package provider

import (
	"context"
	"fmt"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type serviceComputeAutoResource struct {
	providerData *providerData
}

var _ resource.Resource = &serviceComputeAutoResource{}
var _ resource.ResourceWithConfigure = &serviceComputeAutoResource{}
var _ resource.ResourceWithImportState = &serviceComputeAutoResource{}

func NewServiceComputeAutoResource() resource.Resource {
	return &serviceComputeAutoResource{}
}

type serviceComputeAutoModel struct {
	ID             types.String                 `tfsdk:"id"`
	Name           types.String                 `tfsdk:"name"`
	Comment        types.String                 `tfsdk:"comment"`
	ForceDestroy   types.Bool                   `tfsdk:"force_destroy"`
	Reuse          types.Bool                   `tfsdk:"reuse"`
	ActiveVersion  types.Int64                  `tfsdk:"active_version"`
	ManagedVersion types.Int64                  `tfsdk:"managed_version"`
	Domain         []serviceDomainNestedModel   `tfsdk:"domain"`
	Backend        []serviceBackendNestedModel  `tfsdk:"backend"`
	Package        []serviceComputePackageModel `tfsdk:"package"`
}

func (r *serviceComputeAutoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_compute_auto"
}

func (r *serviceComputeAutoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Automatic-lifecycle Fastly Compute service resource with nested versioned configuration. The provider automatically clones, validates, and activates changed versions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The Fastly service ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The service name.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Managed by Terraform"),
				Description: "Optional service comment.",
			},
			"force_destroy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Deactivate the active version before deleting the service. Default `false`.",
			},
			"reuse": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Deactivate the active version but do not delete the service, allowing it to be reused/imported elsewhere. Default `false`.",
			},
			"active_version": schema.Int64Attribute{
				Computed:    true,
				Description: "The currently active service version.",
			},
			"managed_version": schema.Int64Attribute{
				Computed:    true,
				Description: "The latest service version selected and managed by this resource.",
			},
		},
		Blocks: map[string]schema.Block{
			"domain":  domainNestedBlockSchema(),
			"backend": backendNestedBlockSchema(),
			"package": computePackageNestedBlockSchema(),
		},
	}
}

func (r *serviceComputeAutoResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	pd, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *providerData, got: %T", req.ProviderData),
		)
		return
	}

	r.providerData = pd
}

func (r *serviceComputeAutoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceComputeAutoModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(plan.Package) == 0 {
		resp.Diagnostics.AddError(
			"Missing Compute package",
			"`fastly_service_compute_auto` automatically validates and activates service versions, so a package block is required when creating a Compute service.",
		)
		return
	}

	if err := validateComputePackageInput(plan.Package); err != nil {
		resp.Diagnostics.AddError("Invalid Compute package", err.Error())
		return
	}

	service, err := r.providerData.client.CreateService(ctx, &fastly.CreateServiceInput{
		Name:    fastly.ToPointer(plan.Name.ValueString()),
		Comment: fastly.ToPointer(plan.Comment.ValueString()),
		Type:    fastly.ToPointer(serviceTypeCompute),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Fastly Compute service", err.Error())
		return
	}

	serviceID := fastly.ToValue(service.ServiceID)
	version := 1

	tflog.Info(ctx, "Created Fastly Compute service", map[string]any{
		"service_id": serviceID,
		"version":    version,
	})

	if err := reconcileDomains(ctx, r.providerData.client, serviceID, version, plan.Domain); err != nil {
		resp.Diagnostics.AddError("Error reconciling domains", err.Error())
		return
	}

	if err := reconcileBackends(ctx, r.providerData.client, serviceID, version, plan.Backend); err != nil {
		resp.Diagnostics.AddError("Error reconciling backends", err.Error())
		return
	}

	if err := updateComputePackage(ctx, r.providerData.client, serviceID, version, plan.Package); err != nil {
		resp.Diagnostics.AddError("Error updating Compute package", err.Error())
		return
	}

	packages, err := readComputePackageForVersion(ctx, r.providerData.client, serviceID, version, plan.Package)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Compute package", err.Error())
		return
	}
	plan.Package = packages

	if err := validateServiceVersion(ctx, r.providerData.client, serviceID, version); err != nil {
		resp.Diagnostics.AddError("Error validating service version", err.Error())
		return
	}

	plan.ID = types.StringValue(serviceID)
	plan.ManagedVersion = types.Int64Value(int64(version))

	if _, err := r.providerData.client.ActivateVersion(ctx, &fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	}); err != nil {
		resp.Diagnostics.AddError("Error activating service version", err.Error())
		return
	}
	plan.ActiveVersion = types.Int64Value(int64(version))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceComputeAutoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceComputeAutoModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := r.providerData.client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: state.ID.ValueString(),
	})
	if err != nil {
		if httpErr, ok := err.(*fastly.HTTPError); ok && httpErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Fastly Compute service", err.Error())
		return
	}

	serviceType := fastly.ToValue(service.Type)
	if serviceType != serviceTypeCompute {
		resp.Diagnostics.AddError(
			"Unexpected Fastly service type",
			fmt.Sprintf("Expected Compute service %q to have type %q, got %q.", state.ID.ValueString(), serviceTypeCompute, serviceType),
		)
		return
	}

	if service.Name != nil {
		state.Name = types.StringValue(*service.Name)
	}
	if service.Comment != nil {
		state.Comment = types.StringValue(*service.Comment)
	}

	readVersion, active, err := selectServiceReadVersionFromDetails(service, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error selecting service version for read", err.Error())
		return
	}

	if active {
		state.ActiveVersion = types.Int64Value(int64(readVersion))
	} else {
		state.ActiveVersion = types.Int64Null()
	}
	state.ManagedVersion = types.Int64Value(int64(readVersion))

	domains, err := readDomainsForVersion(ctx, r.providerData.client, state.ID.ValueString(), readVersion)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service domains", err.Error())
		return
	}
	backends, err := readBackendsForVersion(ctx, r.providerData.client, state.ID.ValueString(), readVersion)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service backends", err.Error())
		return
	}
	state.Domain = domains
	state.Backend = backends

	packages, err := readComputePackageForVersion(ctx, r.providerData.client, state.ID.ValueString(), readVersion, state.Package)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Compute package", err.Error())
		return
	}
	state.Package = packages

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceComputeAutoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceComputeAutoModel
	var state serviceComputeAutoModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ID.ValueString()

	// Update service metadata in place. Name and comment are versionless service fields.
	_, err := r.providerData.client.UpdateService(ctx, &fastly.UpdateServiceInput{
		ServiceID: serviceID,
		Name:      fastly.ToPointer(plan.Name.ValueString()),
		Comment:   fastly.ToPointer(plan.Comment.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating Fastly Compute service", err.Error())
		return
	}

	nestedChanged := !domainsEqual(plan.Domain, state.Domain) || !backendsEqual(plan.Backend, state.Backend) || !computePackagesEqual(plan.Package, state.Package)
	needsVersionChange := nestedChanged

	targetVersion := 0

	if needsVersionChange {
		sourceVersion, shouldClone, err := r.selectWorkingVersion(ctx, serviceID)
		if err != nil {
			resp.Diagnostics.AddError("Error selecting Fastly service version", err.Error())
			return
		}

		if shouldClone {
			cloned, err := r.providerData.client.CloneVersion(ctx, &fastly.CloneVersionInput{
				ServiceID:      serviceID,
				ServiceVersion: sourceVersion,
			})
			if err != nil {
				resp.Diagnostics.AddError("Error cloning Fastly service version", err.Error())
				return
			}
			targetVersion = fastly.ToValue(cloned.Number)
		} else {
			targetVersion = sourceVersion
		}

		tflog.Info(ctx, "Selected Fastly Compute service working version", map[string]any{
			"service_id":     serviceID,
			"source_version": sourceVersion,
			"target_version": targetVersion,
			"cloned":         shouldClone,
			"nested_changed": nestedChanged,
		})

		if err := reconcileDomains(ctx, r.providerData.client, serviceID, targetVersion, plan.Domain); err != nil {
			resp.Diagnostics.AddError("Error reconciling domains", err.Error())
			return
		}

		if err := reconcileBackends(ctx, r.providerData.client, serviceID, targetVersion, plan.Backend); err != nil {
			resp.Diagnostics.AddError("Error reconciling backends", err.Error())
			return
		}

		if len(state.Package) > 0 && len(plan.Package) == 0 {
			resp.Diagnostics.AddError(
				"Removing Compute packages is not supported",
				"The Fastly API does not currently support deleting a package from a service version. Provide a package block or create a new service/version workflow that does not rely on package removal.",
			)
			return
		}

		if err := updateComputePackage(ctx, r.providerData.client, serviceID, targetVersion, plan.Package); err != nil {
			resp.Diagnostics.AddError("Error updating Compute package", err.Error())
			return
		}

		packages, err := readComputePackageForVersion(ctx, r.providerData.client, serviceID, targetVersion, plan.Package)
		if err != nil {
			resp.Diagnostics.AddError("Error reading Compute package", err.Error())
			return
		}
		plan.Package = packages

		if err := validateServiceVersion(ctx, r.providerData.client, serviceID, targetVersion); err != nil {
			resp.Diagnostics.AddError("Error validating service version", err.Error())
			return
		}

		plan.ManagedVersion = types.Int64Value(int64(targetVersion))

		if _, err := r.providerData.client.ActivateVersion(ctx, &fastly.ActivateVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: targetVersion,
		}); err != nil {
			resp.Diagnostics.AddError("Error activating service version", err.Error())
			return
		}
		plan.ActiveVersion = types.Int64Value(int64(targetVersion))
	} else {
		plan.ManagedVersion = state.ManagedVersion
		plan.ActiveVersion = state.ActiveVersion
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceComputeAutoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceComputeAutoModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := deleteServiceWithPolicy(
		ctx,
		r.providerData.client,
		state.ID.ValueString(),
		boolValue(state.ForceDestroy),
		boolValue(state.Reuse),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Fastly Compute service", err.Error())
	}
}

func (r *serviceComputeAutoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *serviceComputeAutoResource) selectWorkingVersion(ctx context.Context, serviceID string) (version int, shouldClone bool, err error) {
	service, err := r.providerData.client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: serviceID,
	})
	if err != nil {
		return 0, false, err
	}

	return selectServiceWorkingVersionFromDetails(service, serviceID)
}
