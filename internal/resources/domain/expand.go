package domain

import (
	fastly "github.com/fastly/go-fastly/v15/fastly"
)

func expandCreate(m Model) *fastly.CreateDomainInput {
	opts := &fastly.CreateDomainInput{
		ServiceID:      m.Service.ValueString(),
		ServiceVersion: int(m.Version.ValueInt64()),
		Name:           new(m.Name.ValueString()),
	}
	if !m.Comment.IsNull() && m.Comment.ValueString() != "" {
		opts.Comment = new(m.Comment.ValueString())
	}
	return opts
}

func expandUpdate(m Model) *fastly.UpdateDomainInput {
	opts := &fastly.UpdateDomainInput{
		ServiceID:      m.Service.ValueString(),
		ServiceVersion: int(m.Version.ValueInt64()),
		Name:           m.Name.ValueString(),
	}
	if !m.Comment.IsNull() && m.Comment.ValueString() != "" {
		opts.Comment = new(m.Comment.ValueString())
	}
	return opts
}
