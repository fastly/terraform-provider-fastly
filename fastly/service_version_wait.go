package fastly

import (
	"context"
	"fmt"
	"log"
	"time"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
)

// serviceVersionWaitConfig controls how waitForServiceVersionAvailable polls.
type serviceVersionWaitConfig struct {
	// Timeout is the maximum total time to wait.
	Timeout time.Duration
	// InitialDelay is the delay before the second poll. It doubles after each
	// failed poll, up to MaxDelay.
	InitialDelay time.Duration
	// MaxDelay caps the delay between polls.
	MaxDelay time.Duration
}

// defaultServiceVersionWaitConfig returns the wait settings used by the
// provider.
func defaultServiceVersionWaitConfig() serviceVersionWaitConfig {
	return serviceVersionWaitConfig{
		Timeout:      60 * time.Second,
		InitialDelay: 250 * time.Millisecond,
		MaxDelay:     2 * time.Second,
	}
}

// waitForServiceVersionAvailable blocks until a freshly cloned service version
// is visible through the API and unlocked (mutable).
//
// This replaces an unconditional 7 second sleep. A version that is already
// available is detected on the first request, so in the common case this
// returns almost immediately instead of always waiting 7 seconds.
func waitForServiceVersionAvailable(ctx context.Context, conn *gofastly.Client, serviceID string, version int) error {
	return waitForServiceVersionAvailableWithConfig(ctx, conn, serviceID, version, defaultServiceVersionWaitConfig())
}

// waitForServiceVersionAvailableWithConfig is waitForServiceVersionAvailable
// with explicit wait settings. Tests use it to shorten the timing without
// changing package-level state.
func waitForServiceVersionAvailableWithConfig(ctx context.Context, conn *gofastly.Client, serviceID string, version int, cfg serviceVersionWaitConfig) error {
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	delay := cfg.InitialDelay
	var lastErr error
	for attempt := 1; ; attempt++ {
		v, err := conn.GetVersion(gofastly.NewContextForResourceID(ctx, serviceID), &gofastly.GetVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
		})
		switch {
		case err != nil:
			lastErr = err
		case v.Locked != nil && *v.Locked:
			lastErr = fmt.Errorf("version is still locked")
		default:
			log.Printf("[DEBUG] Fastly Service (%s) Version (%d) is available after %d attempt(s)", serviceID, version, attempt)
			return nil
		}

		log.Printf("[DEBUG] Waiting for Fastly Service (%s) Version (%d) to become available (attempt %d): %v", serviceID, version, attempt, lastErr)

		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for service (%s) version (%d) to become available: %w (last error: %v)", serviceID, version, ctx.Err(), lastErr)
		case <-time.After(delay):
		}

		delay *= 2
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}
}
