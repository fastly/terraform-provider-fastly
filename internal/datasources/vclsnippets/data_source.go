package vclsnippets

import (
	"context"
	"fmt"
	"strconv"

	fastlyclient "github.com/fastly/terraform-provider-fastly/internal/client"
	"github.com/fastly/terraform-provider-fastly/internal/datasources/idhash"
	"github.com/fastly/terraform-provider-fastly/internal/service"
	"github.com/fastly/terraform-provider-fastly/internal/validation"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &DataSource{}

const defaultPriority int64 = 100

type DataSource struct {
	providerData *fastlyclient.Data
}

type DataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	ServiceID      types.String `tfsdk:"service_id"`
	ServiceVersion types.Int64  `tfsdk:"service_version"`
	VCLSnippets    types.Set    `tfsdk:"vcl_snippets"`
}

var snippetAttrTypes = map[string]attr.Type{
	"content":  types.StringType,
	"dynamic":  types.BoolType,
	"id":       types.StringType,
	"name":     types.StringType,
	"priority": types.Int64Type,
	"type":     types.StringType,
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vcl_snippets"
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve VCL snippets for a Fastly service version.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform data source identifier.",
			},
			"service_id": schema.StringAttribute{
				Required:    true,
				Description: "Fastly service ID.",
			},
			"service_version": schema.Int64Attribute{
				Required:    true,
				Description: "Fastly service version to read snippets from.",
			},
			"vcl_snippets": schema.SetNestedAttribute{
				Computed:    true,
				Description: "List of all VCL snippets for the configured service version.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"content": schema.StringAttribute{
							Computed:    true,
							Description: "The VCL code that specifies exactly what the snippet does.",
						},
						"dynamic": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this is a dynamic VCL snippet.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Fastly-generated VCL snippet identifier.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name for the snippet.",
						},
						"priority": schema.Int64Attribute{
							Computed:    true,
							Description: "Priority determines execution order. Lower numbers execute first.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The location in generated VCL where the snippet should be placed.",
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

	d.providerData = data
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	version := int(state.ServiceVersion.ValueInt64())

	if err := validation.EnsureServiceTypeSupported(ctx, d.providerData.ServiceTypeChecker, serviceID, "fastly_vcl_snippets", service.TypeVCL); err != nil {
		resp.Diagnostics.AddError("Unsupported Fastly service type", err.Error())
		return
	}

	tflog.Debug(ctx, "Reading Fastly VCL snippets", map[string]any{
		"service_id":      serviceID,
		"service_version": version,
	})

	snippets, err := d.providerData.Client.ListSnippets(ctx, &fastly.ListSnippetsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading VCL snippets", err.Error())
		return
	}

	setVal, ids, diags := flattenVCLSnippets(snippets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.VCLSnippets = setVal
	state.ID = types.StringValue(idhash.HashIDs(ids))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenVCLSnippets(snippets []*fastly.Snippet) (types.Set, []string, diag.Diagnostics) {
	var diags diag.Diagnostics

	ids := make([]string, 0, len(snippets))
	elements := make([]attr.Value, 0, len(snippets))

	for _, snippet := range snippets {
		if snippet == nil {
			continue
		}

		priority, err := parsePriority(snippet.Priority)
		if err != nil {
			diags.AddError("Error parsing VCL snippet priority", err.Error())
			return types.SetNull(types.ObjectType{AttrTypes: snippetAttrTypes}), nil, diags
		}

		id := fastly.ToValue(snippet.SnippetID)
		name := fastly.ToValue(snippet.Name)
		content := fastly.ToValue(snippet.Content)
		snippetType := string(fastly.ToValue(snippet.Type))
		dynamic := fastly.ToValue(snippet.Dynamic) == 1

		ids = append(ids, fmt.Sprintf("%s/%s/%s/%d/%t/%s", id, name, snippetType, priority, dynamic, content))

		obj, objDiags := types.ObjectValue(snippetAttrTypes, map[string]attr.Value{
			"content":  types.StringValue(content),
			"dynamic":  types.BoolValue(dynamic),
			"id":       types.StringValue(id),
			"name":     types.StringValue(name),
			"priority": types.Int64Value(priority),
			"type":     types.StringValue(snippetType),
		})
		diags.Append(objDiags...)
		elements = append(elements, obj)
	}

	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: snippetAttrTypes}), nil, diags
	}

	setVal, setDiags := types.SetValue(types.ObjectType{AttrTypes: snippetAttrTypes}, elements)
	diags.Append(setDiags...)

	return setVal, ids, diags
}

func parsePriority(value *string) (int64, error) {
	if value == nil || *value == "" {
		return defaultPriority, nil
	}

	priority, err := strconv.ParseInt(*value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing VCL snippet priority %q: %w", *value, err)
	}

	return priority, nil
}
