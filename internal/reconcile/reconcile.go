package reconcile

import (
	"context"
	"fmt"
	"sort"

	"github.com/fastly/terraform-provider-fastly/internal/errors"

	"github.com/fastly/go-fastly/v17/fastly"
)

type Ops[T any, API any] interface {
	List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*API, error)
	GetName(api *API) string
	Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error
	Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired T) (*API, error)
	Equal(desired T, remote *API) bool
	Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired T) (*API, error)
	ToModel(api *API) T
}

type Resource[T any, API any] struct {
	Ops      Ops[T, API]
	GetName  func(T) string
	Sortable bool
}

func (r *Resource[T, API]) Run(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []T) error {
	if err := r.DeleteRemoved(ctx, client, serviceID, version, desired); err != nil {
		return err
	}
	return r.CreateOrUpdate(ctx, client, serviceID, version, desired)
}

// DeleteRemoved deletes remote items missing from desired. It is split out from Run so a
// caller can defer deletion until after some other resource type has been reconciled, when a
// remote item might still be referenced by name from that other resource type (e.g. a rate
// limiter's uri_dictionary_name) - see GuardedRun's doc comment for the same concern on the
// guarded path.
func (r *Resource[T, API]) DeleteRemoved(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []T) error {
	remote, err := r.Ops.List(ctx, client, serviceID, version)
	if err != nil {
		return err
	}

	desiredByName := make(map[string]struct{}, len(desired))
	for _, item := range desired {
		desiredByName[r.GetName(item)] = struct{}{}
	}

	for _, item := range remote {
		name := r.Ops.GetName(item)
		if _, ok := desiredByName[name]; !ok {
			err := r.Ops.Delete(ctx, client, serviceID, version, name)
			if errors.IsNotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateOrUpdate creates or updates every item in desired; it never deletes. It is split out
// from Run for the same reason as DeleteRemoved.
func (r *Resource[T, API]) CreateOrUpdate(ctx context.Context, client *fastly.Client, serviceID string, version int, desired []T) error {
	remote, err := r.Ops.List(ctx, client, serviceID, version)
	if err != nil {
		return err
	}

	if r.Sortable {
		sorted := make([]T, len(desired))
		copy(sorted, desired)
		sort.Slice(sorted, func(i, j int) bool {
			return r.GetName(sorted[i]) < r.GetName(sorted[j])
		})
		desired = sorted
	}

	remoteByName := make(map[string]*API, len(remote))
	for _, item := range remote {
		remoteByName[r.Ops.GetName(item)] = item
	}

	for _, desiredItem := range desired {
		name := r.GetName(desiredItem)
		remoteItem, exists := remoteByName[name]

		if !exists {
			if _, err := r.Ops.Create(ctx, client, serviceID, version, desiredItem); err != nil {
				return err
			}
			continue
		}

		if !r.Ops.Equal(desiredItem, remoteItem) {
			if _, err := r.Ops.Update(ctx, client, serviceID, version, desiredItem); err != nil {
				return err
			}
		}
	}

	return nil
}

// GuardedRun applies desired against remote like Run, but first refuses the whole reconcile if
// it would silently throw away live, unrecoverable API-side data attached to a previously
// existing item - e.g. ACL entries, or dictionary items. Several resource types wrap a
// versioned config container (the ACL or dictionary itself) around such data; removing the item
// from config is the obvious destructive transition, but a resource can have others too (a
// dictionary's write_only toggle is implemented via delete-then-create, which is just as
// destructive as removal). discardsItems decides, per previously-existing item, whether the
// transition to its (possibly absent) desired counterpart is one of those destructive
// transitions; when it is, the item must either have forceDestroy set, or isEmpty must confirm
// there's nothing left to lose.
func (r *Resource[T, API]) GuardedRun(
	ctx context.Context,
	client *fastly.Client,
	serviceID string,
	version int,
	previous, desired []T,
	forceDestroy func(T) bool,
	discardsItems func(prev, desired T, stillPresent bool) bool,
	isEmpty func(ctx context.Context, prev T) (bool, error),
	notEmptyErr func(name string, prev T) error,
) error {
	if err := r.CheckGuards(ctx, previous, desired, forceDestroy, discardsItems, isEmpty, notEmptyErr); err != nil {
		return err
	}

	return r.Run(ctx, client, serviceID, version, desired)
}

// CheckGuards is the validation half of GuardedRun, split out so a caller can defer the actual
// deletion (via DeleteRemoved, called separately) until after some other resource type - which
// might still reference a removed item by name - has been reconciled. It performs no changes;
// it only refuses the transition if it would silently discard live, unrecoverable API-side data.
func (r *Resource[T, API]) CheckGuards(
	ctx context.Context,
	previous, desired []T,
	forceDestroy func(T) bool,
	discardsItems func(prev, desired T, stillPresent bool) bool,
	isEmpty func(ctx context.Context, prev T) (bool, error),
	notEmptyErr func(name string, prev T) error,
) error {
	previousByName := make(map[string]T, len(previous))
	for _, p := range previous {
		previousByName[r.GetName(p)] = p
	}

	desiredByName := make(map[string]T, len(desired))
	for _, d := range desired {
		desiredByName[r.GetName(d)] = d
	}

	for name, prevItem := range previousByName {
		desiredItem, stillPresent := desiredByName[name]
		if !discardsItems(prevItem, desiredItem, stillPresent) || forceDestroy(prevItem) {
			continue
		}

		empty, err := isEmpty(ctx, prevItem)
		if err != nil {
			return fmt.Errorf("error checking if %q is empty before removal: %w", name, err)
		}

		if !empty {
			return notEmptyErr(name, prevItem)
		}
	}

	return nil
}

func (r *Resource[T, API]) ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]T, error) {
	remote, err := r.Ops.List(ctx, client, serviceID, version)
	if err != nil {
		return nil, err
	}

	result := make([]T, 0, len(remote))
	for _, item := range remote {
		result = append(result, r.Ops.ToModel(item))
	}

	if r.Sortable {
		sort.Slice(result, func(i, j int) bool {
			return r.GetName(result[i]) < r.GetName(result[j])
		})
	}

	return result, nil
}

// MatchOrder returns items ordered to match the names in order.
// Items with names not present in order are appended in stable name order.
func MatchOrder[T any](items []T, order []T, getName func(T) string) []T {
	itemsByName := make(map[string]T, len(items))
	for _, item := range items {
		itemsByName[getName(item)] = item
	}

	result := make([]T, 0, len(items))
	seen := make(map[string]struct{}, len(items))

	for _, orderedItem := range order {
		name := getName(orderedItem)
		item, ok := itemsByName[name]
		if !ok {
			continue
		}
		result = append(result, item)
		seen[name] = struct{}{}
	}

	unmatched := make([]T, 0, len(items))
	for _, item := range items {
		if _, ok := seen[getName(item)]; ok {
			continue
		}
		unmatched = append(unmatched, item)
	}

	sort.Slice(unmatched, func(i, j int) bool {
		return getName(unmatched[i]) < getName(unmatched[j])
	})

	result = append(result, unmatched...)
	return result
}

func ModelsEqual[T any](a, b []T, getName func(T) string, equals func(T, T) bool, sortable bool) bool {
	if sortable {
		sortedA := make([]T, len(a))
		sortedB := make([]T, len(b))
		copy(sortedA, a)
		copy(sortedB, b)

		sort.Slice(sortedA, func(i, j int) bool {
			return getName(sortedA[i]) < getName(sortedA[j])
		})
		sort.Slice(sortedB, func(i, j int) bool {
			return getName(sortedB[i]) < getName(sortedB[j])
		})

		a = sortedA
		b = sortedB
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !equals(a[i], b[i]) {
			return false
		}
	}

	return true
}
