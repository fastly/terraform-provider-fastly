package fastly

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// SERVICE ATTRIBUTE

type ServiceAttributeDefinition interface {
	Read(d *schema.ResourceData, s *fastly.ServiceDetail, conn *fastly.Client) error
	Process(d *schema.ResourceData, latestVersion int, conn *fastly.Client) error
	Register(d *schema.Resource) error
	HasChange(d *schema.ResourceData) bool
	MustProcess(d *schema.ResourceData, initialVersion bool) bool
}

type DefaultServiceAttributeHandler struct {
	schema *schema.Schema
	key    string
}

func (h *DefaultServiceAttributeHandler) GetKey() string {
	return h.key
}

func (h *DefaultServiceAttributeHandler) HasChange(d *schema.ResourceData) bool {
	return d.HasChange(h.key)
}

func (h *DefaultServiceAttributeHandler) MustProcess(d *schema.ResourceData, initialVersion bool) bool {
	return h.HasChange(d)
}

func (h *DefaultServiceAttributeHandler) MustProcess(d *schema.ResourceData, initialVersion bool) bool {
	return h.HasChange(d)
}


