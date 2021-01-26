// This script is designed to precompile all our documentation template files
// so they are ready to then be passed onto the official terraform
// documentation generation tool:
// https://github.com/hashicorp/terraform-plugin-docs
//
// The reason why we need to precompile the templates is because they use the
// {{ define "..." }} syntax to allow us to reuse Markdown content across
// multiple files, and so we need this script to expose those separate files as
// if they were defined within a single template file.
//
// The reason we need the official terraform documentation generator tool is
// because it helps to pick up missing content that we've not defined by
// reflecting over the schemas of our resources and data sources.
//
// THE COMPLETE DOCUMENTATION STEPS ARE:
//
// 1. acquire all the templates
// 2. render the templates into Markdown
// 3. write the file output to a temp directory and still use .tmpl extension
// 4. append to each rendered .tmpl file the template code needed by tfplugindocs (e.g. {{ .SchemaMarkdown | trimspace }})
// 5. rename repo /templates/{data-sources/resources} directories to avoid being overwritten by next step
// 6. move contents of temp directory (i.e. data-sources/resources) into repo /templates directory
// 7. run tfplugindocs generate function to output final documentation to /docs directory
// 8. replace /templates/{data-sources/resources} directories with their backed up equivalents
//
package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PageData represents a type of service and enables a template to render
// different content depending on the service type.
type PageData struct {
	ServiceType string
}

// Page represents a template page to be rendered.
//
// name: the {{ define "..." }} in each template.
// path: the path to write rendered output to.
// Data: context specific information (e.g. vcl vs wasm).
//
// Data is public as it's called via the template processing logic.
type Page struct {
	name string
	path string
	Data PageData
}

func main() {
	baseDir := getBaseDir()
	tmplDir := baseDir + "/templates"
	// docsDir := baseDir + "/docs/"

	tempDir, err := ioutil.TempDir("", "precompile")
	fmt.Println(tempDir)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	var dataPages = []Page{
		{
			name: "ip_ranges",
			path: tempDir + "/data-sources/ip_ranges.md.tmpl",
		},
		{
			name: "waf_rules",
			path: tempDir + "/data-sources/waf_rules.md.tmpl",
		},
	}

	var resourcePages = []Page{
		{
			name: "service_v1",
			path: tempDir + "/resources/service_v1.md.tmpl",
			Data: PageData{
				"vcl",
			},
		},
		{
			name: "service_compute",
			path: tempDir + "/resources/service_compute.md.tmpl",
			Data: PageData{
				"wasm",
			},
		},
		{
			name: "service_dictionary_items_v1",
			path: tempDir + "/resources/service_dictionary_items_v1.md.tmpl",
		},
		{
			name: "service_acl_entries_v1",
			path: tempDir + "/resources/service_acl_entries_v1.md.tmpl",
		},
		{
			name: "service_dynamic_snippet_content_v1",
			path: tempDir + "/resources/service_dynamic_snippet_content_v1.md.tmpl",
		},
		{
			name: "service_waf_configuration",
			path: tempDir + "/resources/service_waf_configuration.md.tmpl",
		},
		{
			name: "user_v1",
			path: tempDir + "/resources/user_v1.md.tmpl",
		},
	}

	pages := append(resourcePages, dataPages...)

	renderPages(getTemplate(tmplDir), pages)

	appendSyntaxToFiles(tempDir)

	backupTemplatesDir(tmplDir)

	replaceTemplatesDir(tmplDir, tempDir)
}

// getBaseDir returns the terraform repo directory.
//
// TODO: simplify this logic with something like:
// git rev-parse --show-toplevel
func getBaseDir() string {
	_, scriptPath, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not get current working directory")
	}
	for !strings.HasPrefix(filepath.Base(scriptPath), "terraform-provider-") && scriptPath != "/" {
		scriptPath = filepath.Clean(scriptPath + "/..")
	}
	if scriptPath == "/" {
		log.Fatal("Script was run outside of fastly provider directory")
	}
	return scriptPath
}

// getTemplate walks the templates directory filtering non-tmpl extension
// files, and parsing all the templates found (ensuring they must parse).
func getTemplate(tmplDir string) *template.Template {
	var templateFiles []string
	filepath.Walk(tmplDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".tmpl" {
			templateFiles = append(templateFiles, path)
		}
		return nil
	})
	return template.Must(template.ParseFiles(templateFiles...))
}

// renderPages iterates over the given pages and renders each element.
func renderPages(t *template.Template, pages []Page) {
	for _, p := range pages {
		renderPage(t, p)
	}
}

// renderPage creates a new file based on the page information given, and
// renders the associated template for that page.
func renderPage(t *template.Template, p Page) {
	basePath := filepath.Dir(p.path)
	err := makeDirectoryIfNotExists(basePath)
	if err != nil {
		log.Fatal()
	}

	f, err := os.Create(p.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = t.ExecuteTemplate(f, p.name, p)
	if err != nil {
		panic(err)
	}
}

// makeDirectoryIfNotExists asserts whether a directory exists and makes it
// if not. Returns nil if exists or successfully made.
func makeDirectoryIfNotExists(path string) error {
	fi, err := os.Stat(path)
	switch {
	case err == nil && fi.IsDir():
		return nil
	case err == nil && !fi.IsDir():
		return fmt.Errorf("%s already exists as a regular file", path)
	case os.IsNotExist(err):
		return os.MkdirAll(path, 0750)
	case err != nil:
		return err
	}

	return nil
}

// appendSyntaxToFiles walks the temporary directory finding all the rendered
// Markdown files we generated and proceeds to append the required template
// syntax that the tfplugindocs tool needs.
func appendSyntaxToFiles(tempDir string) {
	filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".tmpl" {
			// open file for appending and in write-only mode
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := f.Write([]byte("{{ .SchemaMarkdown | trimspace }}\n")); err != nil {
				f.Close() // ignore error; Write error takes precedence
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
}

// backupTemplatesDir renames the /templates directory.
//
// We do this so that we can create a new /templates directory and move the
// contents of the temporary directory into the new location. Thus allowing
// tfplugindocs to be run within the root of the terraform provider repo.
func backupTemplatesDir(tmplDir string) {
	err := os.Rename(tmplDir, tmplDir+"-backup")
	if err != nil {
		log.Fatal(err)
	}
}

// replaceTemplatesDir removes the template directory from the repo and moves
// the temporary directory to where the template one would have been.
func replaceTemplatesDir(tmplDir string, tempDir string) {
	err := os.RemoveAll(tmplDir)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Rename(tempDir, tmplDir)
	if err != nil {
		log.Fatal(err)
	}
}
