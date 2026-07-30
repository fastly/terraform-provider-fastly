package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMetadataChanged(t *testing.T) {
	tests := []struct {
		name         string
		planName     types.String
		planComment  types.String
		stateName    types.String
		stateComment types.String
		want         bool
	}{
		{
			name:         "unchanged metadata",
			planName:     types.StringValue("service"),
			planComment:  types.StringValue("comment"),
			stateName:    types.StringValue("service"),
			stateComment: types.StringValue("comment"),
			want:         false,
		},
		{
			name:         "name changed",
			planName:     types.StringValue("updated-service"),
			planComment:  types.StringValue("comment"),
			stateName:    types.StringValue("service"),
			stateComment: types.StringValue("comment"),
			want:         true,
		},
		{
			name:         "comment changed",
			planName:     types.StringValue("service"),
			planComment:  types.StringValue("updated comment"),
			stateName:    types.StringValue("service"),
			stateComment: types.StringValue("comment"),
			want:         true,
		},
		{
			name:         "both changed",
			planName:     types.StringValue("updated-service"),
			planComment:  types.StringValue("updated comment"),
			stateName:    types.StringValue("service"),
			stateComment: types.StringValue("comment"),
			want:         true,
		},
		{
			name:         "null comments unchanged",
			planName:     types.StringValue("service"),
			planComment:  types.StringNull(),
			stateName:    types.StringValue("service"),
			stateComment: types.StringNull(),
			want:         false,
		},
		{
			name:         "null comment changed to empty",
			planName:     types.StringValue("service"),
			planComment:  types.StringValue(""),
			stateName:    types.StringValue("service"),
			stateComment: types.StringNull(),
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MetadataChanged(tt.planName, tt.planComment, tt.stateName, tt.stateComment)
			if got != tt.want {
				t.Fatalf("MetadataChanged() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestUpdateMetadataIfChanged(t *testing.T) {
	tests := []struct {
		name         string
		planName     types.String
		planComment  types.String
		stateName    types.String
		stateComment types.String
		wantRequests int
	}{
		{
			name:         "unchanged metadata skips UpdateService",
			planName:     types.StringValue("service"),
			planComment:  types.StringValue("comment"),
			stateName:    types.StringValue("service"),
			stateComment: types.StringValue("comment"),
			wantRequests: 0,
		},
		{
			name:         "changed name calls UpdateService",
			planName:     types.StringValue("updated-service"),
			planComment:  types.StringValue("comment"),
			stateName:    types.StringValue("service"),
			stateComment: types.StringValue("comment"),
			wantRequests: 1,
		},
		{
			name:         "changed comment calls UpdateService",
			planName:     types.StringValue("service"),
			planComment:  types.StringValue("updated comment"),
			stateName:    types.StringValue("service"),
			stateComment: types.StringValue("comment"),
			wantRequests: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &countingTransport{}
			client, err := fastly.NewClient("test-token")
			if err != nil {
				t.Fatalf("fastly.NewClient() error = %v", err)
			}
			client.HTTPClient = &http.Client{Transport: transport}

			err = UpdateMetadataIfChanged(
				context.Background(),
				client,
				"test-service",
				tt.planName,
				tt.planComment,
				tt.stateName,
				tt.stateComment,
			)
			if err != nil {
				t.Fatalf("UpdateMetadataIfChanged() unexpected error = %v", err)
			}

			if transport.requests != tt.wantRequests {
				t.Fatalf("UpdateService requests = %d, want %d", transport.requests, tt.wantRequests)
			}
		})
	}
}

type countingTransport struct {
	requests int
}

func (t *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.requests++

	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Request:    req,
	}, nil
}
