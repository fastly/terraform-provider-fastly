package dynamicsnippet

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	regularsnippet "github.com/fastly/terraform-provider-fastly/internal/resources/snippet"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
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
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Priority  types.Int64  `tfsdk:"priority"`
	SnippetID types.String `tfsdk:"snippet_id"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		normalizeType(service.StringValue(n.Type)) == normalizeType(service.StringValue(other.Type)) &&
		service.Int64Value(n.Priority) == service.Int64Value(other.Priority)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A name that is unique across regular and dynamic VCL snippet configuration blocks. Changing this attribute will delete and recreate the snippet.",
		},
		"type": schema.StringAttribute{
			Required:    true,
			Description: "The location in generated VCL where the dynamic snippet should be placed. Must be one of `init`, `recv`, `hash`, `hit`, `miss`, `pass`, `fetch`, `error`, `deliver`, `log`, or `none`.",
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
		"snippet_id": schema.StringAttribute{
			Computed:    true,
			Description: "The Fastly-generated dynamic snippet ID. Use this value with `fastly_service_dynamic_snippet_content` to manage versionless snippet code.",
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Dynamic VCL snippet metadata attached to this service version. Dynamic snippet content is managed separately by `fastly_service_dynamic_snippet_content`.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

func ValidateConfig(snippets []NestedModel) error {
	seenNames := make(map[string]struct{}, len(snippets))

	for _, item := range snippets {
		if item.Name.IsUnknown() || item.Name.IsNull() {
			continue
		}

		name := strings.TrimSpace(item.Name.ValueString())
		if name == "" {
			return fmt.Errorf("dynamic VCL snippet name cannot be empty")
		}

		if _, ok := seenNames[name]; ok {
			return fmt.Errorf("multiple dynamic snippets with the same name %q; names must be unique within a service version", name)
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
			return fmt.Errorf("dynamic VCL snippet name cannot be empty")
		}

		if _, ok := seenNames[name]; ok {
			return fmt.Errorf("multiple dynamic snippets with the same name %q; names must be unique within a service version", name)
		}
		seenNames[name] = struct{}{}

		if !isValidType(service.StringValue(item.Type)) {
			return fmt.Errorf("invalid dynamic VCL snippet type %q; must be one of %s", service.StringValue(item.Type), strings.Join(validTypes, ", "))
		}
	}

	return nil
}

func ValidateNoNameConflicts(dynamicSnippets []NestedModel, regularSnippets []regularsnippet.NestedModel) error {
	regularByName := make(map[string]struct{}, len(regularSnippets))
	for _, item := range regularSnippets {
		if item.Name.IsUnknown() || item.Name.IsNull() {
			continue
		}
		name := strings.TrimSpace(item.Name.ValueString())
		if name == "" {
			continue
		}
		regularByName[name] = struct{}{}
	}

	for _, item := range dynamicSnippets {
		if item.Name.IsUnknown() || item.Name.IsNull() {
			continue
		}
		name := strings.TrimSpace(item.Name.ValueString())
		if name == "" {
			continue
		}
		if _, ok := regularByName[name]; ok {
			return fmt.Errorf("VCL snippet name %q is used by both regular and dynamic snippets; names must be unique within a service version", name)
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

	dynamic := make([]*fastly.Snippet, 0, len(all))
	for _, item := range all {
		if item == nil {
			continue
		}
		if !regularsnippet.IsDynamic(item) {
			continue
		}
		if _, err := parsePriority(item.Priority); err != nil {
			return nil, err
		}
		dynamic = append(dynamic, item)
	}

	return dynamic, nil
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
	// ops.List validates dynamic-ness and priority before the reconciler calls
	// ToModel, so this should only fail if ToModel is called outside the
	// reconciler read path.
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
	return reconciler.ReadForVersion(ctx, client, serviceID, version)
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

func MatchOrderPreservePlanFields(items, plan []NestedModel) []NestedModel {
	ordered := MatchOrder(items, plan)

	planByName := make(map[string]NestedModel, len(plan))
	for _, item := range plan {
		planByName[service.StringValue(item.Name)] = item
	}

	for i := range ordered {
		name := service.StringValue(ordered[i].Name)
		if planned, ok := planByName[name]; ok {
			// During Create/Update, Terraform requires final state for configured
			// attributes to match the plan. Keep snippet_id from the API because it
			// is computed and is required by the dynamic snippet content resource.
			ordered[i].Name = planned.Name
			ordered[i].Type = planned.Type
			ordered[i].Priority = planned.Priority
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
		return 0, fmt.Errorf("error parsing dynamic VCL snippet priority %q: %w", *value, err)
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

// CopyCommonAttributes returns a copy of the shared attributes. This is useful
// for tests and future explicit resources without allowing callers to mutate the
// package-level schema maps returned by CommonAttributes.
func CopyCommonAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{}
	maps.Copy(attrs, CommonAttributes())
	return attrs
}
