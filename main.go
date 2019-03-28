package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/Wikia/terraform-provider-fastly/fastly"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: fastly.Provider})
}
