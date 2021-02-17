package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

func ParseFile(filePath string) (*ast.File, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return parseFile(filePath, string(b))
}

func parseFile(filePath, content string) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, filePath, content, 0)
}

func FindPackageReferences(f *ast.File, importPath string) ([]ast.Node, error) {
	importName, err := FindImportName(f.Imports, importPath)
	if err != nil {
		return nil, err
	}

	var refs = make([]ast.Node, 0)

	ast.Inspect(f, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		switch node.(type) {
		case *ast.File:
			return true
		case *ast.ImportSpec:
			// Imports are processed separately (above)
			return false
		case *ast.SelectorExpr:
			selector := node.(*ast.SelectorExpr)
			x := selector.X

			ident, ok := x.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name != importName {
				return true
			}

			if ident.Obj != nil {
				// Avoid reporting references to local variables
				// which may have the same name as package
				return true
			}

			refs = append(refs, selector.Sel)
		}

		return true
	})

	return refs, nil
}

// This implementation does not handle case where package
// is imported under different name without explicit alias
// e.g. github.com/hashicorp/go-discover (imported as "discover")
// This would require downloading the module and adding more complexity.
func FindImportName(imports []*ast.ImportSpec, importPath string) (string, error) {
	if len(importPath) == 0 {
		return "", fmt.Errorf("Unknown import path")
	}

	parts := strings.Split(importPath, "/")
	importName := parts[len(parts)-1]
	for _, imp := range imports {
		path := strings.Trim(imp.Path.Value, "\"")
		if path == importPath {
			if imp.Name != nil {
				importName = imp.Name.Name
			}

			return importName, nil
		}
	}

	return "", fmt.Errorf("Import %q not found", importPath)
}
