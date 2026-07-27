// Package acceptancetests provides infrastructure for writing acceptance tests.
//
// # Template System
//
// This package uses Go's text/template for safe placeholder replacement in Terraform configurations.
// All template files (.tf files in the blocks/ directory) use the {{.PLACEHOLDER_NAME}} format.
//
// Example template:
//
//	resource "fastly_service_cdn_auto" "test" {
//	  name = "{{.SERVICE_NAME}}"
//	  {{.RESOURCES}}
//	}
//
// The text/template approach prevents accidental replacements of legitimate Terraform code
// and provides proper error handling for template syntax issues.
package acceptancetests

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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
	ServiceCDN         ServiceType = "service_cdn"
	ServiceCDNAuto     ServiceType = "service_cdn_auto"
	ServiceCompute     ServiceType = "service_compute"
	ServiceComputeAuto ServiceType = "service_compute_auto"
)

// RenderBlock loads a single template file (relative to the repository
// root) and renders it with replacements, returning the raw result with no
// further formatting. Unlike BuildConfig, it doesn't wrap the result in a
// service resource - use it for standalone top-level resources that
// reference a service via a computed ID (e.g. `service_id = {{.SERVICE_ID_REF}}`)
// rather than nesting inside the service resource body.
//
// Placeholders in templates should use the format {{.PLACEHOLDER_NAME}}.
func RenderBlock(blockPath string, replacements map[string]string) string {
	fullPath := filepath.Join(getRepoRoot(), blockPath)
	blockBytes, err := os.ReadFile(fullPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load block %s: %v", fullPath, err))
	}

	tmpl, err := template.New(filepath.Base(blockPath)).Parse(string(blockBytes))
	if err != nil {
		panic(fmt.Sprintf("failed to parse block template %s: %v", blockPath, err))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, replacements); err != nil {
		panic(fmt.Sprintf("failed to execute block template %s: %v", blockPath, err))
	}

	return buf.String()
}

// BuildConfig constructs a Terraform config for a service with nested blocks.
// It loads a service template from the blocks directory, injects nested blocks into the {{.RESOURCES}} placeholder,
// and replaces all placeholders with actual values using text/template.
// The service template file is "internal/acceptance_tests/blocks/{serviceType}.tf".
// Block paths should be relative to the repository root (e.g., "internal/acceptance_tests/blocks/domain_single.tf").
//
// Placeholders in templates should use the format {{.PLACEHOLDER_NAME}}.
func BuildConfig(serviceType ServiceType, replacements map[string]string, blockPaths ...string) string {
	repoRoot := getRepoRoot()

	// Build template data with all replacements
	templateData := make(map[string]string)
	maps.Copy(templateData, replacements)

	// Load and parse nested blocks as templates
	var resourcesBuilder strings.Builder
	for _, blockPath := range blockPaths {
		block := normalizeNestedBlockIndent(RenderBlock(blockPath, templateData))
		if resourcesBuilder.Len() > 0 {
			resourcesBuilder.WriteString("\n")
		}
		resourcesBuilder.WriteString(block)
	}

	// Add the rendered resources to template data
	templateData["RESOURCES"] = resourcesBuilder.String()

	// Load service template
	templatePath := filepath.Join(repoRoot, "internal/acceptance_tests/blocks", string(serviceType)+".tf")
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		panic(fmt.Sprintf("failed to load service template %s: %v", templatePath, err))
	}

	// Parse and execute service template
	tmpl, err := template.New("config").Parse(string(templateBytes))
	if err != nil {
		panic(fmt.Sprintf("failed to parse service template %s: %v", templatePath, err))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		panic(fmt.Sprintf("failed to execute template for %s: %v", templatePath, err))
	}

	return buf.String()
}

func normalizeNestedBlockIndent(block string) string {
	block = strings.Trim(block, "\n")
	if block == "" {
		return ""
	}

	lines := strings.Split(block, "\n")
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if minIndent > 0 && len(line) >= minIndent {
			line = line[minIndent:]
		}
		lines[i] = "  " + line
	}

	return strings.Join(lines, "\n")
}
