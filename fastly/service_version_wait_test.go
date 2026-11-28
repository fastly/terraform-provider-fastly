package fastly

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
)

func shortenServiceVersionWait(t *testing.T, timeout time.Duration) {
	t.Helper()
	origTimeout, origInitial, origMax := serviceVersionWaitTimeout, serviceVersionPollInitialDelay, serviceVersionPollMaxDelay
	serviceVersionWaitTimeout = timeout
	serviceVersionPollInitialDelay = 5 * time.Millisecond
	serviceVersionPollMaxDelay = 20 * time.Millisecond
	t.Cleanup(func() {
		serviceVersionWaitTimeout, serviceVersionPollInitialDelay, serviceVersionPollMaxDelay = origTimeout, origInitial, origMax
	})
}

func newVersionServer(t *testing.T, handler func(call int32, w http.ResponseWriter)) (*gofastly.Client, *int32) {
	t.Helper()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/service/svc/version/2" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		handler(atomic.AddInt32(&calls, 1), w)
	}))
	t.Cleanup(srv.Close)

	conn, err := gofastly.NewClientForEndpoint("test-key", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return conn, &calls
}

func TestWaitForServiceVersionAvailable_Immediate(t *testing.T) {
	shortenServiceVersionWait(t, time.Second)
	conn, calls := newVersionServer(t, func(_ int32, w http.ResponseWriter) {
		_, _ = w.Write([]byte(`{"number":2,"service_id":"svc","locked":false}`))
	})

	if err := waitForServiceVersionAvailable(context.Background(), conn, "svc", 2); err != nil {
		t.Fatal(err)
	}
	if *calls != 1 {
		t.Errorf("calls = %d, want 1", *calls)
	}
}

func TestWaitForServiceVersionAvailable_RetriesUntilAvailable(t *testing.T) {
	shortenServiceVersionWait(t, time.Second)
	conn, calls := newVersionServer(t, func(call int32, w http.ResponseWriter) {
		switch call {
		case 1:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"msg":"Record not found"}`))
		case 2:
			_, _ = w.Write([]byte(`{"number":2,"service_id":"svc","locked":true}`))
		default:
			_, _ = w.Write([]byte(`{"number":2,"service_id":"svc","locked":false}`))
		}
	})

	if err := waitForServiceVersionAvailable(context.Background(), conn, "svc", 2); err != nil {
		t.Fatal(err)
	}
	if *calls != 3 {
		t.Errorf("calls = %d, want 3", *calls)
	}
}

func TestWaitForServiceVersionAvailable_Timeout(t *testing.T) {
	shortenServiceVersionWait(t, 100*time.Millisecond)
	conn, _ := newVersionServer(t, func(_ int32, w http.ResponseWriter) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"msg":"Record not found"}`))
	})

	err := waitForServiceVersionAvailable(context.Background(), conn, "svc", 2)
	if err == nil || !strings.Contains(err.Error(), "version (2)") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}
