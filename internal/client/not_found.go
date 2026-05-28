package client

import fastly "github.com/fastly/go-fastly/v15/fastly"

func IsNotFound(err error) bool {
	httpErr, ok := err.(*fastly.HTTPError)
	return ok && httpErr.StatusCode == 404
}
