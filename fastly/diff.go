package fastly

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// KeyFunc calculates a key from an element.
type KeyFunc func(any) (any, error)

// SetDiff diffs two sets using a key to identify which elements have been added, changed, removed or not modified.
//
// This object compares sets using Terraform's schema.Set methods (e.g. Difference() and Intersection())
// so that the same differences displayed to the user are honoured here.
//
// SetDiff however is able to tell if two elements from two distinct sets have the same key. This is useful to detect
// that an element should be updated instead of recreated on the remote server.
type SetDiff struct {
	keyFunc KeyFunc
}

// DiffResult contains the differences between two sets.
type DiffResult struct {
	Added      []any
	Modified   []any
	Deleted    []any
	Unmodified []any
}

// NewSetDiff creates a new SetDiff with a provided KeyFunc.
func NewSetDiff(keyFunc KeyFunc) *SetDiff {
	return &SetDiff{
		keyFunc: keyFunc,
	}
}

// Diff diffs two Set objects and returns a DiffResult object containing the diffs.
//
// The DiffResult object will contain the elements from newSet on the Modified field.
//
// NOTE: there is a caveat with the current implementation which is related to
// the lookup 'key' you specify. If the key you use (to lookup a resource
// within the comparable set) is also updatable via the fastly API, then that
// means you'll end up deleting and recreating the resource rather than simply
// updating it (which is less efficient, as it's two separate operations).
//
// For example, a 'domain' can be updated by changing either its 'name' or its
// 'comment' attribute, but in order to compare changes using SetDiff we only
// really have the option to use 'name' as the lookup key.
func (h *SetDiff) Diff(oldSet, newSet *schema.Set) (*DiffResult, error) {
	// Convert the set into a map to facilitate lookup
	oldSetMap := map[any]any{}
	newSetMap := map[any]any{}

	for _, elem := range oldSet.List() {
		key, err := h.computeKey(elem)
		if err != nil {
			return nil, newElementKeyError(elem, err)
		}
		oldSetMap[key] = elem
	}

	for _, elem := range newSet.List() {
		key, err := h.computeKey(elem)
		if err != nil {
			return nil, newElementKeyError(elem, err)
		}
		newSetMap[key] = elem
	}

	// Compute all added and modified elements using Terraform set comparison method.
	var added, modified []any
	for _, newElem := range newSet.Difference(oldSet).List() {
		key, err := h.computeKey(newElem)
		if err != nil {
			return nil, newElementKeyError(newElem, err)
		}
		if oldSetMap[key] != nil {
			modified = append(modified, newElem)
		} else {
			added = append(added, newElem)
		}
	}

	var deleted []any
	for _, oldElem := range oldSet.Difference(newSet).List() {
		key, err := h.computeKey(oldElem)
		if err != nil {
			return nil, newElementKeyError(oldElem, err)
		}
		if newSetMap[key] == nil {
			deleted = append(deleted, oldElem)
		}
	}

	unmodified := oldSet.Intersection(newSet).List()

	return &DiffResult{
		Added:      added,
		Modified:   modified,
		Deleted:    deleted,
		Unmodified: unmodified,
	}, nil
}

func (h *SetDiff) computeKey(elem any) (any, error) {
	key, err := h.keyFunc(elem)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, fmt.Errorf("invalid key for element %v, %v", elem, err)
	}
	return key, nil
}

// Filter filters out unmodified fields of a Set elements map data structure by ranging over
// the original data and comparing each field against the new data.
//
// The motivation for this function is to avoid resetting an attribute on a
// resource to a value that hasn't actually changed because (depending on the
// attribute) it might have unexpected consequences (e.g. a nested resource
// gets replaced/recreated). Safer to only update attributes that need to be.
func (h *SetDiff) Filter(modified map[string]any, oldSet *schema.Set) map[string]any {
	elements := oldSet.List()
	filtered := make(map[string]any)

	for _, e := range elements {
		m := e.(map[string]any)

		if m["name"].(string) == modified["name"].(string) {
			for k, v := range m {
				if !reflect.DeepEqual(v, modified[k]) {
					filtered[k] = modified[k]
				}
			}
		}
	}

	return filtered
}

func newElementKeyError(elem any, err error) error {
	return fmt.Errorf("error computing the key for element %v, %v", elem, err)
}
