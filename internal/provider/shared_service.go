package provider

import (
	"context"
	"fmt"

	fastly "github.com/fastly/go-fastly/v15/fastly"
)

const (
	serviceTypeVCL     = "vcl"
	serviceTypeCompute = "wasm"
)

func isFastlyNotFound(err error) bool {
	httpErr, ok := err.(*fastly.HTTPError)
	return ok && httpErr.StatusCode == 404
}

func serviceTypeLabel(serviceType string) string {
	switch serviceType {
	case serviceTypeVCL:
		return "CDN"
	case serviceTypeCompute:
		return "Compute"
	default:
		return serviceType
	}
}

func serviceTypeSupported(serviceType string, supportedTypes ...string) bool {
	for _, supported := range supportedTypes {
		if serviceType == supported {
			return true
		}
	}
	return false
}

func supportedServiceTypeLabels(supportedTypes []string) string {
	if len(supportedTypes) == 0 {
		return ""
	}

	out := ""
	for i, supported := range supportedTypes {
		if i > 0 {
			out += ", "
		}
		out += serviceTypeLabel(supported)
	}
	return out
}

// ensureServiceTypeSupported checks that a service ID belongs to one of the
// supported Fastly service types for a resource. This is primarily used by
// first-class resources. Some resources are valid for both CDN and
// Compute services, while future resources may only support one service type.
//
// This check runs during CRUD when service_id is known. It cannot reliably catch
// all invalid combinations during `terraform validate`, because service_id may
// be computed or come from a different state/workspace.
func ensureServiceTypeSupported(ctx context.Context, client *fastly.Client, serviceID string, resourceName string, supportedTypes ...string) error {
	service, err := client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: serviceID,
	})
	if err != nil {
		return err
	}

	serviceType := fastly.ToValue(service.Type)
	if serviceTypeSupported(serviceType, supportedTypes...) {
		return nil
	}

	return fmt.Errorf(
		"%s does not support Fastly service %q of type %q. Supported service types: %s",
		resourceName,
		serviceID,
		serviceTypeLabel(serviceType),
		supportedServiceTypeLabels(supportedTypes),
	)
}

// selectServiceReadVersion selects the service version that should be used for
// read-only discovery operations such as terraform query and compatibility
// import/read.
//
// This helper must never mutate remote state. It reads from the active version
// when one exists and otherwise falls back to the latest service version.
func selectServiceReadVersion(ctx context.Context, client *fastly.Client, serviceID string) (version int, active bool, err error) {
	service, err := client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
		ServiceID: serviceID,
	})
	if err != nil {
		return 0, false, err
	}

	return selectServiceReadVersionFromDetails(service, serviceID)
}

// selectServiceReadVersionFromServiceSummary avoids an extra service details API
// call when the service list response already includes an active version. When
// no active version is present in the list response, it falls back to service
// details so it can read the latest version without mutating remote state.
func selectServiceReadVersionFromServiceSummary(ctx context.Context, client *fastly.Client, service *fastly.Service) (version int, active bool, err error) {
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

	return selectServiceReadVersion(ctx, client, serviceID)
}

func selectServiceReadVersionFromDetails(service *fastly.ServiceDetail, serviceID string) (version int, active bool, err error) {
	if service == nil {
		return 0, false, fmt.Errorf("service details are nil for service %s", serviceID)
	}

	if activeVersion := serviceVersionNumber(service.ActiveVersion); activeVersion > 0 {
		return activeVersion, true, nil
	}

	if latestVersion := serviceVersionNumber(service.Version); latestVersion > 0 {
		return latestVersion, false, nil
	}

	if latestVersion := latestServiceVersionNumber(service.Versions); latestVersion > 0 {
		return latestVersion, false, nil
	}

	return 0, false, fmt.Errorf("no usable version found for service %s", serviceID)
}

// selectServiceWorkingVersionFromDetails selects the source version that a
// compatibility service resource should clone for CRUD updates.
//
// Initial service creation is handled separately and writes directly to
// Fastly-created version 1. For updates to an existing service, compatibility
// mode always clones the active version when one exists, otherwise it clones the
// latest existing version. The helper never mutates remote state itself.
func selectServiceWorkingVersionFromDetails(service *fastly.ServiceDetail, serviceID string) (version int, shouldClone bool, err error) {
	if service == nil {
		return 0, false, fmt.Errorf("service details are nil for service %s", serviceID)
	}

	if activeVersion := serviceVersionNumber(service.ActiveVersion); activeVersion > 0 {
		return activeVersion, true, nil
	}

	if latestVersion := serviceVersionNumber(service.Version); latestVersion > 0 {
		return latestVersion, true, nil
	}

	if latestVersion := latestServiceVersionNumber(service.Versions); latestVersion > 0 {
		return latestVersion, true, nil
	}

	return 0, false, fmt.Errorf("no usable version found for service %s", serviceID)
}

func serviceVersionNumber(version *fastly.Version) int {
	if version == nil || version.Number == nil {
		return 0
	}
	return *version.Number
}

func latestServiceVersionNumber(versions []*fastly.Version) int {
	latest := 0
	for _, version := range versions {
		if number := serviceVersionNumber(version); number > latest {
			latest = number
		}
	}
	return latest
}

func deleteServiceWithPolicy(ctx context.Context, client *fastly.Client, serviceID string, forceDestroy bool, reuse bool) error {
	if forceDestroy || reuse {
		service, err := client.GetServiceDetails(ctx, &fastly.GetServiceDetailsInput{
			ServiceID: serviceID,
		})
		if isFastlyNotFound(err) {
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
			if err != nil && !isFastlyNotFound(err) {
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
	if isFastlyNotFound(err) {
		return nil
	}
	return err
}
