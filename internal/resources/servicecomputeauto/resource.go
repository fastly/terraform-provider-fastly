package servicecomputeauto

import (
	"context"
	"fmt"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/computepackage"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/resources/acl"
	"github.com/fastly/terraform-provider-fastly/internal/resources/backend"
	"github.com/fastly/terraform-provider-fastly/internal/resources/domain"
	"github.com/fastly/terraform-provider-fastly/internal/service"

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

type Resource struct {
	providerData *fastlyclient.Data
}

var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

type Model struct {
	ID             types.String           `tfsdk:"id"`
	Name           types.String           `tfsdk:"name"`
	Comment        types.String           `tfsdk:"comment"`
	ForceDestroy   types.Bool             `tfsdk:"force_destroy"`
	Reuse          types.Bool             `tfsdk:"reuse"`
	ActiveVersion  types.Int64            `tfsdk:"active_version"`
	ManagedVersion types.Int64            `tfsdk:"managed_version"`
	Domain         []domain.NestedModel   `tfsdk:"domain"`
	Backend        []backend.NestedModel  `tfsdk:"backend"`
	ACL            []acl.NestedModel      `tfsdk:"acl"`
	Package        []computepackage.Model `tfsdk:"package"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_compute_auto"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"domain":  domain.NestedBlockSchema(),
			"backend": backend.NestedBlockSchema(),
			"acl":     acl.NestedBlockSchema(),
			"package": computepackage.NestedBlockSchema(),
		},
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

	if len(plan.Package) == 0 {
		resp.Diagnostics.AddError(
			"Missing Compute package",
			"`fastly_service_compute_auto` automatically validates and activates service versions, so a package block is required when creating a Compute service.",
		)
		return
	}

	if err := computepackage.ValidateInput(plan.Package); err != nil {
		resp.Diagnostics.AddError("Invalid Compute package", err.Error())
		return
	}

	created, err := r.providerData.AutoClient().CreateService(ctx, &fastly.CreateServiceInput{
		Name:    new(plan.Name.ValueString()),
		Comment: new(plan.Comment.ValueString()),
		Type:    new(service.TypeCompute),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Fastly Compute service", err.Error())
		return
	}

	serviceID := fastly.ToValue(created.ServiceID)
	version := 1

	tflog.Info(ctx, "Created Fastly Compute service", map[string]any{
		"service_id": serviceID,
		"version":    version,
	})

	if err := domain.Reconcile(ctx, r.providerData.AutoClient(), serviceID, version, plan.Domain); err != nil {
		resp.Diagnostics.AddError("Error reconciling domains", err.Error())
		return
	}

	domains, err := domain.ReadForVersion(ctx, r.providerData.AutoClient(), serviceID, version)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service domains", err.Error())
		return
	}
	plan.Domain = domain.MatchOrder(domains, plan.Domain)

	if err := backend.Reconcile(ctx, r.providerData.AutoClient(), serviceID, version, plan.Backend); err != nil {
		resp.Diagnostics.AddError("Error reconciling backends", err.Error())
		return
	}

	backends, err := backend.ReadForVersion(ctx, r.providerData.AutoClient(), serviceID, version)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service backends", err.Error())
		return
	}
	plan.Backend = backend.MatchOrder(backends, plan.Backend)

	if err := acl.Reconcile(ctx, r.providerData.AutoClient(), serviceID, version, plan.ACL); err != nil {
		resp.Diagnostics.AddError("Error reconciling ACLs", err.Error())
		return
	}

	acls, err := acl.ReadForVersionWithPlan(ctx, r.providerData.AutoClient(), serviceID, version, plan.ACL)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service ACLs", err.Error())
		return
	}
	plan.ACL = acls

	if err := computepackage.Update(ctx, r.providerData.AutoClient(), serviceID, version, plan.Package); err != nil {
		resp.Diagnostics.AddError("Error updating Compute package", err.Error())
		return
	}

	packages, err := computepackage.ReadForVersion(ctx, r.providerData.AutoClient(), serviceID, version, plan.Package)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Compute package", err.Error())
		return
	}
	plan.Package = packages

	if err := service.ValidateVersion(ctx, r.providerData.AutoClient(), serviceID, version); err != nil {
		resp.Diagnostics.AddError("Error validating service version", err.Error())
		return
	}

	plan.ID = types.StringValue(serviceID)
	plan.ManagedVersion = types.Int64Value(int64(version))

	if _, err := r.providerData.AutoClient().ActivateVersion(ctx, &fastly.ActivateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	}); err != nil {
		resp.Diagnostics.AddError("Error activating service version", err.Error())
		return
	}
	plan.ActiveVersion = types.Int64Value(int64(version))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	details, err := r.providerData.AutoClient().GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: state.ID.ValueString(),
	})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Fastly Compute service", err.Error())
		return
	}

	serviceType := fastly.ToValue(details.Type)
	if serviceType != service.TypeCompute {
		resp.Diagnostics.AddError(
			"Unexpected Fastly service type",
			fmt.Sprintf("Expected Compute service %q to have type %q, got %q.", state.ID.ValueString(), service.TypeCompute, serviceType),
		)
		return
	}

	if details.Name != nil {
		state.Name = types.StringValue(*details.Name)
	}
	if details.Comment != nil {
		state.Comment = types.StringValue(*details.Comment)
	}

	readVersion, active, err := service.SelectReadVersionFromDetails(details, state.ID.ValueString())
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

	domains, err := domain.ReadForVersion(ctx, r.providerData.AutoClient(), state.ID.ValueString(), readVersion)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service domains", err.Error())
		return
	}
	backends, err := backend.ReadForVersion(ctx, r.providerData.AutoClient(), state.ID.ValueString(), readVersion)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service backends", err.Error())
		return
	}
	acls, err := acl.ReadForVersionWithPlan(ctx, r.providerData.AutoClient(), state.ID.ValueString(), readVersion, state.ACL)
	if err != nil {
		resp.Diagnostics.AddError("Error reading service ACLs", err.Error())
		return
	}
	state.Domain = domain.MatchOrder(domains, state.Domain)
	state.Backend = backend.MatchOrder(backends, state.Backend)
	state.ACL = acl.MatchOrder(acls, state.ACL)

	packages, err := computepackage.ReadForVersion(ctx, r.providerData.AutoClient(), state.ID.ValueString(), readVersion, state.Package)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Compute package", err.Error())
		return
	}
	state.Package = packages

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Model
	var state Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ID.ValueString()

	// Update service metadata in place. Name and comment are versionless service fields.
	_, err := r.providerData.AutoClient().UpdateService(ctx, &fastly.UpdateServiceInput{
		ServiceID: serviceID,
		Name:      new(plan.Name.ValueString()),
		Comment:   new(plan.Comment.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating Fastly Compute service", err.Error())
		return
	}

	nestedChanged := !domain.Equal(plan.Domain, state.Domain) || !backend.Equal(plan.Backend, state.Backend) || !acl.Equal(plan.ACL, state.ACL) || !computepackage.Equal(plan.Package, state.Package)
	needsVersionChange := nestedChanged

	targetVersion := 0

	if needsVersionChange {
		sourceVersion, shouldClone, err := r.selectWorkingVersion(ctx, serviceID)
		if err != nil {
			resp.Diagnostics.AddError("Error selecting Fastly service version", err.Error())
			return
		}

		if shouldClone {
			cloned, err := r.providerData.AutoClient().CloneVersion(ctx, &fastly.CloneVersionInput{
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

		if err := domain.Reconcile(ctx, r.providerData.AutoClient(), serviceID, targetVersion, plan.Domain); err != nil {
			resp.Diagnostics.AddError("Error reconciling domains", err.Error())
			return
		}

		domains, err := domain.ReadForVersion(ctx, r.providerData.AutoClient(), serviceID, targetVersion)
		if err != nil {
			resp.Diagnostics.AddError("Error reading service domains", err.Error())
			return
		}
		plan.Domain = domain.MatchOrder(domains, plan.Domain)

		if err := backend.Reconcile(ctx, r.providerData.AutoClient(), serviceID, targetVersion, plan.Backend); err != nil {
			resp.Diagnostics.AddError("Error reconciling backends", err.Error())
			return
		}

		backends, err := backend.ReadForVersion(ctx, r.providerData.AutoClient(), serviceID, targetVersion)
		if err != nil {
			resp.Diagnostics.AddError("Error reading service backends", err.Error())
			return
		}
		plan.Backend = backend.MatchOrder(backends, plan.Backend)

		if err := acl.Reconcile(ctx, r.providerData.AutoClient(), serviceID, targetVersion, plan.ACL); err != nil {
			resp.Diagnostics.AddError("Error reconciling ACLs", err.Error())
			return
		}

		acls, err := acl.ReadForVersionWithPlan(ctx, r.providerData.AutoClient(), serviceID, targetVersion, plan.ACL)
		if err != nil {
			resp.Diagnostics.AddError("Error reading service ACLs", err.Error())
			return
		}
		plan.ACL = acls

		if len(state.Package) > 0 && len(plan.Package) == 0 {
			resp.Diagnostics.AddError(
				"Removing Compute packages is not supported",
				"The Fastly API does not currently support deleting a package from a service version. Provide a package block or create a new service/version workflow that does not rely on package removal.",
			)
			return
		}

		if err := computepackage.Update(ctx, r.providerData.AutoClient(), serviceID, targetVersion, plan.Package); err != nil {
			resp.Diagnostics.AddError("Error updating Compute package", err.Error())
			return
		}

		packages, err := computepackage.ReadForVersion(ctx, r.providerData.AutoClient(), serviceID, targetVersion, plan.Package)
		if err != nil {
			resp.Diagnostics.AddError("Error reading Compute package", err.Error())
			return
		}
		plan.Package = packages

		if err := service.ValidateVersion(ctx, r.providerData.AutoClient(), serviceID, targetVersion); err != nil {
			resp.Diagnostics.AddError("Error validating service version", err.Error())
			return
		}

		plan.ManagedVersion = types.Int64Value(int64(targetVersion))

		if _, err := r.providerData.AutoClient().ActivateVersion(ctx, &fastly.ActivateVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: targetVersion,
		}); err != nil {
			resp.Diagnostics.AddError("Error activating service version", err.Error())
			return
		}
		plan.ActiveVersion = types.Int64Value(int64(targetVersion))
	} else {
		// No version change needed - preserve version numbers and order nested state to match the plan
		plan.ManagedVersion = state.ManagedVersion
		plan.ActiveVersion = state.ActiveVersion
		plan.Domain = domain.MatchOrder(state.Domain, plan.Domain)
		plan.Backend = backend.MatchOrder(state.Backend, plan.Backend)
		plan.ACL = acl.MatchOrder(state.ACL, plan.ACL)
		plan.Package = state.Package
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := service.DeleteWithPolicy(
		ctx,
		r.providerData.AutoClient(),
		state.ID.ValueString(),
		service.BoolValue(state.ForceDestroy),
		service.BoolValue(state.Reuse),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Fastly Compute service", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) selectWorkingVersion(ctx context.Context, serviceID string) (version int, shouldClone bool, err error) {
	details, err := r.providerData.AutoClient().GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: serviceID,
	})
	if err != nil {
		return 0, false, err
	}

	return service.SelectWorkingVersionFromDetails(details, serviceID)
}
