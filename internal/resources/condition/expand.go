package condition

import (
	fastly "github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateConditionInput {
	return &fastly.CreateConditionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           new(service.StringValue(m.Name)),
		Type:           new(service.StringValue(m.Type)),
		Statement:      new(service.StringValue(m.Statement)),
		Priority:       new(int(service.Int64Value(m.Priority))),
	}
}

// BuildUpdateInput deliberately omits Type: the Fastly API doesn't support changing a condition's
// type via update, so a type change is handled by ops.Update as a delete+create instead.
func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateConditionInput {
	return &fastly.UpdateConditionInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           service.StringValue(m.Name),
		Statement:      new(service.StringValue(m.Statement)),
		Priority:       new(int(service.Int64Value(m.Priority))),
	}
}
