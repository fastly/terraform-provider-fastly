package fastly

import "strings"

func trimSpaceStateFunc(v interface{}) string {
	return strings.TrimSpace(v.(string))
}
