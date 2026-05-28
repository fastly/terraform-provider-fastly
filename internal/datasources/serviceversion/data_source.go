package serviceversion

import (
	"context"

	"sort"
	"time"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DataSource{}

type DataSource struct {
	client *fastly.Client
}

type DataSourceModel struct {
	ID             types.String    `tfsdk:"id"`
	ServiceID      types.String    `tfsdk:"service_id"`
	LatestVersion  types.Int64     `tfsdk:"latest_version"`
	ActiveVersion  types.Int64     `tfsdk:"active_version"`
	StagingVersion types.Int64     `tfsdk:"staging_version"`
	LockedVersions []types.Int64   `tfsdk:"locked_versions"`
	Versions       []MetadataModel `tfsdk:"versions"`
}

type MetadataModel struct {
	Number       types.Int64        `tfsdk:"number"`
	Active       types.Bool         `tfsdk:"active"`
	Staging      types.Bool         `tfsdk:"staging"`
	Locked       types.Bool         `tfsdk:"locked"`
	Deployed     types.Bool         `tfsdk:"deployed"`
	Testing      types.Bool         `tfsdk:"testing"`
	Comment      types.String       `tfsdk:"comment"`
	ServiceID    types.String       `tfsdk:"service_id"`
	CreatedAt    types.String       `tfsdk:"created_at"`
	UpdatedAt    types.String       `tfsdk:"updated_at"`
	DeletedAt    types.String       `tfsdk:"deleted_at"`
	Environments []EnvironmentModel `tfsdk:"environments"`
}

type EnvironmentModel struct {
	Name          types.String `tfsdk:"name"`
	ActiveVersion types.Int64  `tfsdk:"active_version"`
	ServiceID     types.String `tfsdk:"service_id"`
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_version"
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read-only view of Fastly service versions for a service. This data source is for observability only and never creates, clones, activates, or mutates versions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform data source identifier. Mirrors service_id.",
			},
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "Fastly service ID.",
			},
			"latest_version": schema.Int64Attribute{
				Computed:    true,
				Description: "Highest version number that exists for the service.",
			},
			"active_version": schema.Int64Attribute{
				Computed:    true,
				Description: "Currently active production version, if any.",
			},
			"staging_version": schema.Int64Attribute{
				Computed:    true,
				Description: "Currently staged version, if any.",
			},
			"locked_versions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Versions that are locked and therefore not editable.",
			},
			"versions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Per-version metadata for the service.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"number": schema.Int64Attribute{
							Computed:    true,
							Description: "Service version number.",
						},
						"active": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this version is active in production.",
						},
						"staging": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this version is active in staging.",
						},
						"locked": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this version is locked.",
						},
						"deployed": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this version has been deployed.",
						},
						"testing": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this version is in testing mode.",
						},
						"comment": schema.StringAttribute{
							Computed:    true,
							Description: "Version comment.",
						},
						"service_id": schema.StringAttribute{
							Computed:    true,
							Description: "Fastly service ID this version belongs to.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Creation timestamp in RFC3339 format.",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "Last update timestamp in RFC3339 format.",
						},
						"deleted_at": schema.StringAttribute{
							Computed:    true,
							Description: "Deletion timestamp in RFC3339 format, if any.",
						},
						"environments": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Deployment environment metadata reported by Fastly.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed: true,
									},
									"active_version": schema.Int64Attribute{
										Computed: true,
									},
									"service_id": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	data, diags := fastlyclient.FromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || data == nil {
		return
	}

	d.client = data.Client
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()

	tflog.Debug(ctx, "Reading Fastly service versions", map[string]any{
		"service_id": serviceID,
	})

	// Use ListVersions instead of GetServiceDetails because this data source only
	// needs version metadata. GetServiceDetails can return full service version
	// configuration, which is unnecessary and potentially large for this use case.
	versions, err := d.client.ListVersions(ctx, &fastly.ListVersionsInput{
		ServiceID: serviceID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing Fastly service versions", err.Error())
		return
	}

	state.ID = types.StringValue(serviceID)
	state.LatestVersion = types.Int64Null()
	state.ActiveVersion = types.Int64Null()
	state.StagingVersion = types.Int64Null()
	state.LockedVersions = make([]types.Int64, 0)
	state.Versions = make([]MetadataModel, 0, len(versions))

	lockedSet := make([]int64, 0)

	var latestVersion int64
	latestVersionSet := false

	for _, version := range versions {
		if version == nil {
			continue
		}

		number, numberOK := intPointerToInt64(version.Number)
		if numberOK {
			if !latestVersionSet || number > latestVersion {
				latestVersion = number
				latestVersionSet = true
			}
		}

		active := boolPointerValueRaw(version.Active)
		staging := boolPointerValueRaw(version.Staging)
		locked := boolPointerValueRaw(version.Locked)

		if active && numberOK {
			state.ActiveVersion = types.Int64Value(number)
		}
		if staging && numberOK {
			state.StagingVersion = types.Int64Value(number)
		}
		if locked && numberOK {
			lockedSet = append(lockedSet, number)
		}

		envs := make([]EnvironmentModel, 0, len(version.Environments))
		for _, env := range version.Environments {
			if env == nil {
				continue
			}
			envs = append(envs, EnvironmentModel{
				Name:          stringPointerValue(env.Name),
				ActiveVersion: int64PointerValue(env.ServiceVersion),
				ServiceID:     stringPointerValue(env.ServiceID),
			})
		}

		state.Versions = append(state.Versions, MetadataModel{
			Number:       intPointerValue(version.Number),
			Active:       types.BoolValue(active),
			Staging:      types.BoolValue(staging),
			Locked:       types.BoolValue(locked),
			Deployed:     types.BoolValue(boolPointerValueRaw(version.Deployed)),
			Testing:      types.BoolValue(boolPointerValueRaw(version.Testing)),
			Comment:      stringPointerValue(version.Comment),
			ServiceID:    stringPointerValue(version.ServiceID),
			CreatedAt:    timePointerValue(version.CreatedAt),
			UpdatedAt:    timePointerValue(version.UpdatedAt),
			DeletedAt:    timePointerValue(version.DeletedAt),
			Environments: envs,
		})
	}

	if latestVersionSet {
		state.LatestVersion = types.Int64Value(latestVersion)
	}

	sort.Slice(lockedSet, func(i, j int) bool { return lockedSet[i] < lockedSet[j] })
	for _, v := range lockedSet {
		state.LockedVersions = append(state.LockedVersions, types.Int64Value(v))
	}

	sort.Slice(state.Versions, func(i, j int) bool {
		return state.Versions[i].Number.ValueInt64() < state.Versions[j].Number.ValueInt64()
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func boolPointerValueRaw(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func stringPointerValue(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func intPointerValue(v *int) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}

func intPointerToInt64(v *int) (int64, bool) {
	if v == nil {
		return 0, false
	}
	return int64(*v), true
}

func int64PointerValue(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

func timePointerValue(v *time.Time) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(v.Format(time.RFC3339))
}
