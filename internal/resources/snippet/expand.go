package snippet

import (
	"strconv"

	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v17/fastly"
)

func BuildCreateInput(serviceID string, version int, m NestedModel) *fastly.CreateSnippetInput {
	name := service.StringValue(m.Name)
	content := service.StringValue(m.Content)
	priority := strconv.FormatInt(service.Int64Value(m.Priority), 10)
	snippetType := fastly.SnippetType(normalizeType(service.StringValue(m.Type)))
	dynamic := 0

	return &fastly.CreateSnippetInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           &name,
		Content:        &content,
		Priority:       &priority,
		Type:           &snippetType,
		Dynamic:        &dynamic,
	}
}

func BuildUpdateInput(serviceID string, version int, m NestedModel) *fastly.UpdateSnippetInput {
	name := service.StringValue(m.Name)
	content := service.StringValue(m.Content)
	priority := strconv.FormatInt(service.Int64Value(m.Priority), 10)
	snippetType := fastly.SnippetType(normalizeType(service.StringValue(m.Type)))

	return &fastly.UpdateSnippetInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		Name:           name,
		NewName:        &name,
		Content:        &content,
		Priority:       &priority,
		Type:           &snippetType,
	}
}
