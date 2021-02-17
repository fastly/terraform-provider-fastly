package check

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/tf-sdk-migrator/util"
	"github.com/mitchellh/cli"
)

const (
	CommandName = "check"

	goVersionConstraint = ">=1.12"

	tfModPath           = "github.com/hashicorp/terraform"
	tfVersionConstraint = ">=0.12.7"

	sdkModPath           = "github.com/hashicorp/terraform-plugin-sdk"
	sdkVersionConstraint = ">=1.0.0"
)

type AlreadyMigrated struct {
	sdkVersion string
}

func (am *AlreadyMigrated) Error() string {
	return fmt.Sprintf("Provider already migrated to SDK version %s", am.sdkVersion)
}

type command struct {
	ui cli.Ui
}

func CommandFactory(ui cli.Ui) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return &command{ui}, nil
	}
}

func (c *command) Help() string {
	return `Usage: tf-sdk-migrator check [--help] [--csv] [IMPORT_PATH]

  Checks whether the Terraform provider at PATH is ready to be migrated to the
  new Terraform provider SDK (v1).

  IMPORT_PATH is resolved relative to $GOPATH/src/IMPORT_PATH. If it is not supplied,
  it is assumed that the current working directory contains a Terraform provider.

  By default, outputs a human-readable report and exits 0 if the provider is
  ready for migration, 1 otherwise.

Options:
  --csv    Output results in CSV format.

Example:
  tf-sdk-migrator check github.com/terraform-providers/terraform-provider-local
`
}

func (c *command) Synopsis() string {
	return "Checks whether a Terraform provider is ready to be migrated to the new SDK (v1)."
}

func (c *command) Run(args []string) int {
	flags := flag.NewFlagSet(CommandName, flag.ExitOnError)
	var csv bool
	flags.BoolVar(&csv, "csv", false, "CSV output")
	flags.Parse(args)

	var providerRepoName string
	var providerPath string
	if flags.NArg() == 1 {
		var err error
		providerRepoName := flags.Args()[0]
		providerPath, err = util.GetProviderPath(providerRepoName)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error finding provider %s: %s", providerRepoName, err))
			return 1
		}
	} else if flags.NArg() == 0 {
		var err error
		providerPath, err = os.Getwd()
		if err != nil {
			c.ui.Error(fmt.Sprintf("Error finding current working directory: %s", err))
			return 1
		}
	} else {
		return cli.RunResultHelp
	}

	err := runCheck(c.ui, providerPath, providerRepoName, csv)
	if err != nil {
		msg, alreadyMigrated := err.(*AlreadyMigrated)
		if alreadyMigrated {
			c.ui.Info(msg.Error())
			return 0
		}

		if !csv {
			c.ui.Error(err.Error())
		}
		return 1
	}

	return 0
}

func RunCheck(ui cli.Ui, providerPath, repoName string) error {
	return runCheck(ui, providerPath, repoName, false)
}

