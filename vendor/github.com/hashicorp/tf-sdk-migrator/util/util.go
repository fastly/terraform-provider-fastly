package util

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

var printConfig = printer.Config{
	Mode:     printer.TabIndent | printer.UseSpaces,
	Tabwidth: 8,
}

func StringSliceContains(ss []string, s string) bool {
	for _, i := range ss {
		if i == s {
			return true
		}
	}
	return false
}

func ReadOneOf(dir string, filenames ...string) (fullpath string, content []byte, err error) {
	for _, filename := range filenames {
		fullpath = filepath.Join(dir, filename)
		content, err = ioutil.ReadFile(fullpath)
		if err == nil {
			break
		}
	}
	return
}

func SearchLines(lines []string, search string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], search) {
			return i
		}
	}
	return -1
}

func SearchLinesPrefix(lines []string, search string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], search) {
			return i
		}
	}
	return -1
}

func GetProviderPath(providerRepoName string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Printf("GOPATH is empty")
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	paths := append([]string{wd}, filepath.SplitList(gopath)...)

	for _, p := range paths {
		fullPath := filepath.Join(p, "src", providerRepoName)
		info, err := os.Stat(fullPath)

		if err == nil {
			if !info.IsDir() {
				return "", fmt.Errorf("%s is not a directory", fullPath)
			} else {
				return fullPath, nil
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}

	return "", fmt.Errorf("Could not find %s in working directory or GOPATH: %s", providerRepoName, gopath)
}

func RewriteGoMod(providerPath string, sdkVersion string, oldPackagePath string, newPackagePath string) error {
	goModPath := filepath.Join(providerPath, "go.mod")

	input, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return err
	}

	pf, err := modfile.Parse(goModPath, input, nil)
	if err != nil {
		return err
	}

	err = pf.DropRequire(oldPackagePath)
	if err != nil {
		return err
	}

	pf.AddNewRequire(newPackagePath, sdkVersion, false)

	pf.Cleanup()
	formattedOutput, err := pf.Format()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(goModPath, formattedOutput, 0644)
	if err != nil {
		return err
	}

	return nil
}

func RewriteImportedPackageImports(filePath string, stringToReplace string, replacement string) error {
	if _, err := os.Stat(filePath); err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, impSpec := range f.Imports {
		impPath, err := strconv.Unquote(impSpec.Path.Value)
		if err != nil {
			log.Print(err)
		}
		// prevent partial matches on package names
		if impPath == stringToReplace || strings.HasPrefix(impPath, stringToReplace+"/") {
			newImpPath := strings.Replace(impPath, stringToReplace, replacement, -1)
			impSpec.Path.Value = strconv.Quote(newImpPath)
		}
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	w := bufio.NewWriter(out)
	if err := printConfig.Fprint(w, fset, f); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

func GoModTidy(providerPath string) error {
	args := []string{"go", "mod", "tidy"}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Dir = providerPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("[DEBUG] Executing command %q", args)
	err := cmd.Run()
	if err != nil {
		return NewExecError(err, stderr.String())
	}

	return nil
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
