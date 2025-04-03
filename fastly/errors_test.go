package fastly

import (
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/stretchr/testify/assert"
)

func TestHandleNotFoundError_NotFound(t *testing.T) {
	err := &gofastly.HTTPError{
		StatusCode: 404,
		Errors: []*gofastly.ErrorObject{
			{
				Code: "404",
			},
		},
	}
	resultErr := HandleNotFoundError(err, "c04ef", "Fastly Service")

	assert.Error(t, resultErr, "Expected an error to be returned")
	assert.Equal(t, resultErr.Error(), "Fastly Service with ID 'c04ef' was not found.\nThis could mean the resource doesn't exist, or that the Fastly API key used doesn't have the necessary permissions.\nPlease verify the resource ID and API key permissions.", "Expected detailed error message")
}

func TestHandleNotFoundError_OtherError(t *testing.T) {
	err := &gofastly.HTTPError{
		StatusCode: 500,
		Errors: []*gofastly.ErrorObject{
			{
				Code:   "500",
				Title:  "Internal server error",
				Detail: "This should be returned now",
			},
		},
	}
	resultErr := HandleNotFoundError(err, "c04ef", "Fastly Service")

	assert.Error(t, resultErr, "Expected an error to be returned")
	assert.Equal(t, err, resultErr, "Expected the original error to propagate")
}
