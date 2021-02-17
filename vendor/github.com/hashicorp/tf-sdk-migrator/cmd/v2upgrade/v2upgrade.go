package v2upgrade

import (
	"flag"
	"fmt"
	"go/printer"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/tf-sdk-migrator/util"
	"github.com/mitchellh/cli"
)

const (
	CommandName    = "v2upgrade"
	oldPackagePath = "github.com/hashicorp/terraform-plugin-sdk"
	newPackagePath = "github.com/hashicorp/terraform-plugin-sdk/v2"
	defaultVersion = "v2.4.3"
)

var printConfig = printer.Config{
	Mode:     printer.TabIndent | printer.UseSpaces,
	Tabwidth: 8,
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
	return `Usage: tf-sdk-migrator v2upgrade [--help] [--sdk-version SDK_VERSION] [IMPORT_PATH]

  Upgrades the Terraform provider to major version 2 of the Terraform
  provider SDK, defaulting to the git reference ` + defaultVersion + `.

  Rewrites import paths and go.mod. No backup is made before files are
  overwritten.

  IMPORT_PATH is resolved relative to $GOPATH/src/IMPORT_PATH. If it is not supplied,
  it is assumed that the current working directory contains a Terraform provider.

  Optionally, an SDK_VERSION can be passed, which is parsed as a Go module
  release version. For example: v2.0.1, latest, master.

Example:
  tf-sdk-migrator v2upgrade --sdk-version v2.0.0-rc.1 github.com/terraform-providers/terraform-provider-local`
}

func (c *command) Synopsis() string {
	return "Upgrades the Terraform provider SDK version to v2."
}

func (c *command) Run(args []string) int {
	flags := flag.NewFlagSet(CommandName, flag.ExitOnError)
	var sdkVersion string
	flags.StringVar(&sdkVersion, "sdk-version", defaultVersion, "SDK version")
	flags.Parse(args)

	var providerRepoName string
	var providerPath string
	if flags.NArg() == 1 {
		var err error
		providerRepoName = flags.Args()[0]
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

	c.ui.Output("Rewriting provider go.mod file...")
	err := util.RewriteGoMod(providerPath, sdkVersion, oldPackagePath, newPackagePath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error rewriting go.mod file: %s", err))
		return 1
	}

	c.ui.Output("Rewriting SDK package imports...")
	err = filepath.Walk(providerPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "vendor" {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			err := util.RewriteImportedPackageImports(path, oldPackagePath, newPackagePath)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error rewriting SDK imports: %s", err))
		return 1
	}

	c.ui.Output("Running `go mod tidy`...")
	err = util.GoModTidy(providerPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Error running go mod tidy: %s", err))
		return 1
	}

	var prettyProviderName string
	if providerRepoName != "" {
		prettyProviderName = " " + providerRepoName
	}
	c.ui.Info(fmt.Sprintf("Success! Provider%s is upgraded to %s %s.",
		prettyProviderName, newPackagePath, sdkVersion))

	hasVendor, err := HasVendorFolder(providerPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to check vendor folder: %s", err))
		return 1
	}

	if hasVendor {
		c.ui.Info("\nIt looks like this provider vendors dependencies. " +
			"Don't forget to run `go mod vendor`.")
	}

	c.ui.Info(fmt.Sprintf("Make sure to review all changes and run all tests."))
	return 0
}

func HasVendorFolder(providerPath string) (bool, error) {
	vendorPath := filepath.Join(providerPath, "vendor")
	fs, err := os.Stat(vendorPath)
	if err != nil {
		return false, err
	}
	if !fs.Mode().IsDir() {
		return false, fmt.Errorf("%s is not folder (expected folder)", vendorPath)
	}

	return true, nil
}

type ExecError struct {
	Err    error
	Stderr string
}

func (ee *ExecError) Error() string {
	return fmt.Sprintf("%s\n%s", ee.Err, ee.Stderr)
}

func NewExecError(err error, stderr string) *ExecError {
	return &ExecError{err, stderr}
}
