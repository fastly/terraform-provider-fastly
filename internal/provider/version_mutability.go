package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"golang.org/x/sync/singleflight"
)

type versionCheckKey struct {
	ServiceID string
	Version   int
}

type versionMutabilityResult struct {
	Active bool
	Locked bool
}

type providerData struct {
	client *fastly.Client

	versionCheckMu    sync.Mutex
	versionCheckCache map[versionCheckKey]versionMutabilityResult
	versionCheckGroup singleflight.Group
}

func newProviderData(client *fastly.Client) *providerData {
	return &providerData{
		client:            client,
		versionCheckCache: make(map[versionCheckKey]versionMutabilityResult),
	}
}

func (p *providerData) getVersionMutability(
	ctx context.Context,
	serviceID string,
	version int,
) (versionMutabilityResult, error) {
	key := versionCheckKey{
		ServiceID: serviceID,
		Version:   version,
	}

	p.versionCheckMu.Lock()
	if cached, ok := p.versionCheckCache[key]; ok {
		p.versionCheckMu.Unlock()
		return cached, nil
	}
	p.versionCheckMu.Unlock()

	cacheKey := fmt.Sprintf("%s:%d", serviceID, version)

	value, err, _ := p.versionCheckGroup.Do(cacheKey, func() (interface{}, error) {
		p.versionCheckMu.Lock()
		if cached, ok := p.versionCheckCache[key]; ok {
			p.versionCheckMu.Unlock()
			return cached, nil
		}
		p.versionCheckMu.Unlock()

		v, err := p.client.GetVersion(ctx, &fastly.GetVersionInput{
			ServiceID:      serviceID,
			ServiceVersion: version,
		})
		if err != nil {
			return versionMutabilityResult{}, err
		}

		result := versionMutabilityResult{}
		if v != nil && v.Active != nil {
			result.Active = *v.Active
		}
		if v != nil && v.Locked != nil {
			result.Locked = *v.Locked
		}

		p.versionCheckMu.Lock()
		p.versionCheckCache[key] = result
		p.versionCheckMu.Unlock()

		return result, nil
	})
	if err != nil {
		return versionMutabilityResult{}, err
	}

	result, ok := value.(versionMutabilityResult)
	if !ok {
		return versionMutabilityResult{}, fmt.Errorf("unexpected version mutability result type %T", value)
	}

	return result, nil
}

func (p *providerData) ensureVersionMutable(
	ctx context.Context,
	serviceID string,
	version int,
) diag.Diagnostics {
	var diags diag.Diagnostics

	mutability, err := p.getVersionMutability(ctx, serviceID, version)
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

	if mutability.Active {
		diags.AddError(
			"Fastly service version is not mutable",
			fmt.Sprintf(
				"Service %q version %d is currently active and cannot be modified. Clone the active version first and update your Terraform input to pin the new draft version before applying changes.",
				serviceID,
				version,
			),
		)
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

func (p *providerData) ensureVersionMutableForDelete(
	ctx context.Context,
	serviceID string,
	version int,
) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	mutability, err := p.getVersionMutability(ctx, serviceID, version)
	if err != nil {
		if isFastlyNotFound(err) {
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

	if mutability.Active {
		diags.AddError(
			"Fastly service version is not mutable",
			fmt.Sprintf(
				"Service %q version %d is currently active and cannot be modified. Clone the active version first and update your Terraform input to pin the new draft version before applying changes.",
				serviceID,
				version,
			),
		)
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
