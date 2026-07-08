package resourcelink

import (
	fastly "github.com/fastly/go-fastly/v16/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateResourceInput {
	return &fastly.CreateResourceInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		ResourceID:     new(service.StringValue(m.ResourceID)),
	}
}
