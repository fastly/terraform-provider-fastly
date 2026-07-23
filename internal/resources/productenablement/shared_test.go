package productenablement

import (
	"errors"
	"net/http"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/stretchr/testify/assert"
)

func TestIsEntitlementError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-HTTP error",
			err:  errors.New("boom"),
			want: false,
		},
		{
			name: "not entitled to disable",
			err: &fastly.HTTPError{
				StatusCode: http.StatusBadRequest,
				Errors:     []*fastly.ErrorObject{{Title: "user is not entitled to disable this product"}},
			},
			want: true,
		},
		{
			name: "product cannot be disabled",
			err: &fastly.HTTPError{
				StatusCode: http.StatusBadRequest,
				Errors:     []*fastly.ErrorObject{{Title: "product cannot be disabled"}},
			},
			want: true,
		},
		{
			name: "cannot self-disable",
			err: &fastly.HTTPError{
				StatusCode: http.StatusBadRequest,
				Errors:     []*fastly.ErrorObject{{Title: "cannot self-disable this product"}},
			},
			want: true,
		},
		{
			name: "unrelated 400",
			err: &fastly.HTTPError{
				StatusCode: http.StatusBadRequest,
				Errors:     []*fastly.ErrorObject{{Title: "invalid input"}},
			},
			want: false,
		},
		{
			name: "wrong status code",
			err: &fastly.HTTPError{
				StatusCode: http.StatusInternalServerError,
				Errors:     []*fastly.ErrorObject{{Title: "not entitled to disable"}},
			},
			want: false,
		},
		{
			name: "wrapped HTTPError",
			err: errWrap{
				err: &fastly.HTTPError{
					StatusCode: http.StatusBadRequest,
					Errors:     []*fastly.ErrorObject{{Title: "not entitled to disable"}},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isEntitlementError(tt.err))
		})
	}
}

// errWrap wraps an error to exercise errors.As unwrapping in isEntitlementError.
type errWrap struct{ err error }

func (e errWrap) Error() string { return e.err.Error() }
func (e errWrap) Unwrap() error { return e.err }
