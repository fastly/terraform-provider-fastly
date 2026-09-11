package fastly

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"
	arcprovider "github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/provider"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"
)

func TestDiffSuppressEquivalentRFC3339(t *testing.T) {
	for _, tc := range []struct {
		name string
		old  string
		new  string
		want bool
	}{
		{
			name: "identical",
			old:  "2026-08-05T17:13:36Z",
			new:  "2026-08-05T17:13:36Z",
			want: true,
		},
		{
			name: "same instant with offset",
			old:  "2026-08-05T17:13:36Z",
			new:  "2026-08-05T13:13:36-04:00",
			want: true,
		},
		{
			name: "different instant",
			old:  "2026-08-05T17:13:36Z",
			new:  "2026-08-06T17:13:36Z",
			want: false,
		},
		{
			name: "empty old",
			old:  "",
			new:  "2026-08-05T17:13:36Z",
			want: false,
		},
		{
			name: "empty new",
			old:  "2026-08-05T17:13:36Z",
			new:  "",
			want: false,
		},
		{
			name: "unparseable",
			old:  "not-a-timestamp",
			new:  "2026-08-05T17:13:36Z",
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := diffSuppressEquivalentRFC3339("expires_at", tc.old, tc.new, nil); got != tc.want {
				t.Errorf("want %t, got %t", tc.want, got)
			}
		})
	}
}

func TestFlattenAIRuntimeControlProviders(t *testing.T) {
	remoteState := []arcprovider.Provider{
		{
			ID:             "anthropic",
			DisplayName:    "Anthropic",
			DefaultBaseURL: "https://api.anthropic.com",
			Models: []arcprovider.Model{
				{ID: "claude-sonnet-4-20250514", DisplayName: "Claude Sonnet 4", ProviderID: "anthropic"},
			},
		},
		{
			ID:             "openai",
			DisplayName:    "Open AI",
			DefaultBaseURL: "https://api.openai.com",
		},
	}

	want := []map[string]any{
		{
			"id":               "anthropic",
			"display_name":     "Anthropic",
			"default_base_url": "https://api.anthropic.com",
			"models": []map[string]any{
				{
					"id":           "claude-sonnet-4-20250514",
					"display_name": "Claude Sonnet 4",
					"provider_id":  "anthropic",
				},
			},
		},
		{
			"id":               "openai",
			"display_name":     "Open AI",
			"default_base_url": "https://api.openai.com",
			"models":           []map[string]any{},
		},
	}

	if diff := cmp.Diff(want, flattenAIRuntimeControlProviders(remoteState)); diff != "" {
		t.Fatalf("unexpected result (-want +got):\n%s", diff)
	}
}

func TestFlattenAIRuntimeControlProviderConnections(t *testing.T) {
	remoteState := []providerconnection.ProviderConnection{
		{
			ID:        "3Yv86B123RGMSfghHajB",
			Name:      "OpenAI",
			Models:    []string{"gpt-4o", "gpt-4o-mini"},
			BaseURL:   "https://api.openai.com",
			CreatedAt: "2026-05-07T17:13:56Z",
			UpdatedAt: "2026-05-20T12:10:02Z",
		},
	}

	want := []map[string]any{
		{
			"id":         "3Yv86B123RGMSfghHajB",
			"name":       "OpenAI",
			"models":     []string{"gpt-4o", "gpt-4o-mini"},
			"base_url":   "https://api.openai.com",
			"created_at": "2026-05-07T17:13:56Z",
			"updated_at": "2026-05-20T12:10:02Z",
		},
	}

	if diff := cmp.Diff(want, flattenAIRuntimeControlProviderConnections(remoteState)); diff != "" {
		t.Fatalf("unexpected result (-want +got):\n%s", diff)
	}
}

func TestFlattenAIRuntimeControlVirtualKeys(t *testing.T) {
	remoteState := []key.VirtualKeyListItem{
		{
			ID:         "1Ui8CsZVOkUS1234567",
			Type:       "token",
			Name:       "production chatbot",
			Model:      "claude-sonnet-4-20250514",
			Provider:   "Anthropic",
			UserID:     "6zhNCY1236787aJIwUgN",
			UserName:   "Ellen Ripley",
			CustomerID: "2Hrt8ZCgxHgkZ123nNX6",
			CreatedAt:  "2026-05-07T16:42:01Z",
			CreatedBy:  "Ellen Ripley",
			UpdatedAt:  "2026-05-07T16:42:01Z",
			ExpiresAt:  gofastly.ToPointer("2026-05-08T04:42:01Z"),
			LastUsedAt: gofastly.ToPointer("2026-05-07T16:52:14Z"),
		},
		{
			// Null timestamps must flatten to empty strings rather than panic.
			ID:   "2Ui8CsZVOkUS7654321",
			Name: "staging chatbot",
		},
	}

	want := []map[string]any{
		{
			"id":           "1Ui8CsZVOkUS1234567",
			"type":         "token",
			"name":         "production chatbot",
			"model":        "claude-sonnet-4-20250514",
			"provider":     "Anthropic",
			"user_id":      "6zhNCY1236787aJIwUgN",
			"user_name":    "Ellen Ripley",
			"customer_id":  "2Hrt8ZCgxHgkZ123nNX6",
			"created_at":   "2026-05-07T16:42:01Z",
			"created_by":   "Ellen Ripley",
			"updated_at":   "2026-05-07T16:42:01Z",
			"expires_at":   "2026-05-08T04:42:01Z",
			"deleted_at":   "",
			"last_used_at": "2026-05-07T16:52:14Z",
		},
		{
			"id":           "2Ui8CsZVOkUS7654321",
			"type":         "",
			"name":         "staging chatbot",
			"model":        "",
			"provider":     "",
			"user_id":      "",
			"user_name":    "",
			"customer_id":  "",
			"created_at":   "",
			"created_by":   "",
			"updated_at":   "",
			"expires_at":   "",
			"deleted_at":   "",
			"last_used_at": "",
		},
	}

	if diff := cmp.Diff(want, flattenAIRuntimeControlVirtualKeys(remoteState)); diff != "" {
		t.Fatalf("unexpected result (-want +got):\n%s", diff)
	}
}

