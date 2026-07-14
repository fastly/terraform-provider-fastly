package kvstores

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"

	"github.com/fastly/go-fastly/v16/fastly"
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
	ID     types.String `tfsdk:"id"`
	Stores types.Set    `tfsdk:"stores"`
}

var storeAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kvstores"
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a list of Fastly KV Stores.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform data source identifier.",
			},
			"stores": schema.SetNestedAttribute{
				Computed:    true,
				Description: "List of all KV Stores.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Identifier of the KV Store.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the KV Store.",
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

	tflog.Debug(ctx, "Reading Fastly KV Stores")

	var stores []fastly.KVStore
	p := d.client.NewListKVStoresPaginator(ctx, &fastly.ListKVStoresInput{})
	for p.Next() {
		stores = append(stores, p.Stores()...)
	}
	if err := p.Err(); err != nil {
		resp.Diagnostics.AddError("Error listing KV Stores", err.Error())
		return
	}

	ids := make([]string, 0, len(stores))
	elements := make([]attr.Value, 0, len(stores))
	for _, store := range stores {
		ids = append(ids, store.StoreID)

		obj, diags := types.ObjectValue(storeAttrTypes, map[string]attr.Value{
			"id":   types.StringValue(store.StoreID),
			"name": types.StringValue(store.Name),
		})
		resp.Diagnostics.Append(diags...)
		elements = append(elements, obj)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	setVal, diags := types.SetValue(types.ObjectType{AttrTypes: storeAttrTypes}, elements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Stores = setVal
	state.ID = types.StringValue(hashIDs(ids))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// hashIDs derives a stable data source ID from the set of store IDs,
// so the ID changes whenever the underlying set of KV Stores changes.
func hashIDs(ids []string) string {
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)

	sum := sha256.Sum256([]byte(strings.Join(sorted, ",")))
	return hex.EncodeToString(sum[:])
}
