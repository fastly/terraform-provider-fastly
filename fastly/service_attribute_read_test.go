package fastly

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
)

// fakeReadAttribute mimics the shape of a real attribute handler Read: it
// reads from d, makes an API request, then writes the result back into d.
type fakeReadAttribute struct {
	DefaultServiceAttributeHandler
	err error
}

func (h *fakeReadAttribute) Register(_ *schema.Resource) error { return nil }

func (h *fakeReadAttribute) Process(_ context.Context, _ *schema.ResourceData, _ int, _ *gofastly.Client) error {
	return nil
}

func (h *fakeReadAttribute) Read(ctx context.Context, d *schema.ResourceData, _ *gofastly.ServiceDetail, conn *gofastly.Client) error {
	_ = d.Get(h.key)
	_ = d.Get("name")

	if _, err := conn.ListBackends(ctx, &gofastly.ListBackendsInput{ServiceID: "svc", ServiceVersion: 1}); err != nil {
		return err
	}
	if h.err != nil {
		return h.err
	}
	return d.Set(h.key, "read:"+h.key)
}

func newFakeReadFixture(t *testing.T, n int, latency time.Duration) (*gofastly.Client, *schema.ResourceData, []ServiceAttributeDefinition, *int32) {
	t.Helper()

	var inFlight, maxInFlight int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cur := atomic.AddInt32(&inFlight, 1)
		for {
			prev := atomic.LoadInt32(&maxInFlight)
			if cur <= prev || atomic.CompareAndSwapInt32(&maxInFlight, prev, cur) {
				break
			}
		}
		time.Sleep(latency)
		atomic.AddInt32(&inFlight, -1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(srv.Close)

	conn, err := gofastly.NewClientForEndpoint("test-key", srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	s := map[string]*schema.Schema{"name": {Type: schema.TypeString, Optional: true}}
	handlers := make([]ServiceAttributeDefinition, n)
	for i := range n {
		key := fmt.Sprintf("attr_%d", i)
		s[key] = &schema.Schema{Type: schema.TypeString, Optional: true}
		handlers[i] = &fakeReadAttribute{DefaultServiceAttributeHandler: DefaultServiceAttributeHandler{key: key}}
	}
	d := schema.TestResourceDataRaw(t, s, map[string]any{"name": "svc"})

	return conn, d, handlers, &maxInFlight
}

func TestReadServiceAttributes_RunsRequestsConcurrently(t *testing.T) {
	const (
		n       = 8
		latency = 200 * time.Millisecond
	)
	conn, d, handlers, maxInFlight := newFakeReadFixture(t, n, latency)

	start := time.Now()
	if err := readServiceAttributes(context.Background(), d, &gofastly.ServiceDetail{}, conn, handlers); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)

	for i := range n {
		key := fmt.Sprintf("attr_%d", i)
		if got := d.Get(key).(string); got != "read:"+key {
			t.Errorf("%s = %q, want %q", key, got, "read:"+key)
		}
	}
	if *maxInFlight < 2 {
		t.Errorf("expected overlapping requests, max in flight was %d", *maxInFlight)
	}
	if serial := n * latency; elapsed >= serial {
		t.Errorf("elapsed %s, expected faster than serial %s", elapsed, serial)
	}
}

func TestReadServiceAttributes_ReturnsFirstErrorInHandlerOrder(t *testing.T) {
	conn, d, handlers, _ := newFakeReadFixture(t, 4, 10*time.Millisecond)
	errA, errB := errors.New("a"), errors.New("b")
	handlers[1].(*fakeReadAttribute).err = errA
	handlers[3].(*fakeReadAttribute).err = errB

	err := readServiceAttributes(context.Background(), d, &gofastly.ServiceDetail{}, conn, handlers)
	if !errors.Is(err, errA) {
		t.Fatalf("got %v, want %v", err, errA)
	}
}

func TestReadServiceAttributes_Cancelled(t *testing.T) {
	conn, d, handlers, _ := newFakeReadFixture(t, 4, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := readServiceAttributes(ctx, d, &gofastly.ServiceDetail{}, conn, handlers)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}
