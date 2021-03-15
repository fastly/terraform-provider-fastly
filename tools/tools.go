// +build tools

package tools

import (
	// document generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	// linter specifically for TF plugins
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlint"
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlintx"
)
