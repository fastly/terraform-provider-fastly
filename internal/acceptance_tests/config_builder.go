package acceptancetests

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

// ServiceType represents a Fastly service resource type
type ServiceType string

const (
	ServiceCDNAuto     ServiceType = "service_cdn_auto"
	ServiceComputeAuto ServiceType = "service_compute_auto"
)

// BuildConfig constructs a Terraform config for a service with nested blocks.
// It loads a service template from the blocks directory, injects nested blocks into the RESOURCES placeholder,
// and replaces all placeholders with actual values.
// The service template file is "internal/acceptance_tests/blocks/{serviceType}.tf".
// Block paths should be relative to the repository root (e.g., "internal/acceptance_tests/blocks/domain_single.tf").
func BuildConfig(serviceType ServiceType, replacements map[string]string, blockPaths ...string) string {
	repoRoot := getRepoRoot()

	// Load service template
	templatePath := filepath.Join(repoRoot, "internal/acceptance_tests/blocks", string(serviceType)+".tf")
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		panic(fmt.Sprintf("failed to load service template %s: %v", templatePath, err))
	}

	// Load and concatenate nested blocks
	var resourcesBuilder strings.Builder
	for _, blockPath := range blockPaths {
		fullPath := filepath.Join(repoRoot, blockPath)
		blockBytes, err := os.ReadFile(fullPath)
		if err != nil {
			panic(fmt.Sprintf("failed to load block %s: %v", fullPath, err))
		}
		resourcesBuilder.WriteString(string(blockBytes))
	}

	// Start with the template and inject RESOURCES
	config := string(templateBytes)
	resources := strings.TrimSuffix(resourcesBuilder.String(), "\n")
	config = strings.ReplaceAll(config, "RESOURCES", resources)

	// Apply all other replacements
	for placeholder, value := range replacements {
		config = strings.ReplaceAll(config, placeholder, value)
	}

	return config
}
