package reconcile

import (
	"context"
	"sort"

	"github.com/fastly/terraform-provider-fastly/internal/errors"

	"github.com/fastly/go-fastly/v15/fastly"
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
	remote, err := r.Ops.List(ctx, client, serviceID, version)
	if err != nil {
		return err
	}

	if r.Sortable {
		sort.Slice(desired, func(i, j int) bool {
			return r.GetName(desired[i]) < r.GetName(desired[j])
		})
	}

	remoteByName := make(map[string]*API, len(remote))
	for _, item := range remote {
		remoteByName[r.Ops.GetName(item)] = item
	}

	desiredByName := make(map[string]T, len(desired))
	for _, item := range desired {
		desiredByName[r.GetName(item)] = item
	}

	for name := range remoteByName {
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
