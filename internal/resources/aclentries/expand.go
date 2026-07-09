package aclentries

import (
	"context"

	"github.com/fastly/go-fastly/v16/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	createOperation = "create"
	updateOperation = "update"
	deleteOperation = "delete"
)

func expandEntries(ctx context.Context, entries types.Map, diags *diag.Diagnostics) map[string]string {
	if entries.IsNull() || entries.IsUnknown() {
		return nil
	}

	var result map[string]string
	diags.Append(entries.ElementsAs(ctx, &result, false)...)
	return result
}

// buildBatchEntries diffs oldEntries against newEntries and returns the batch
// operations needed to reconcile them. Deletions are only included when
// manage is true, since unmanaged entries left out of newEntries are assumed
// to be intentionally omitted from Terraform's view of the ACL, not removed.
func buildBatchEntries(oldEntries, newEntries map[string]string, manage bool) []*computeacls.BatchComputeACLEntry {
	var batch []*computeacls.BatchComputeACLEntry

	if manage {
		for prefix := range oldEntries {
			if _, ok := newEntries[prefix]; !ok {
				batch = append(batch, &computeacls.BatchComputeACLEntry{
					Prefix:    new(prefix),
					Operation: new(deleteOperation),
				})
			}
		}
	}

	for prefix, action := range newEntries {
		op := createOperation
		if _, ok := oldEntries[prefix]; ok {
			op = updateOperation
		}
		batch = append(batch, &computeacls.BatchComputeACLEntry{
			Prefix:    new(prefix),
			Action:    new(action),
			Operation: new(op),
		})
	}

	return batch
}
