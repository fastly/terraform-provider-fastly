package director

import (
	"context"
	"fmt"
	"sort"

	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/reconcile"
	"github.com/fastly/terraform-provider-fastly/internal/resources/backend"
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultComment = ""
	DefaultQuorum  = 75
	DefaultRetries = 5
	DefaultShield  = ""
	DefaultType    = "random"
)

// directorTypeByString maps the string enum to the integer DirectorType the Fastly API expects.
// round_robin is included so Create/Update can round-trip a director that already has that type
// (see directorTypeByAPI) without panicking, even though the type validator below still rejects
// round_robin in new config - matching the legacy provider's validateDirectorType on main.
var directorTypeByString = map[string]fastly.DirectorType{
	"random":      fastly.DirectorTypeRandom,
	"hash":        fastly.DirectorTypeHash,
	"client":      fastly.DirectorTypeClient,
	"round_robin": fastly.DirectorTypeRoundRobin,
}

// directorTypeByAPI maps the integer DirectorType back to its string enum for reading state.
// Unlike directorTypeByString, it includes round_robin: a director created before this schema
// existed (or outside Terraform) can already have that type, and ToModel must represent it
// accurately rather than fall through to the empty string, which would misreport the director's
// actual type and could drive an unintended type change on the next apply.
var directorTypeByAPI = map[fastly.DirectorType]string{
	fastly.DirectorTypeRandom:     "random",
	fastly.DirectorTypeHash:       "hash",
	fastly.DirectorTypeClient:     "client",
	fastly.DirectorTypeRoundRobin: "round_robin",
}

type NestedModel struct {
	Name     types.String `tfsdk:"name"`
	Backends types.Set    `tfsdk:"backends"`
	Comment  types.String `tfsdk:"comment"`
	Quorum   types.Int64  `tfsdk:"quorum"`
	Retries  types.Int64  `tfsdk:"retries"`
	Shield   types.String `tfsdk:"shield"`
	Type     types.String `tfsdk:"type"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.StringValue(n.Name) == service.StringValue(other.Name) &&
		n.Backends.Equal(other.Backends) &&
		service.StringValue(n.Comment) == service.StringValue(other.Comment) &&
		service.Int64Value(n.Quorum) == service.Int64Value(other.Quorum) &&
		service.Int64Value(n.Retries) == service.Int64Value(other.Retries) &&
		service.StringValue(n.Shield) == service.StringValue(other.Shield) &&
		service.StringValue(n.Type) == service.StringValue(other.Type)
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Unique name for this Director.",
		},
		"backends": schema.SetAttribute{
			ElementType: types.StringType,
			Required:    true,
			Description: "Names of defined backends to map the director to. Example: `[\"origin1\", \"origin2\"]`.",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"comment": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultComment),
			Description: "An optional comment about the Director.",
		},
		"quorum": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultQuorum),
			Description: "Percentage of capacity that needs to be up for the director itself to be considered up. Default `75`.",
			Validators: []validator.Int64{
				int64validator.Between(0, 100),
			},
		},
		"retries": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultRetries),
			Description: "How many backends to search if it fails. Default `5`.",
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
		},
		"shield": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultShield),
			Description: "Selected POP to serve as a \"shield\" for backends. Valid values for `shield` are included in the [`GET /datacenters`](https://developer.fastly.com/reference/api/utils/datacenter/) API response.",
		},
		"type": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Type of load balance group to use. One of `random`, `hash`, or `client`. Default `random`.",
			PlanModifiers: []planmodifier.String{
				typeStickyDefault{},
			},
			Validators: []validator.String{
				stringvalidator.OneOf("random", "hash", "client"),
			},
		},
	}
}

func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Directors attached to this service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
	}
}

// typeStickyDefault applies DefaultType only when a director is first created (no prior state).
// A plain Default plan modifier (e.g. stringdefault.StaticString) reapplies its value on every
// plan where config omits the attribute, regardless of prior state - which would silently plan a
// type change back to "random" for a pre-existing round_robin director (see directorTypeByAPI)
// every time type is left unset in config. This modifier instead carries the prior state value
// forward untouched once the director exists, so an out-of-band type survives.
type typeStickyDefault struct{}

func (m typeStickyDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q on create; otherwise preserves the existing value when omitted from config", DefaultType)
}

func (m typeStickyDefault) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m typeStickyDefault) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}
	if req.StateValue.IsNull() {
		resp.PlanValue = types.StringValue(DefaultType)
		return
	}
	resp.PlanValue = req.StateValue
}

// ops holds the remote directors by name produced by the most recent List call within a single
// reconcile run, so Update can diff a director's current backend associations against desired
// without an extra List round-trip. A fresh ops must be used per Reconcile/ReadForVersion call -
// this cache must not be shared across calls for different services/versions.
type ops struct {
	remoteByName map[string]*fastly.Director
}

