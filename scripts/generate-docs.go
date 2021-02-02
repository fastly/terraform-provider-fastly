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
// 5. copy the index.md file (which requires no pre-compiling) to the temporary directory so tfplugindocs can include it
// 6. rename repo /templates/{data-sources/resources} directories to avoid being overwritten by next step
// 7. move contents of temp directory (i.e. data-sources/resources) into repo /templates directory
// 8. run tfplugindocs generate function to output final documentation to /docs directory
// 9. replace /templates/{data-sources/resources} directories with their backed up equivalents
//
package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	tmplDir := baseDir + "/templates"

	tempDir, err := ioutil.TempDir("", "precompile")
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

	copyIndexToTempDir(tmplDir, tempDir)

	backupTemplatesDir(tmplDir)

	replaceTemplatesDir(tmplDir, tempDir)

	runTFPluginDocs()

	replaceTemplatesDir(tmplDir, tmplDir+"-backup")
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
	log.Println(p.path)
	if err != nil {
		log.Fatal(err)
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

// copyIndexToTempDir copies the non-templated index.md into our temporary
// directory so that when we come to replace the repo's /templates with our
// pre-compiled version, then the tfplugindocs command will be able to include
// the index.md in the generated output in the /docs directory.
//
// The reason the index.md isn't already in our temporary directory of
// pre-compiled Markdown templates is because the renderPages function is
// designed to include only files with a .tmpl extension, where as index.md
// doesn't require any templating.
func copyIndexToTempDir(tmplDir string, tempDir string) {
	filename := "/index.md"
	srcFile := tmplDir + filename
	dstFile := tempDir + filename

	src, err := os.Open(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	dst, err := os.Create(dstFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		log.Fatal(err)
	}
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

// runTFPluginDocs executes the tfplugindocs binary which generates
// documentation Markdown files from our terraform code, while also utilizing
// any templates we have defined in the /templates directory.
//
// NOTE: it is presumed that the /templates directory that is referenced will
// consist of precompiled templates and that the original untouched templates
// will still exist in the /templates-backup directory ready to be restored
// once the /docs content has been generated.
func runTFPluginDocs() {
	cmd := exec.Command("tfplugindocs", "generate")
	err := cmd.Run()
	if err != nil {
		log.Fatal()
	}
}
