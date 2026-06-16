package apierrors

import (
	"errors"

	fastly "github.com/fastly/go-fastly/v15/fastly"
)

func IsNotFound(err error) bool {
	var httpErr *fastly.HTTPError
	return errors.As(err, &httpErr) && httpErr.StatusCode == 404
}
