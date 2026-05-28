package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"golang.org/x/sync/singleflight"
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
	group singleflight.Group
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
	if cached, ok := c.cache[key]; ok {
		c.mu.Unlock()
		return cached, nil
	}
	c.mu.Unlock()

	cacheKey := fmt.Sprintf("%s:%d", serviceID, version)

	value, err, _ := c.group.Do(cacheKey, func() (interface{}, error) {
		c.mu.Lock()
		if cached, ok := c.cache[key]; ok {
			c.mu.Unlock()
			return cached, nil
		}
		c.mu.Unlock()

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

		c.mu.Lock()
		c.cache[key] = result
		c.mu.Unlock()

		return result, nil
	})
	if err != nil {
		return VersionMutabilityResult{}, err
	}

	result, ok := value.(VersionMutabilityResult)
	if !ok {
		return VersionMutabilityResult{}, fmt.Errorf("unexpected version mutability result type %T", value)
	}

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

func (c *VersionChecker) EnsureMutableForDelete(
	ctx context.Context,
	serviceID string,
	version int,
) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	mutability, err := c.GetMutability(ctx, serviceID, version)
	if err != nil {
		if isNotFound(err) {
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
				"Service %q version %d is locked and cannot be modified. Select a different editable version, or clone this version and pin Terraform to the new draft version before applying changes.",
				serviceID,
				version,
			),
		)
	}

	return false, diags
}
