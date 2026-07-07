package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/fastly/terraform-provider-fastly/internal/errors"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type VersionCheckKey struct {
	ServiceID string
	Version   int
}

type VersionMutabilityResult struct {
	Locked bool
}

type VersionChecker struct {
	client *fastly.Client

	mu    sync.Mutex
	cache map[VersionCheckKey]VersionMutabilityResult
}

func NewVersionChecker(client *fastly.Client) *VersionChecker {
	return &VersionChecker{
		client: client,
		cache:  make(map[VersionCheckKey]VersionMutabilityResult),
	}
}

func (c *VersionChecker) GetMutability(
	ctx context.Context,
	serviceID string,
	version int,
) (VersionMutabilityResult, error) {
	key := VersionCheckKey{
		ServiceID: serviceID,
		Version:   version,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, ok := c.cache[key]; ok {
		return cached, nil
	}

	v, err := c.client.GetVersion(ctx, &fastly.GetVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return VersionMutabilityResult{}, err
	}

	result := VersionMutabilityResult{}
	if v != nil && v.Locked != nil {
		result.Locked = *v.Locked
	}

	c.cache[key] = result

	return result, nil
}

func (c *VersionChecker) EnsureMutable(
	ctx context.Context,
	serviceID string,
	version int,
) diag.Diagnostics {
	var diags diag.Diagnostics

	mutability, err := c.GetMutability(ctx, serviceID, version)
	if err != nil {
		diags.AddError(
			"Failed to inspect Fastly service version",
			fmt.Sprintf(
				"Could not read Fastly service version %d for service %q: %s",
				version,
				serviceID,
				err,
			),
		)
		return diags
	}

	if mutability.Locked {
		diags.AddError(
			"Fastly service version is not mutable",
			fmt.Sprintf(
				"Service %q version %d is locked and cannot be modified. Select a different editable version, or clone this version and pin Terraform to the new draft version before applying changes.",
				serviceID,
				version,
			),
		)
	}

	return diags
}

// EnsureMutableForDelete verifies a version can be deleted from. Unlike EnsureMutable,
// a not-found version is not an error: the service or version may already be gone
// (e.g. deleted by a prior step in the same apply), so there is nothing left to clean up.
// A locked version, however, always returns a diagnostic — the API cannot delete from
// a locked version, and silently succeeding would remove the resource from Terraform
// state while the remote object still exists, causing drift.
func (c *VersionChecker) EnsureMutableForDelete(
	ctx context.Context,
	serviceID string,
	version int,
) (notFound bool, diags diag.Diagnostics) {
	mutability, err := c.GetMutability(ctx, serviceID, version)
	if err != nil {
		if errors.IsNotFound(err) {
			return true, diags
		}

		diags.AddError(
			"Failed to inspect Fastly service version",
			fmt.Sprintf(
				"Could not read Fastly service version %d for service %q: %s",
				version,
				serviceID,
				err,
			),
		)
		return false, diags
	}

	if mutability.Locked {
		diags.AddError(
			"Fastly service version is not mutable",
			fmt.Sprintf(
				"Service %q version %d is locked and cannot be modified. Clone this version to remove the resource from an editable version, or destroy the entire service to remove all versions and their contents.",
				serviceID,
				version,
			),
		)
	}

	return false, diags
}
