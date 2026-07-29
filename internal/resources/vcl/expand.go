package vcl

import (
	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateVCLInput {
	name := service.StringValue(m.Name)
	content := service.StringValue(m.Content)
	main := service.BoolValue(m.Main)

	return &fastly.CreateVCLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           &name,
		Content:        &content,
		Main:           &main,
	}
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateVCLInput {
	content := service.StringValue(m.Content)

	return &fastly.UpdateVCLInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		Content:        &content,
	}
}
