package service

import (
	"context"
	"fmt"

	"github.com/fastly/terraform-provider-fastly/internal/apierrors"

	fastly "github.com/fastly/go-fastly/v15/fastly"
)

// SelectReadVersion selects the service version that should be used for
// read-only discovery operations such as terraform query and compatibility
// import/read.
//
// This helper must never mutate remote state. It reads from the active version
// when one exists and otherwise falls back to the latest service version.
func SelectReadVersion(ctx context.Context, client *fastly.Client, serviceID string) (version int, active bool, err error) {
	service, err := client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: serviceID,
	})
	if err != nil {
		return 0, false, err
	}

	return SelectReadVersionFromDetails(service, serviceID)
}

// SelectReadVersionFromServiceSummary avoids an extra service details API
// call when the service list response already includes an active version. When
// no active version is present in the list response, it falls back to service
// details so it can read the latest version without mutating remote state.
func SelectReadVersionFromServiceSummary(ctx context.Context, client *fastly.Client, service *fastly.Service) (version int, active bool, err error) {
	if service == nil {
		return 0, false, fmt.Errorf("service is nil")
	}

	serviceID := fastly.ToValue(service.ServiceID)
	if serviceID == "" {
		return 0, false, fmt.Errorf("service ID is empty")
	}

	if service.ActiveVersion != nil && *service.ActiveVersion > 0 {
		return *service.ActiveVersion, true, nil
	}

	return SelectReadVersion(ctx, client, serviceID)
}

func SelectReadVersionFromDetails(service *fastly.ServiceDetail, serviceID string) (version int, active bool, err error) {
	if service == nil {
		return 0, false, fmt.Errorf("service details are nil for service %s", serviceID)
	}

	if activeVersion := VersionNumber(service.ActiveVersion); activeVersion > 0 {
		return activeVersion, true, nil
	}

	if latestVersion := VersionNumber(service.Version); latestVersion > 0 {
		return latestVersion, false, nil
	}

	if latestVersion := LatestVersionNumber(service.Versions); latestVersion > 0 {
		return latestVersion, false, nil
	}

	return 0, false, fmt.Errorf("no usable version found for service %s", serviceID)
}

// SelectWorkingVersionFromDetails selects the source version that a
// compatibility service resource should clone for CRUD updates.
//
// Initial service creation is handled separately and writes directly to
// Fastly-created version 1. For updates to an existing service, compatibility
// mode always clones the active version when one exists, otherwise it clones the
// latest existing version. The helper never mutates remote state itself.
func SelectWorkingVersionFromDetails(service *fastly.ServiceDetail, serviceID string) (version int, shouldClone bool, err error) {
	if service == nil {
		return 0, false, fmt.Errorf("service details are nil for service %s", serviceID)
	}

	if activeVersion := VersionNumber(service.ActiveVersion); activeVersion > 0 {
		return activeVersion, true, nil
	}

	if latestVersion := VersionNumber(service.Version); latestVersion > 0 {
		return latestVersion, true, nil
	}

	if latestVersion := LatestVersionNumber(service.Versions); latestVersion > 0 {
		return latestVersion, true, nil
	}

	return 0, false, fmt.Errorf("no usable version found for service %s", serviceID)
}

func VersionNumber(version *fastly.Version) int {
	if version == nil || version.Number == nil {
		return 0
	}
	return *version.Number
}

func LatestVersionNumber(versions []*fastly.Version) int {
	latest := 0
	for _, version := range versions {
		if number := VersionNumber(version); number > latest {
			latest = number
		}
	}
	return latest
}

func DeleteWithPolicy(ctx context.Context, client *fastly.Client, serviceID string, forceDestroy bool, reuse bool) error {
	if forceDestroy || reuse {
		service, err := client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
			ServiceID: serviceID,
		})
		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		if service.ActiveVersion != nil && service.ActiveVersion.Number != nil && *service.ActiveVersion.Number != 0 {
			_, err := client.DeactivateVersion(ctx, &fastly.DeactivateVersionInput{
				ServiceID:      serviceID,
				ServiceVersion: *service.ActiveVersion.Number,
			})
			if err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}
	}

	if reuse {
		return nil
	}

	err := client.DeleteService(ctx, &fastly.DeleteServiceInput{
		ServiceID: serviceID,
	})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func ValidateVersion(ctx context.Context, client *fastly.Client, serviceID string, version int) error {
	valid, msg, err := client.ValidateVersion(ctx, &fastly.ValidateVersionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid configuration for Fastly service %s version %d: %s", serviceID, version, msg)
	}
	return nil
}
