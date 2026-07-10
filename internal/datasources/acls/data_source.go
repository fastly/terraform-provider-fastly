package acls

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/go-fastly/v16/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	ID   types.String `tfsdk:"id"`
	Acls types.Set    `tfsdk:"acls"`
}

var aclAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acls"
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of Fastly ACLs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform data source identifier.",
			},
			"acls": schema.SetNestedAttribute{
				Computed:    true,
				Description: "List of all ACLs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Identifier of the ACL.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the ACL.",
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

	tflog.Debug(ctx, "Reading Fastly ACLs")

	acls, err := computeacls.ListACLs(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Error listing ACLs", err.Error())
		return
	}

	ids := make([]string, 0, len(acls.Data))
	elements := make([]attr.Value, 0, len(acls.Data))
	for _, acl := range acls.Data {
		ids = append(ids, acl.ComputeACLID)

		obj, diags := types.ObjectValue(aclAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(acl.ComputeACLID),
			"name": types.StringValue(acl.Name),
		})
		resp.Diagnostics.Append(diags...)
		elements = append(elements, obj)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	setVal, diags := types.SetValue(types.ObjectType{AttrTypes: aclAttrTypes}, elements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Acls = setVal
	state.ID = types.StringValue(hashIDs(ids))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// hashIDs derives a stable data source ID from the set of ACL IDs,
// so the ID changes whenever the underlying set of ACLs changes.
func hashIDs(ids []string) string {
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)

	sum := sha256.Sum256([]byte(strings.Join(sorted, ",")))
	return hex.EncodeToString(sum[:])
}
