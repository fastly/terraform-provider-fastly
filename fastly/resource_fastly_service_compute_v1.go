package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Ordering is important - stored is processing order
// Some objects may need to be updated first, as they can be referenced by other
// configuration objects (Backends, Request Headers, etc).
var computeService = &BaseServiceDefinition{
	Type: "wasm",
	Attributes: []ServiceAttributeDefinition{
		NewServiceDomain(),
		NewServiceHealthCheck(),
		NewServiceBackend(),
		NewServicePackage(),
	},
}

func resourceServiceComputeV1() *schema.Resource {
	return resourceService(computeService)
}
