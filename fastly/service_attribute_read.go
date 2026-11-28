package fastly

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
)

// maxConcurrentAttributeReads caps how many attribute handlers can have an
// API request in flight at the same time during a service Read.
const maxConcurrentAttributeReads = 8

// readServiceAttributes calls Read on every attribute handler of a service and
// lets their API requests overlap.
//
// The handlers were written to run one after another, and
// *schema.ResourceData is not safe for concurrent use. So each handler runs
// while holding a shared mutex, which keeps all of its d.Get and d.Set calls
// strictly serialized. The mutex is released only while an HTTP request is in
// flight. To do that, each handler gets a copy of the Fastly client whose
// transport unlocks around RoundTrip. The handler code does not change, and
// only the network wait overlaps.
//
// Errors are reported in handler order, so the result is deterministic.
func readServiceAttributes(ctx context.Context, d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client, handlers []ServiceAttributeDefinition) error {
	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		sem  = make(chan struct{}, maxConcurrentAttributeReads)
		errs = make([]error, len(handlers))
	)

	for i, a := range handlers {
		wg.Add(1)
		go func(i int, a ServiceAttributeDefinition) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				errs[i] = ctx.Err()
				return
			}
			defer func() { <-sem }()

			mu.Lock()
			defer mu.Unlock()

			if err := ctx.Err(); err != nil {
				errs[i] = err
				return
			}
			errs[i] = a.Read(ctx, d, s, clientUnlockingDuringRequests(conn, &mu))
		}(i, a)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// clientUnlockingDuringRequests returns a shallow copy of conn whose HTTP
// transport releases mu while each request is in flight. The caller must hold
// mu whenever it uses the returned client.
func clientUnlockingDuringRequests(conn *gofastly.Client, mu *sync.Mutex) *gofastly.Client {
	c := *conn

	base := http.DefaultClient
	if conn.HTTPClient != nil {
		base = conn.HTTPClient
	}
	hc := *base
	rt := hc.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	hc.Transport = &unlockingTransport{mu: mu, underlying: rt}
	c.HTTPClient = &hc

	return &c
}

// unlockingTransport releases mu for the duration of a request, including
// reading the response body, and reacquires it before returning.
type unlockingTransport struct {
	mu         *sync.Mutex
	underlying http.RoundTripper
}

func (t *unlockingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Unlock()
	defer t.mu.Lock()

	resp, err := t.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Read the body while still unlocked so the download overlaps too.
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	return resp, nil
}
