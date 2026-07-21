package fastly

import (
	"testing"
)

// The validation resource cannot be exercised via acceptance tests because
// completing domain validation requires a real, publicly resolvable
// domain. This at least pins the schema contract for certificate_id, which
// downstream fastly_tls_activation configurations depend on.
func TestResourceFastlyTLSSubscriptionValidation_CertificateIDSchema(t *testing.T) {
	s := resourceFastlyTLSSubscriptionValidation().Schema

	certificateID, ok := s["certificate_id"]
	if !ok {
		t.Fatal("expected schema to contain certificate_id")
	}
	if !certificateID.Computed {
		t.Error("expected certificate_id to be computed")
	}
	if certificateID.Required || certificateID.Optional {
		t.Error("expected certificate_id to be read-only")
	}
}
