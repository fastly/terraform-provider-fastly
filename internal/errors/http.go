package errors

import (
	"errors"

	fastly "github.com/fastly/go-fastly/v16/fastly"
)

func IsNotFound(err error) bool {
	var httpErr *fastly.HTTPError
	return errors.As(err, &httpErr) && httpErr.StatusCode == 404
}
