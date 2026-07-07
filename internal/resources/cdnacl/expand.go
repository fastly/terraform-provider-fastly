package cdnacl

import (
	fastly "github.com/fastly/go-fastly/v16/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateACLInput {
	input := &fastly.CreateACLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
	}

	return input
}

func BuildDeleteInput(serviceID string, version int, name string) *fastly.DeleteACLInput {
	return &fastly.DeleteACLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
	}
}
