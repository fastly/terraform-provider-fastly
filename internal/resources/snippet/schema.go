package snippet

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const DefaultPriority int64 = 100

var validTypes = []string{
	"init",
	"recv",
	"hash",
	"hit",
	"miss",
	"pass",
	"fetch",
	"error",
	"deliver",
	"log",
	"none",
}

type NestedModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Priority types.Int64  `tfsdk:"priority"`
	Content  types.String `tfsdk:"content"`
}

func IsDynamic(api *fastly.Snippet) bool {
	return api != nil && api.Dynamic != nil && *api.Dynamic == 1
}

func ContentEqual(a, b string) bool {
	return strings.TrimRight(a, "\n") == strings.TrimRight(b, "\n")
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		normalizeType(service.StringValue(n.Type)) == normalizeType(service.StringValue(other.Type)) &&
		service.Int64Value(n.Priority) == service.Int64Value(other.Priority) &&
		ContentEqual(service.StringValue(n.Content), service.StringValue(other.Content))
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A name that is unique across regular and dynamic VCL snippet configuration blocks. Changing this attribute will delete and recreate the snippet.",
		},
		"type": schema.StringAttribute{
			Required:    true,
			Description: "The location in generated VCL where the snippet should be placed. Must be one of `init`, `recv`, `hash`, `hit`, `miss`, `pass`, `fetch`, `error`, `deliver`, `log`, or `none`.",
			Validators: []validator.String{
				stringvalidator.OneOf(validTypes...),
			},
		},
		"priority": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultPriority),
			Description: "Priority determines execution order. Lower numbers execute first. Default `100`.",
		},
		"content": schema.StringAttribute{
			Required:    true,
			Description: "The VCL code that specifies exactly what the snippet does. Can be configured with a quoted string, HEREDOC, file(\"${path.module}/snippet.vcl\"), or templatefile(...).",
		},
	}
}

func ResourceAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Terraform resource identifier.",
		},
		"service_id": schema.StringAttribute{
			Required:    true,
			Description: "Fastly service ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"version": schema.Int64Attribute{
			Required:    true,
			Description: "Writable Fastly service version to modify.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
	}
	maps.Copy(attrs, CommonAttributes())

	nameAttr := attrs["name"].(schema.StringAttribute)
	nameAttr.PlanModifiers = []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
	attrs["name"] = nameAttr

	return attrs
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Regular VCL snippets attached to this service version.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

func ID(serviceID string, version int, name string) string {
	return fmt.Sprintf("%s-%d-%s", serviceID, version, name)
}

func ValidateConfig(snippets []NestedModel) error {
	seenNames := make(map[string]struct{}, len(snippets))

	for _, item := range snippets {
		if item.Name.IsUnknown() || item.Name.IsNull() {
			continue
		}

		name := strings.TrimSpace(item.Name.ValueString())
		if name == "" {
			return fmt.Errorf("VCL snippet name cannot be empty")
		}

		if _, ok := seenNames[name]; ok {
			return fmt.Errorf("multiple snippets with the same name %q; names must be unique within a service version", name)
		}
		seenNames[name] = struct{}{}
	}

	return nil
}

func Validate(snippets []NestedModel) error {
	seenNames := make(map[string]struct{}, len(snippets))

	for _, item := range snippets {
		name := strings.TrimSpace(service.StringValue(item.Name))
		if name == "" {
			return fmt.Errorf("VCL snippet name cannot be empty")
		}

		if _, ok := seenNames[name]; ok {
			return fmt.Errorf("multiple snippets with the same name %q; names must be unique within a service version", name)
		}
		seenNames[name] = struct{}{}

		if !isValidType(service.StringValue(item.Type)) {
			return fmt.Errorf("invalid VCL snippet type %q; must be one of %s", service.StringValue(item.Type), strings.Join(validTypes, ", "))
		}
	}

	return nil
}

type ops struct{}

func (o ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Snippet, error) {
	all, err := client.ListSnippets(ctx, &fastly.ListSnippetsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return nil, err
	}

	regular := make([]*fastly.Snippet, 0, len(all))
	for _, item := range all {
		if item == nil {
			continue
		}
		if IsDynamic(item) {
			continue
		}
		if _, err := parsePriority(item.Priority); err != nil {
			return nil, err
		}
		regular = append(regular, item)
	}

	return regular, nil
}

func (o ops) GetName(api *fastly.Snippet) string {
	return fastly.ToValue(api.Name)
}

func (o ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteSnippet(ctx, &fastly.DeleteSnippetInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Snippet, error) {
	return client.CreateSnippet(ctx, BuildCreateInput(serviceID, version, desired))
}

func (o ops) Equal(desired NestedModel, remote *fastly.Snippet) bool {
	remoteModel, err := FlattenToNestedModel(remote)
	if err != nil {
		return false
	}
	return desired.ModelsEqual(remoteModel)
}

func (o ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Snippet, error) {
	return client.UpdateSnippet(ctx, BuildUpdateInput(serviceID, version, desired))
}

func (o ops) ToModel(api *fastly.Snippet) NestedModel {
	model, _ := FlattenToNestedModel(api)
	return model
}

var reconciler = &reconcile.Resource[NestedModel, fastly.Snippet]{
	Ops: ops{},
	GetName: func(m NestedModel) string {
		return service.StringValue(m.Name)
	},
	Sortable: true,
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	remote, err := ops{}.List(ctx, client, serviceID, version)
	if err != nil {
		return nil, err
	}

	result := make([]NestedModel, 0, len(remote))
	for _, item := range remote {
		model, err := FlattenToNestedModel(item)
		if err != nil {
			return nil, err
		}
		result = append(result, model)
	}

	sort.Slice(result, func(i, j int) bool {
		return service.StringValue(result[i].Name) < service.StringValue(result[j].Name)
	})

	return result, nil
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	if err := Validate(desired); err != nil {
		return err
	}
	return reconciler.Run(ctx, client, serviceID, version, desired)
}

func Equal(a, b []NestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.Name) }, NestedModel.ModelsEqual, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	return reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
}

func MatchOrderPreservePlanContent(items, plan []NestedModel) []NestedModel {
	ordered := MatchOrder(items, plan)

	planByName := make(map[string]NestedModel, len(plan))
	for _, item := range plan {
		planByName[service.StringValue(item.Name)] = item
	}

	for i := range ordered {
		name := service.StringValue(ordered[i].Name)
		if planned, ok := planByName[name]; ok {
			// During Create/Update, Terraform requires the final state for configured
			// attributes to match the planned values. The Fastly API can normalize or
			// briefly return stale snippet fields immediately after updating a cloned
			// version, so preserve the configured values for the post-apply state.
			//
			// Read paths still use MatchOrder directly and reflect remote API values.
			ordered[i].Type = planned.Type
			ordered[i].Priority = planned.Priority
			ordered[i].Content = planned.Content
		}
	}

	return ordered
}

func parsePriority(value *string) (int64, error) {
	if value == nil || *value == "" {
		return DefaultPriority, nil
	}

	priority, err := strconv.ParseInt(*value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing VCL snippet priority %q: %w", *value, err)
	}

	return priority, nil
}

func normalizeType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isValidType(value string) bool {
	value = normalizeType(value)
	for _, valid := range validTypes {
		if value == valid {
			return true
		}
	}
	return false
}
