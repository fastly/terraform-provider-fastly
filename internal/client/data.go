package client

import (
	"fmt"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

type Data struct {
	Client         *fastly.Client
	VersionChecker *service.VersionChecker
}

func NewData(client *fastly.Client) *Data {
	return &Data{
		Client:         client,
		VersionChecker: service.NewVersionChecker(client),
	}
}

func FromProviderData(raw any) (*Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == nil {
		return nil, diags
	}

	data, ok := raw.(*Data)
	if !ok {
		diags.AddError(
			"Unexpected ProviderData type",
			fmt.Sprintf("Expected *client.Data, got: %T", raw),
		)
		return nil, diags
	}

	return data, diags
}
