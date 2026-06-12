package importutil

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseCompositeID parses a legacy composite import ID in the format "service_id/version/name"
// and returns the individual components. This supports backwards compatibility for resources
// that were imported with version information before the identity schema was updated.
func ParseCompositeID(id string) (serviceID string, version int, name string, err error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 {
		return "", 0, "", fmt.Errorf("invalid composite import ID format: expected service_id/version/name, got %q", id)
	}

	serviceID = parts[0]
	if serviceID == "" {
		return "", 0, "", fmt.Errorf("service_id cannot be empty in import ID %q", id)
	}

	version, err = strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid version number in import ID %q: %w", id, err)
	}

	if version < 1 {
		return "", 0, "", fmt.Errorf("version must be greater than 0 in import ID %q", id)
	}

	name = parts[2]
	if name == "" {
		return "", 0, "", fmt.Errorf("name cannot be empty in import ID %q", id)
	}

	return serviceID, version, name, nil
}