func runCheck(ui cli.Ui, providerPath, repoName string, csv bool) error {
	if !csv {
		ui.Output("Checking Go runtime version ...")
	}
	goVersion, goVersionSatisfied := CheckGoVersion(providerPath)
	if !csv {
		if goVersionSatisfied {
			ui.Info(fmt.Sprintf("Go version %s: OK.", goVersion))
		} else {
			ui.Warn(fmt.Sprintf("Go version does not satisfy constraint %s. Found Go version: %s.", goVersionConstraint, goVersion))
		}
	}

	if !csv {
		ui.Output("Checking whether provider uses Go modules...")
	}
	goModulesUsed := CheckForGoModules(providerPath)
	if !csv {
		if goModulesUsed {
			ui.Info("Go modules in use: OK.")
		} else {
			ui.Warn("Go modules not in use. Provider must use Go modules.")
		}
	}

	if !csv {
		ui.Output(fmt.Sprintf("Checking version of %s to determine if provider was already migrated...", sdkModPath))
	}
	sdkVersion, sdkVersionSatisfied, err := CheckDependencyVersion(providerPath, sdkModPath, sdkVersionConstraint)
	if err != nil {
		return fmt.Errorf("Error getting SDK version for provider %s: %s", providerPath, err)
	}
	if !csv {
		if sdkVersionSatisfied {
			return &AlreadyMigrated{sdkVersion}
		} else if sdkVersion != "" {
			return fmt.Errorf("Provider already migrated, but SDK version %s does not satisfy constraint %s.",
				sdkVersion, sdkVersionConstraint)
		}
	}

	if !csv {
		ui.Output(fmt.Sprintf("Checking version of %s used in provider...", tfModPath))
	}
	tfVersion, tfVersionSatisfied, err := CheckDependencyVersion(providerPath, tfModPath, tfVersionConstraint)
	if err != nil {
		return fmt.Errorf("Error getting Terraform version for provider %s: %s", providerPath, err)
	}
	if !csv {
		if tfVersionSatisfied {
			ui.Info(fmt.Sprintf("Terraform version %s: OK.", tfVersion))
		} else if tfVersion != "" {
			ui.Warn(fmt.Sprintf("Terraform version does not satisfy constraint %s. Found Terraform version: %s", tfVersionConstraint, tfVersion))
		} else {
			return fmt.Errorf("This directory (%s) doesn't seem to be a Terraform provider.\nProviders depend on %s", providerPath, tfModPath)
		}
	}

	if !csv {
		ui.Output("Checking whether provider uses deprecated SDK packages or identifiers...")
	}
	removedPackagesInUse, removedIdentsInUse, err := CheckSDKPackageImportsAndRefs(providerPath)
	if err != nil {
		return err
	}
	usesRemovedPackagesOrIdents := len(removedPackagesInUse) > 0 || len(removedIdentsInUse) > 0
	if !csv {
		if err != nil {
			return fmt.Errorf("Error determining use of deprecated SDK packages and identifiers: %s", err)
		}
		if !usesRemovedPackagesOrIdents {
			ui.Info("No imports of deprecated SDK packages or identifiers: OK.")
		}
		formatRemovedPackages(ui, removedPackagesInUse)
		formatRemovedIdents(ui, removedIdentsInUse)
	}
	constraintsSatisfied := goVersionSatisfied && goModulesUsed && tfVersionSatisfied && !usesRemovedPackagesOrIdents
	if csv {
		ui.Output(fmt.Sprintf("go_version,go_version_satisfies_constraint,uses_go_modules,sdk_version,sdk_version_satisfies_constraint,does_not_use_removed_packages,all_constraints_satisfied\n%s,%t,%t,%s,%t,%t,%t",
			goVersion, goVersionSatisfied, goModulesUsed, tfVersion, tfVersionSatisfied, !usesRemovedPackagesOrIdents, constraintsSatisfied))
	} else {
		var prettyProviderName string
		if repoName != "" {
			prettyProviderName = " " + repoName
		}
		if constraintsSatisfied {
			ui.Info(fmt.Sprintf("\nAll constraints satisfied. Provider%s can be migrated to the new SDK.\n", prettyProviderName))
			return nil
		} else if goModulesUsed && tfVersionSatisfied && !usesRemovedPackagesOrIdents {
			ui.Info(fmt.Sprintf("\nProvider%s can be migrated to the new SDK, but Go version %s is recommended.\n", prettyProviderName, goVersionConstraint))
			return nil
		}
	}

	return fmt.Errorf("\nSome constraints not satisfied. Please resolve these before migrating to the new SDK.")
}

func formatRemovedPackages(ui cli.Ui, removedPackagesInUse []string) {
	if len(removedPackagesInUse) == 0 {
		return
	}

	ui.Warn("Deprecated SDK packages in use:")
	for _, pkg := range removedPackagesInUse {
		ui.Warn(fmt.Sprintf(" * %s", pkg))
	}
}

func formatRemovedIdents(ui cli.Ui, removedIdentsInUse []*Offence) {
	if len(removedIdentsInUse) == 0 {
		return
	}
	ui.Warn("Deprecated SDK identifiers in use:")
	for _, ident := range removedIdentsInUse {
		d := ident.IdentDeprecation
		ui.Warn(fmt.Sprintf(" * %s (%s)", d.Identifier.Name, d.ImportPath))

		for _, pos := range ident.Positions {
			ui.Warn(fmt.Sprintf("   * %s", pos))
		}
	}
}

func CheckGoVersion(providerPath string) (goVersion string, satisfiesConstraint bool) {
	c, err := version.NewConstraint(goVersionConstraint)

	runtimeVersion := strings.TrimLeft(runtime.Version(), "go")
	v, err := version.NewVersion(runtimeVersion)
	if err != nil {
		log.Printf("[WARN] Could not parse Go version %s", runtimeVersion)
		return "", false
	}

	return runtimeVersion, c.Check(v)
}

func CheckForGoModules(providerPath string) (usingModules bool) {
	if _, err := os.Stat(filepath.Join(providerPath, "go.mod")); err != nil {
		log.Printf("[WARN] 'go.mod' file not found - provider %s is not using Go modules", providerPath)
		return false
	}
	return true
}

func CheckSDKPackageImportsAndRefs(providerPath string) (removedPackagesInUse []string, packageRefsOffences []*Offence, err error) {
	var providerImportDetails *ProviderImportDetails

	providerImportDetails, err = GoListPackageImports(providerPath)
	if err != nil {
		return nil, nil, err
	}

	removedPackagesInUse, err = CheckSDKPackageImports(providerImportDetails)
	if err != nil {
		return nil, nil, err
	}

	packageRefsOffences, err = CheckSDKPackageRefs(providerImportDetails)
	if err != nil {
		return nil, nil, err
	}

	return
}
