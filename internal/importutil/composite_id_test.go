package importutil

import (
	"testing"
)

func TestParseCompositeID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		wantServiceID string
		wantVersion   int
		wantName      string
		wantErr       bool
	}{
		{
			name:          "valid backend import",
			id:            "service123/3/origin",
			wantServiceID: "service123",
			wantVersion:   3,
			wantName:      "origin",
			wantErr:       false,
		},
		{
			name:          "valid domain import",
			id:            "service123/3/www.example.com",
			wantServiceID: "service123",
			wantVersion:   3,
			wantName:      "www.example.com",
			wantErr:       false,
		},
		{
			name:    "version 0",
			id:      "service123/0/backend1",
			wantErr: true,
		},
		{
			name:          "name with slashes",
			id:            "service123/5/backend/with/slashes",
			wantServiceID: "service123",
			wantVersion:   5,
			wantName:      "backend/with/slashes",
			wantErr:       false,
		},
		{
			name:    "missing parts",
			id:      "service123/3",
			wantErr: true,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "only one part",
			id:      "service123",
			wantErr: true,
		},
		{
			name:    "invalid version",
			id:      "service123/notanumber/backend1",
			wantErr: true,
		},
		{
			name:    "empty service_id",
			id:      "/3/backend1",
			wantErr: true,
		},
		{
			name:    "empty name",
			id:      "service123/3/",
			wantErr: true,
		},
		{
			name:    "negative version",
			id:      "service123/-1/backend1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceID, version, name, err := ParseCompositeID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCompositeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if serviceID != tt.wantServiceID {
				t.Errorf("ParseCompositeID() serviceID = %v, want %v", serviceID, tt.wantServiceID)
			}
			if version != tt.wantVersion {
				t.Errorf("ParseCompositeID() version = %v, want %v", version, tt.wantVersion)
			}
			if name != tt.wantName {
				t.Errorf("ParseCompositeID() name = %v, want %v", name, tt.wantName)
			}
		})
	}
}
