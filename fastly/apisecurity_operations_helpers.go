package fastly

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const apiSecurityDefaultPageLimit = 100

// parseTwoPartImportID parses IDs in the format: <service_id>/<object_id>.
func parseTwoPartImportID(id string) (serviceID string, objectID string, err error) {
	left, right, ok := strings.Cut(id, "/")
	if !ok || left == "" || right == "" {
		return "", "", fmt.Errorf("invalid ID format: %q. Expected format: <service_id>/<id>", id)
	}
	return left, right, nil
}

// expandStringSet converts a schema.Set to a sorted []string.
func expandStringSet(set *schema.Set) []string {
	if set == nil {
		return nil
	}

	raw := set.List()
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// flattenStringSliceToSet converts []string to *schema.Set.
func flattenStringSliceToSet(in []string) *schema.Set {
	out := make([]any, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}
	return schema.NewSet(schema.HashString, out)
}
