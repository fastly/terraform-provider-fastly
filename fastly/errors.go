package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
)

// HandleNotFoundError provides reusable handling for 404 errors from Fastly API
func HandleNotFoundError(err error, id string, resourceName string) error {
	if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
		log.Printf("[WARN] %s not found for ID (%s)", resourceName, id)
		return fmt.Errorf(
			"%s with ID '%s' was not found.\nThis could mean the resource doesn't exist, or that the Fastly API key used doesn't have the necessary permissions.\nPlease verify the resource ID and API key permissions.",
			resourceName, id,
		)
	}
	return err
}
