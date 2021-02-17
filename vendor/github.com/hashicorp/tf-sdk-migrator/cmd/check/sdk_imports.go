package check

import (
	"strings"
)

var sdkPackages = map[string]bool{
	"github.com/hashicorp/terraform/helper/acctest":        true,
	"github.com/hashicorp/terraform/helper/customdiff":     true,
	"github.com/hashicorp/terraform/helper/encryption":     true,
	"github.com/hashicorp/terraform/helper/hashcode":       true,
	"github.com/hashicorp/terraform/helper/logging":        true,
	"github.com/hashicorp/terraform/helper/mutexkv":        true,
	"github.com/hashicorp/terraform/helper/pathorcontents": true,
	"github.com/hashicorp/terraform/helper/resource":       true,
	"github.com/hashicorp/terraform/helper/schema":         true,
	"github.com/hashicorp/terraform/helper/structure":      true,
	"github.com/hashicorp/terraform/helper/validation":     true,
	"github.com/hashicorp/terraform/httpclient":            true,
	"github.com/hashicorp/terraform/plugin":                true,
	"github.com/hashicorp/terraform/terraform":             true,
}

func CheckSDKPackageImports(details *ProviderImportDetails) ([]string, error) {
	removedPackagesInUse := []string{}

	for importPath := range details.AllImportPathsHash {
		if !strings.HasPrefix(importPath, "github.com/hashicorp/terraform/") {
			continue
		}

		if isSDK := sdkPackages[importPath]; !isSDK {
			removedPackagesInUse = append(removedPackagesInUse, importPath)
		}
	}

	return removedPackagesInUse, nil
}
