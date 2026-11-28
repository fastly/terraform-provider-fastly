package fastly

import (
	"context"
	"strconv"
	"sync/atomic"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
)

var anonymousResourceLockSeq atomic.Uint64

// withResourceLock returns a context that tells go-fastly which resource a
// mutating request targets.
//
// go-fastly serializes non-GET requests that share a resource ID. Requests
// without an ID all fall back to one shared "unknown" key. That forces every
// such request in the provider process to run one at a time, even when it
// targets an unrelated resource, so Terraform's -parallelism has no effect.
//
// If id is empty (for example, when a resource is being created and has no ID
// yet), a unique key is generated. The request then does not wait on unrelated
// requests.
func withResourceLock(ctx context.Context, id string) context.Context {
	if id == "" {
		id = "anonymous-" + strconv.FormatUint(anonymousResourceLockSeq.Add(1), 10)
	}
	return gofastly.NewContextForResourceID(ctx, id)
}
