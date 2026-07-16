// Package idhash derives stable IDs for list-style data sources from the set
// of resource IDs they return, so the data source's own ID changes whenever
// the underlying set of resources changes.
package idhash

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// HashIDs returns a stable hash of ids, independent of their order.
func HashIDs(ids []string) string {
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)

	sum := sha256.Sum256([]byte(strings.Join(sorted, ",")))
	return hex.EncodeToString(sum[:])
}
