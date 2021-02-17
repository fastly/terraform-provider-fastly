package list

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

// types extracted from the output of `go help list`

type Package struct {
	Dir           string   // directory containing package sources
	ImportPath    string   // import path of package in dir
	ImportComment string   // path in import comment on package statement
	Name          string   // package name
	Doc           string   // package documentation string
	Target        string   // install path
	Shlib         string   // the shared library that contains this package (only set when -linkshared)
	Goroot        bool     // is this package in the Go root?
	Standard      bool     // is this package part of the standard Go library?
	Stale         bool     // would 'go install' do anything for this package?
	StaleReason   string   // explanation for Stale==true
	Root          string   // Go root or Go path dir containing this package
	ConflictDir   string   // this directory shadows Dir in $GOPATH
	BinaryOnly    bool     // binary-only package: cannot be recompiled from sources
	ForTest       string   // package is only for use in named test
	Export        string   // file containing export data (when using -export)
	Module        *Module  // info about package's containing module, if any (can be nil)
	Match         []string // command-line patterns matching this package
	DepOnly       bool     // package is only a dependency, not explicitly listed

	// Source files
	GoFiles         []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles        []string // .go source files that import "C"
	CompiledGoFiles []string // .go files presented to compiler (when using -compiled)
	IgnoredGoFiles  []string // .go source files ignored due to build constraints
	CFiles          []string // .c source files
	CXXFiles        []string // .cc, .cxx and .cpp source files
	MFiles          []string // .m source files
	HFiles          []string // .h, .hh, .hpp and .hxx source files
	FFiles          []string // .f, .F, .for and .f90 Fortran source files
	SFiles          []string // .s source files
	SwigFiles       []string // .swig files
	SwigCXXFiles    []string // .swigcxx files
	SysoFiles       []string // .syso object files to add to archive
	TestGoFiles     []string // _test.go files in package
	XTestGoFiles    []string // _test.go files outside package

	// Cgo directives
	CgoCFLAGS    []string // cgo: flags for C compiler
	CgoCPPFLAGS  []string // cgo: flags for C preprocessor
	CgoCXXFLAGS  []string // cgo: flags for C++ compiler
	CgoFFLAGS    []string // cgo: flags for Fortran compiler
	CgoLDFLAGS   []string // cgo: flags for linker
	CgoPkgConfig []string // cgo: pkg-config names

	// Dependency information
	Imports      []string          // import paths used by this package
	ImportMap    map[string]string // map from source import to ImportPath (identity entries omitted)
	Deps         []string          // all (recursively) imported dependencies
	TestImports  []string          // imports from TestGoFiles
	XTestImports []string          // imports from XTestGoFiles

	// Error information
	Incomplete bool            // this package or a dependency has an error
	Error      *PackageError   // error loading package
	DepsErrors []*PackageError // errors loading dependencies
}

type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}

type Context struct {
	GOARCH        string   // target architecture
	GOOS          string   // target operating system
	GOROOT        string   // Go root
	GOPATH        string   // Go path
	CgoEnabled    bool     // whether cgo can be used
	UseAllFiles   bool     // use files regardless of +build lines, file names
	Compiler      string   // compiler to assume when computing target paths
	BuildTags     []string // build constraints to match in +build lines
	ReleaseTags   []string // releases the current release is compatible with
	InstallSuffix string   // suffix to use in the name of the install dir
}

type Module struct {
	Path     string       // module path
	Version  string       // module version
	Versions []string     // available module versions (with -versions)
	Replace  *Module      // replaced by this module
	Time     *time.Time   // time version was created
	Update   *Module      // available update, if any (with -u)
	Main     bool         // is this the main module?
	Indirect bool         // is this module only an indirect dependency of main module?
	Dir      string       // directory holding files for this module, if any
	GoMod    string       // path to go.mod file for this module, if any
	Error    *ModuleError // error loading module
}

type ModuleError struct {
	Err string // the error itself
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

func GoList(workDir string, path string, args ...string) ([]Package, error) {
	cmdName := "go"
	cmdArgs := append([]string{"list", "-json"}, args...)
	cmdArgs = append(cmdArgs, path)
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("[DEBUG] Executing command %s %q", cmdName, cmdArgs)

	err := cmd.Run()
	if err != nil {
		return nil, NewExecError(err, stderr.String())
	}

	packages := []Package{}
	dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))

	for {
		var p Package
		if err := dec.Decode(&p); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		packages = append(packages, p)
	}

	return packages, nil
}

func GoListModule(workDir string, path string, args ...string) ([]Module, error) {
	cmdName := "go"
	cmdArgs := append([]string{"list", "-m", "-json"}, args...)
	cmdArgs = append(cmdArgs, path)
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("[DEBUG] Executing command %s %q", cmdName, cmdArgs)

	err := cmd.Run()
	if err != nil {
		return nil, NewExecError(err, stderr.String())
	}

	modules := []Module{}
	dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))

	for {
		var m Module
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		modules = append(modules, m)
	}

	return modules, nil
}
