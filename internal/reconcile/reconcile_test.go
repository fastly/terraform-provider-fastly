package reconcile

import (
	"context"
	"errors"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/stretchr/testify/assert"
)

type testModel struct {
	Name  string
	Value string
}

type testAPI struct {
	Name  string
	Value string
}

type testOps struct {
	listResult   []*testAPI
	listError    error
	deleteError  error
	createError  error
	updateError  error
	deletedItems []string
	createdItems []testModel
	updatedItems []testModel
}

func (o *testOps) List(ctx context.Context, client *fastly.Client, serviceID string, version int) ([]*testAPI, error) {
	if o.listError != nil {
		return nil, o.listError
	}
	return o.listResult, nil
}

func (o *testOps) GetName(api *testAPI) string {
	return api.Name
}

func (o *testOps) Delete(ctx context.Context, client *fastly.Client, serviceID string, version int, name string) error {
	if o.deleteError != nil {
		return o.deleteError
	}
	o.deletedItems = append(o.deletedItems, name)
	return nil
}

func (o *testOps) Create(ctx context.Context, client *fastly.Client, serviceID string, version int, desired testModel) (*testAPI, error) {
	if o.createError != nil {
		return nil, o.createError
	}
	o.createdItems = append(o.createdItems, desired)
	return &testAPI{Name: desired.Name, Value: desired.Value}, nil
}

func (o *testOps) Equal(desired testModel, remote *testAPI) bool {
	return desired.Name == remote.Name && desired.Value == remote.Value
}

func (o *testOps) Update(ctx context.Context, client *fastly.Client, serviceID string, version int, desired testModel) (*testAPI, error) {
	if o.updateError != nil {
		return nil, o.updateError
	}
	o.updatedItems = append(o.updatedItems, desired)
	return &testAPI{Name: desired.Name, Value: desired.Value}, nil
}

func (o *testOps) ToModel(api *testAPI) testModel {
	return testModel{Name: api.Name, Value: api.Value}
}

