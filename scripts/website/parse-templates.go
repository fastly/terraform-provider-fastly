//go:generate go run parse-templates.go
package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

type Page struct {
	name string
	path string
	data PageData
}

type PageData struct {
	ServiceType string
}





func main() {
	baseDir := getBaseDir()
	tmplDir := baseDir + "/website_src/docs/r/"
	docsDir := baseDir + "/website/docs/"

	var pages = []Page{
		{
			name: "service_v1",
			path: docsDir + "r/service_v1.html.markdown",
			data: PageData{
				"vcl",
			},
		},
		{
			name: "service_compute",
			path: docsDir + "r/service_compute.html.markdown",
			data: PageData{
				"wasm",
			},
		},
	}

	renderPages(getTemplate(tmplDir), pages)
}


func renderPages(t *template.Template, pages []Page) {
	for _, p := range pages {
		renderPage(t, p)
	}
}

func renderPage(t *template.Template, p Page) {
	f, err := os.Create(p.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = t.ExecuteTemplate(f, p.name, p.data)
	if err != nil {
		panic(err)
	}
}

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


func getBaseDir() string {
	_, scriptPath, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not get current working directory")
	}
	tpgDir := scriptPath
	for !strings.HasPrefix(filepath.Base(tpgDir), "terraform-provider-") && tpgDir != "/" {
		tpgDir = filepath.Clean(tpgDir + "/..")
	}
	if tpgDir == "/" {
		log.Fatal("Script was run outside of fastly provider directory")
	}
	return tpgDir
}