func (o *ops) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*fastly.Director, error) {
	directors, err := client.ListDirectors(ctx, &fastly.ListDirectorsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return nil, err
	}

	o.remoteByName = make(map[string]*fastly.Director, len(directors))
	for _, d := range directors {
		o.remoteByName[fastly.ToValue(d.Name)] = d
	}

	return directors, nil
}

func (o ops) GetName(api *fastly.Director) string {
	return fastly.ToValue(api.Name)
}

func (o *ops) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	return client.DeleteDirector(ctx, &fastly.DeleteDirectorInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	})
}

func (o ops) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Director, error) {
	name := service.StringValue(desired.Name)

	d, err := client.CreateDirector(ctx, &fastly.CreateDirectorInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(name),
		Comment:        new(service.StringValue(desired.Comment)),
		Shield:         new(service.StringValue(desired.Shield)),
		Quorum:         new(int(service.Int64Value(desired.Quorum))),
		Retries:        new(int(service.Int64Value(desired.Retries))),
		Type:           directorTypePointer(desired.Type),
	})
	if err != nil {
		return nil, err
	}

	for _, backendName := range setToStringSlice(desired.Backends) {
		if _, err := client.CreateDirectorBackend(ctx, &fastly.CreateDirectorBackendInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Director:       name,
			Backend:        backendName,
		}); err != nil {
			return nil, err
		}
	}

	return d, nil
}

func (o ops) Equal(desired NestedModel, remote *fastly.Director) bool {
	return desired.ModelsEqual(o.ToModel(remote))
}

// metadataFieldsEqual compares everything ops.Equal does except Backends, which Update
// reconciles separately via CreateDirectorBackend/DeleteDirectorBackend calls. This lets Update
// skip the UpdateDirector call entirely when a backends-only change is what triggered it.
func (o ops) metadataFieldsEqual(desired NestedModel, remote *fastly.Director) bool {
	m := o.ToModel(remote)
	return service.StringValue(desired.Name) == service.StringValue(m.Name) &&
		service.StringValue(desired.Comment) == service.StringValue(m.Comment) &&
		service.Int64Value(desired.Quorum) == service.Int64Value(m.Quorum) &&
		service.Int64Value(desired.Retries) == service.Int64Value(m.Retries) &&
		service.StringValue(desired.Shield) == service.StringValue(m.Shield) &&
		service.StringValue(desired.Type) == service.StringValue(m.Type)
}

func (o *ops) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired NestedModel) (*fastly.Director, error) {
	name := service.StringValue(desired.Name)

	remote := o.remoteByName[name]

	d := remote
	if remote == nil || !o.metadataFieldsEqual(desired, remote) {
		var err error
		d, err = client.UpdateDirector(ctx, &fastly.UpdateDirectorInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Name:           name,
			Comment:        new(service.StringValue(desired.Comment)),
			Shield:         new(service.StringValue(desired.Shield)),
			Quorum:         new(int(service.Int64Value(desired.Quorum))),
			Retries:        new(int(service.Int64Value(desired.Retries))),
			Type:           directorTypePointer(desired.Type),
		})
		if err != nil {
			return nil, err
		}
	}

	var currentBackends []string
	if remote != nil {
		currentBackends = remote.Backends
	}

	current := make(map[string]struct{}, len(currentBackends))
	for _, b := range currentBackends {
		current[b] = struct{}{}
	}

	desiredBackends := setToStringSlice(desired.Backends)
	wanted := make(map[string]struct{}, len(desiredBackends))
	for _, b := range desiredBackends {
		wanted[b] = struct{}{}
	}

	for _, b := range currentBackends {
		if _, ok := wanted[b]; ok {
			continue
		}
		err := client.DeleteDirectorBackend(ctx, &fastly.DeleteDirectorBackendInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Director:       name,
			Backend:        b,
		})
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
	}

	for _, b := range desiredBackends {
		if _, ok := current[b]; ok {
			continue
		}
		if _, err := client.CreateDirectorBackend(ctx, &fastly.CreateDirectorBackendInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
			Director:       name,
			Backend:        b,
		}); err != nil {
			return nil, err
		}
	}

	return d, nil
}

func (o ops) ToModel(api *fastly.Director) NestedModel {
	return NestedModel{
		Name:     types.StringValue(fastly.ToValue(api.Name)),
		Backends: stringSliceToSet(api.Backends),
		Comment:  types.StringValue(fastly.ToValue(api.Comment)),
		Quorum:   types.Int64Value(int64(fastly.ToValue(api.Quorum))),
		Retries:  types.Int64Value(int64(fastly.ToValue(api.Retries))),
		Shield:   types.StringValue(fastly.ToValue(api.Shield)),
		Type:     types.StringValue(directorTypeByAPI[fastly.ToValue(api.Type)]),
	}
}

