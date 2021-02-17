package check

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func CheckDependencyVersion(providerPath, modPath, constaint string) (string, bool, error) {
	c, err := version.NewConstraint(constaint)

	modVersion, err := ReadVersionFromModFile(providerPath, modPath)
	if err != nil {
		return "", false, err
	}
	if modVersion == nil {
		return "", false, nil
	}

	return modVersion.String(), c.Check(modVersion), nil
}

func ReadVersionFromModFile(path, dependencyPath string) (*version.Version, error) {
	fullPath := filepath.Join(path, "go.mod")
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %s", fullPath, err)
	}

	pf, err := modfile.Parse(fullPath, content, nil)
	if err != nil {
		return nil, err
	}

	mv, err := findMatchingRequireStmt(pf.Require, dependencyPath)
	if err != nil {
		return nil, err
	}
	if mv == nil {
		return nil, nil
	}

	return version.NewVersion(mv.Version)
}

func findMatchingRequireStmt(requires []*modfile.Require, modPath string) (*module.Version, error) {
	for _, requireStmt := range requires {
		mod := requireStmt.Mod
		if mod.Path == modPath {
			return &mod, nil
		}
	}

	return nil, nil
}
