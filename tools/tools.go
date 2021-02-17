// +build tools

package tools

import (
	// document generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	// SDK upgrade
	_ "github.com/hashicorp/tf-sdk-migrator"
)
