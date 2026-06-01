package validation

import (
	"context"
	"fmt"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

// EnsureServiceTypeSupported checks that a service ID belongs to one of the
// supported Fastly service types for a resource. This is primarily used by
// first-class resources. Some resources are valid for both CDN and
// Compute services, while future resources may only support one service type.
//
// This check runs during CRUD when service_id is known. It cannot reliably catch
// all invalid combinations during `terraform validate`, because service_id may
// be computed or come from a different state/workspace.
func EnsureServiceTypeSupported(ctx context.Context, checker *service.ServiceTypeChecker, serviceID string, resourceName string, supportedTypes ...string) error {
	serviceType, err := checker.GetType(ctx, serviceID)
	if err != nil {
		return err
	}

	if service.TypeSupported(serviceType, supportedTypes...) {
		return nil
	}

	return fmt.Errorf(
		"%s does not support Fastly service %q of type %q. Supported service types: %s",
		resourceName,
		serviceID,
		service.TypeLabel(serviceType),
		service.SupportedTypeLabels(supportedTypes),
	)
}
