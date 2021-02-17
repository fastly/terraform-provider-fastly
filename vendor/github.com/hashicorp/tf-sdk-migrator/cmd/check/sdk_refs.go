package check

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"

	"github.com/hashicorp/tf-sdk-migrator/util"
	goList "github.com/kmoe/go-list"
	refsParser "github.com/radeksimko/go-refs/parser"
)

type Offence struct {
	IdentDeprecation *identDeprecation
	Positions        []*token.Position
}

type identDeprecation struct {
	ImportPath string
	Identifier *ast.Ident
	Message    string
}

var deprecations = []*identDeprecation{
	{
		"github.com/hashicorp/terraform/httpclient",
		ast.NewIdent("UserAgentString"),
		"This function has been removed, please use httpclient.TerraformUserAgent(version) instead",
	},
	{
		"github.com/hashicorp/terraform/httpclient",
		ast.NewIdent("New"),
		"This function has been removed, please use DefaultPooledClient() with custom Transport/round-tripper from github.com/hashicorp/go-cleanhttp instead",
	},
	{
		"github.com/hashicorp/terraform/terraform",
		ast.NewIdent("UserAgentString"),
		"This function has been removed, please use httpclient.TerraformUserAgent(version) instead",
	},
	{
		"github.com/hashicorp/terraform/terraform",
		ast.NewIdent("VersionString"),
		"This function has been removed, please use helper/schema's Provider.TerraformVersion available from Provider.ConfigureFunc",
	},
	{
		"github.com/hashicorp/terraform/config",
		ast.NewIdent("UserAgentString"),
		"Please don't use this",
	},
	{
		"github.com/hashicorp/terraform/config",
		ast.NewIdent("NewRawConfig"),
		"terraform.NewResourceConfig and config.NewRawConfig have been removed, please use terraform.NewResourceConfigRaw",
	},
	{
		"github.com/hashicorp/terraform/terraform",
		ast.NewIdent("NewResourceConfig"),
		"terraform.NewResourceConfig and config.NewRawConfig have been removed, please use terraform.NewResourceConfigRaw",
	},
}

// ProviderImports is a data structure we parse the `go list` output into
// for efficient searching
type ProviderImportDetails struct {
	AllImportPathsHash map[string]bool
	Packages           map[string]ProviderPackage
}

type ProviderPackage struct {
	Dir         string
	ImportPath  string
	GoFiles     []string
	TestGoFiles []string
	Imports     []string
	TestImports []string
}

func GoListPackageImports(providerPath string) (*ProviderImportDetails, error) {
	packages, err := goList.GoList(providerPath, "./...", "-mod=vendor")
	if err != nil {
		return nil, err
	}

	allImportPathsHash := make(map[string]bool)
	providerPackages := make(map[string]ProviderPackage)

	for _, p := range packages {
		for _, i := range p.Imports {
			allImportPathsHash[i] = true
		}

		providerPackages[p.ImportPath] = ProviderPackage{
			Dir:         p.Dir,
			ImportPath:  p.ImportPath,
			GoFiles:     p.GoFiles,
			TestGoFiles: p.TestGoFiles,
			Imports:     p.Imports,
			TestImports: p.TestImports,
		}
	}

	return &ProviderImportDetails{
		AllImportPathsHash: allImportPathsHash,
		Packages:           providerPackages,
	}, nil
}

func CheckSDKPackageRefs(providerImportDetails *ProviderImportDetails) ([]*Offence, error) {
	offences := make([]*Offence, 0, 0)

	for _, d := range deprecations {
		fset := token.NewFileSet()
		files, err := filesWhichImport(providerImportDetails, d.ImportPath)
		if err != nil {
			return nil, err
		}

		foundPositions := make([]*token.Position, 0, 0)

		for _, filePath := range files {
			f, err := parser.ParseFile(fset, filePath, nil, 0)
			if err != nil {
				return nil, err
			}

			identifiers, err := refsParser.FindPackageReferences(f, d.ImportPath)
			if err != nil {
				// package not imported in this file
				continue
			}

			positions, err := findIdentifierPositions(fset, identifiers, d.Identifier)
			if err != nil {
				return nil, err
			}

			if len(positions) > 0 {
				foundPositions = append(foundPositions, positions...)
			}
		}

		if len(foundPositions) > 0 {
			offences = append(offences, &Offence{
				IdentDeprecation: d,
				Positions:        foundPositions,
			})
		}
	}

	return offences, nil
}

func findIdentifierPositions(fset *token.FileSet, nodes []ast.Node, ident *ast.Ident) ([]*token.Position, error) {
	positions := make([]*token.Position, 0, 0)

	for _, node := range nodes {
		nodeName := fmt.Sprint(node)
		if nodeName == ident.String() {
			position := fset.Position(node.Pos())
			positions = append(positions, &position)
		}
	}

	return positions, nil
}

func filesWhichImport(providerImportDetails *ProviderImportDetails, importPath string) (files []string, e error) {
	files = []string{}
	for _, p := range providerImportDetails.Packages {
		if util.StringSliceContains(p.Imports, importPath) {
			files = append(files, prependDirToFilePaths(p.GoFiles, p.Dir)...)
		}
		if util.StringSliceContains(p.TestImports, importPath) {
			files = append(files, prependDirToFilePaths(p.TestGoFiles, p.Dir)...)
		}
	}

	return files, nil
}

func prependDirToFilePaths(filePaths []string, dir string) []string {
	newFilePaths := []string{}
	for _, f := range filePaths {
		newFilePaths = append(newFilePaths, path.Join(dir, f))
	}
	return newFilePaths
}
