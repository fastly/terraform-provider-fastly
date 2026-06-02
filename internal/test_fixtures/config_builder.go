package testfixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getRepoRoot finds the repository root by looking for go.mod
func getRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find repository root (go.mod not found)")
		}
		dir = parent
	}
}

// BuildConfig constructs a Terraform config by loading a template file, inserting block files,
// and replacing placeholders with actual values. The first path is the template, remaining paths are blocks.
// Paths should be relative to the repository root (e.g., "internal/test_fixtures/blocks/domain_single.tf").
func BuildConfig(replacements map[string]string, paths ...string) string {
	if len(paths) == 0 {
		panic("BuildConfig requires at least a template path")
	}

	repoRoot := getRepoRoot()

	// Load template (first path)
	templatePath := filepath.Join(repoRoot, paths[0])
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		panic(fmt.Sprintf("failed to load template %s: %v", templatePath, err))
	}
	template := string(templateBytes)

	// Load and insert blocks before closing brace (remaining paths)
	template = strings.TrimSpace(template)
	template = strings.TrimSuffix(template, "}")
	for _, blockPath := range paths[1:] {
		fullPath := filepath.Join(repoRoot, blockPath)
		blockBytes, err := os.ReadFile(fullPath)
		if err != nil {
			panic(fmt.Sprintf("failed to load block %s: %v", fullPath, err))
		}
		template = template + string(blockBytes) + "\n"
	}
	template = template + "}"

	// Apply all replacements
	for placeholder, value := range replacements {
		template = strings.ReplaceAll(template, placeholder, value)
	}

	return template
}