// directorTypePointer looks up the API DirectorType for desired.Type. Every value the schema can
// produce - the stringvalidator.OneOf values for new config, plus round_robin read back from a
// director that predates this schema (see directorTypeByAPI) - has an entry in directorTypeByString,
// so the map lookup should never miss; panic rather than silently send the API an undefined
// DirectorType zero-value if that invariant is ever violated.
func directorTypePointer(v types.String) *fastly.DirectorType {
	s := service.StringValue(v)
	t, ok := directorTypeByString[s]
	if !ok {
		panic(fmt.Sprintf("director: unrecognized type %q; must be one of random, hash, client, round_robin", s))
	}
	return &t
}

// setToStringSlice converts a Terraform set of strings into a []string. backends is a required
// attribute, so the set itself is only ever known and non-null here, but an individual element
// can still be unknown (e.g. `backends = [some_resource.x.name]` before x.name is known) - such
// elements are skipped rather than turned into a bogus empty-string backend name.
func setToStringSlice(s types.Set) []string {
	elems := s.Elements()
	parts := make([]string, 0, len(elems))
	for _, e := range elems {
		v, ok := e.(types.String)
		if !ok || v.IsUnknown() || v.IsNull() {
			continue
		}
		parts = append(parts, service.StringValue(v))
	}
	return parts
}

// stringSliceToSet converts the Fastly API's []string into a Terraform set, sorting first so
// state renders deterministically even though set membership - not order - is what matters for
// plan diffing. Duplicates are dropped rather than passed to SetValueMust, which panics on a
// duplicate element; a Set has no duplicates by definition, and a duplicate backend name in the
// API response (e.g. a transient stale-cache read) should degrade gracefully, not crash the
// provider.
func stringSliceToSet(s []string) types.Set {
	sorted := make([]string, len(s))
	copy(sorted, s)
	sort.Strings(sorted)

	elems := make([]attr.Value, 0, len(sorted))
	for i, v := range sorted {
		if i > 0 && v == sorted[i-1] {
			continue
		}
		elems = append(elems, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elems)
}

// newReconciler builds a fresh Resource (and backing ops cache) per call, since ops.remoteByName
// is populated per List call and must not be shared across concurrent reconciles of different
// services/versions.
func newReconciler() *reconcile.Resource[NestedModel, fastly.Director] {
	return &reconcile.Resource[NestedModel, fastly.Director]{
		Ops: &ops{},
		GetName: func(m NestedModel) string {
			return service.StringValue(m.Name)
		},
		Sortable: true,
	}
}

func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]NestedModel, error) {
	return newReconciler().ReadForVersion(ctx, client, serviceID, version)
}

func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []NestedModel) error {
	return newReconciler().Run(ctx, client, serviceID, version, desired)
}

func Equal(a, b []NestedModel) bool {
	return reconcile.ModelsEqual(a, b, func(m NestedModel) string { return service.StringValue(m.Name) }, NestedModel.ModelsEqual, true)
}

func MatchOrder(items, order []NestedModel) []NestedModel {
	return reconcile.MatchOrder(items, order, func(m NestedModel) string { return service.StringValue(m.Name) })
}

// ValidateConfig enforces name uniqueness among a service's directors, catching misconfigurations
// at plan time rather than deferring to a failed apply. reconcile.Run keys remote/desired
// directors by name, so duplicate names would otherwise silently collapse to a single director
// instead of failing at plan time.
func ValidateConfig(directors []NestedModel) error {
	seenNames := make(map[string]struct{}, len(directors))

	for _, item := range directors {
		if item.Name.IsUnknown() || item.Name.IsNull() {
			continue
		}
		name := service.StringValue(item.Name)
		if _, ok := seenNames[name]; ok {
			return fmt.Errorf("multiple directors with the same name %q; names must be unique within a service version", name)
		}
		seenNames[name] = struct{}{}
	}

	return nil
}

// ValidateBackendReferences confirms every backend named in a director's backends matches a
// backend present in the same service config, catching at plan time the case where a backend
// block is renamed or removed but a director's reference to it is left stale. Left unvalidated,
// that reaches the Fastly API as a 404 on the director-backend association instead.
func ValidateBackendReferences(directors []NestedModel, backends []backend.NestedModel) error {
	backendNames := make(map[string]struct{}, len(backends))
	for _, b := range backends {
		if b.Name.IsUnknown() || b.Name.IsNull() {
			continue
		}
		backendNames[service.StringValue(b.Name)] = struct{}{}
	}

	for _, d := range directors {
		if d.Backends.IsUnknown() || d.Backends.IsNull() {
			continue
		}
		name := service.StringValue(d.Name)
		for _, backendName := range setToStringSlice(d.Backends) {
			if _, ok := backendNames[backendName]; !ok {
				return fmt.Errorf("director %q: backend %q does not match any configured backend", name, backendName)
			}
		}
	}

	return nil
}
