package main

import (
	"github.com/fastly/terraform-provider-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: fastly.Provider})
}