func TestResource_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("creates new items", func(t *testing.T) {
		ops := &testOps{
			listResult: []*testAPI{},
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		desired := []testModel{
			{Name: "item1", Value: "value1"},
			{Name: "item2", Value: "value2"},
		}

		err := r.Run(ctx, nil, "service123", 1, desired)
		assert.NoError(t, err)
		assert.Len(t, ops.createdItems, 2)
		assert.Equal(t, "item1", ops.createdItems[0].Name)
		assert.Equal(t, "item2", ops.createdItems[1].Name)
	})

	t.Run("deletes removed items", func(t *testing.T) {
		ops := &testOps{
			listResult: []*testAPI{
				{Name: "item1", Value: "value1"},
				{Name: "item2", Value: "value2"},
			},
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		desired := []testModel{
			{Name: "item1", Value: "value1"},
		}

		err := r.Run(ctx, nil, "service123", 1, desired)
		assert.NoError(t, err)
		assert.Len(t, ops.deletedItems, 1)
		assert.Equal(t, "item2", ops.deletedItems[0])
	})

	t.Run("updates changed items", func(t *testing.T) {
		ops := &testOps{
			listResult: []*testAPI{
				{Name: "item1", Value: "oldvalue"},
			},
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		desired := []testModel{
			{Name: "item1", Value: "newvalue"},
		}

		err := r.Run(ctx, nil, "service123", 1, desired)
		assert.NoError(t, err)
		assert.Len(t, ops.updatedItems, 1)
		assert.Equal(t, "item1", ops.updatedItems[0].Name)
		assert.Equal(t, "newvalue", ops.updatedItems[0].Value)
	})

	t.Run("skips unchanged items", func(t *testing.T) {
		ops := &testOps{
			listResult: []*testAPI{
				{Name: "item1", Value: "value1"},
			},
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		desired := []testModel{
			{Name: "item1", Value: "value1"},
		}

		err := r.Run(ctx, nil, "service123", 1, desired)
		assert.NoError(t, err)
		assert.Len(t, ops.updatedItems, 0)
		assert.Len(t, ops.createdItems, 0)
		assert.Len(t, ops.deletedItems, 0)
	})

	t.Run("handles list error", func(t *testing.T) {
		ops := &testOps{
			listError: errors.New("list failed"),
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		err := r.Run(ctx, nil, "service123", 1, []testModel{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list failed")
	})

	t.Run("handles delete error", func(t *testing.T) {
		ops := &testOps{
			listResult: []*testAPI{
				{Name: "item1", Value: "value1"},
			},
			deleteError: errors.New("delete failed"),
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		err := r.Run(ctx, nil, "service123", 1, []testModel{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("handles create error", func(t *testing.T) {
		ops := &testOps{
			listResult:  []*testAPI{},
			createError: errors.New("create failed"),
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		desired := []testModel{
			{Name: "item1", Value: "value1"},
		}

		err := r.Run(ctx, nil, "service123", 1, desired)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create failed")
	})

	t.Run("handles update error", func(t *testing.T) {
		ops := &testOps{
			listResult: []*testAPI{
				{Name: "item1", Value: "oldvalue"},
			},
			updateError: errors.New("update failed"),
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		desired := []testModel{
			{Name: "item1", Value: "newvalue"},
		}

		err := r.Run(ctx, nil, "service123", 1, desired)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
	})
}

func TestResource_ReadForVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("reads and converts items", func(t *testing.T) {
		ops := &testOps{
			listResult: []*testAPI{
				{Name: "item1", Value: "value1"},
				{Name: "item2", Value: "value2"},
			},
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		result, err := r.ReadForVersion(ctx, nil, "service123", 1)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "item1", result[0].Name)
		assert.Equal(t, "item2", result[1].Name)
	})

	t.Run("handles empty list", func(t *testing.T) {
		ops := &testOps{
			listResult: []*testAPI{},
		}
		r := &Resource[testModel, testAPI]{
			Ops:      ops,
			GetName:  func(m testModel) string { return m.Name },
			Sortable: true,
		}

		result, err := r.ReadForVersion(ctx, nil, "service123", 1)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestMatchOrder(t *testing.T) {
	t.Run("orders items to match reference order", func(t *testing.T) {
		items := []testModel{
			{Name: "a", Value: "remote-a"},
			{Name: "b", Value: "remote-b"},
		}
		order := []testModel{
			{Name: "b"},
			{Name: "a"},
		}

		result := MatchOrder(items, order, func(m testModel) string { return m.Name })

		assert.Equal(t, []testModel{
			{Name: "b", Value: "remote-b"},
			{Name: "a", Value: "remote-a"},
		}, result)
	})

	t.Run("appends unmatched items in stable name order", func(t *testing.T) {
		items := []testModel{
			{Name: "c", Value: "remote-c"},
			{Name: "a", Value: "remote-a"},
			{Name: "b", Value: "remote-b"},
		}
		order := []testModel{
			{Name: "b"},
		}

		result := MatchOrder(items, order, func(m testModel) string { return m.Name })

		assert.Equal(t, []testModel{
			{Name: "b", Value: "remote-b"},
			{Name: "a", Value: "remote-a"},
			{Name: "c", Value: "remote-c"},
		}, result)
	})

	t.Run("skips ordered items not present in result", func(t *testing.T) {
		items := []testModel{
			{Name: "b", Value: "remote-b"},
		}
		order := []testModel{
			{Name: "missing"},
			{Name: "b"},
		}

		result := MatchOrder(items, order, func(m testModel) string { return m.Name })

		assert.Equal(t, []testModel{
			{Name: "b", Value: "remote-b"},
		}, result)
	})
}

func TestModelsEqual(t *testing.T) {
	t.Run("equal when same", func(t *testing.T) {
		a := []testModel{{Name: "item1", Value: "value1"}}
		b := []testModel{{Name: "item1", Value: "value1"}}

		result := ModelsEqual(a, b,
			func(m testModel) string { return m.Name },
			func(a, b testModel) bool { return a.Name == b.Name && a.Value == b.Value },
			true)

		assert.True(t, result)
	})

	t.Run("equal when different order but sortable", func(t *testing.T) {
		a := []testModel{{Name: "item2", Value: "value2"}, {Name: "item1", Value: "value1"}}
		b := []testModel{{Name: "item1", Value: "value1"}, {Name: "item2", Value: "value2"}}

		result := ModelsEqual(a, b,
			func(m testModel) string { return m.Name },
			func(a, b testModel) bool { return a.Name == b.Name && a.Value == b.Value },
			true)

		assert.True(t, result)
	})

	t.Run("not equal when different lengths", func(t *testing.T) {
		a := []testModel{{Name: "item1", Value: "value1"}}
		b := []testModel{{Name: "item1", Value: "value1"}, {Name: "item2", Value: "value2"}}

		result := ModelsEqual(a, b,
			func(m testModel) string { return m.Name },
			func(a, b testModel) bool { return a.Name == b.Name && a.Value == b.Value },
			true)

		assert.False(t, result)
	})

	t.Run("not equal when different values", func(t *testing.T) {
		a := []testModel{{Name: "item1", Value: "value1"}}
		b := []testModel{{Name: "item1", Value: "value2"}}

		result := ModelsEqual(a, b,
			func(m testModel) string { return m.Name },
			func(a, b testModel) bool { return a.Name == b.Name && a.Value == b.Value },
			true)

		assert.False(t, result)
	})
}
