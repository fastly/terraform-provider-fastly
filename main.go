package main

import (
	"context"

	"terraform-provider-fastly-dual-model-poc/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/fastly/fastly",
	})
}
