package fastly

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyPackageHash() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyPackageHashRead,

		Schema: map[string]*schema.Schema{
			"content": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The contents of the Wasm deployment package as a base64 encoded string (e.g. could be provided using an input variable or via external data source output variable). Conflicts with `filename`. Exactly one of these two arguments must be specified",
				ExactlyOneOf: []string{"filename"},
			},
			"filename": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The path to the Wasm deployment package within your local filesystem. Conflicts with `content`. Exactly one of these two arguments must be specified",
				ExactlyOneOf: []string{"content"},
			},
			"hash": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A SHA512 hash of all files (in sorted order) within the package.",
			},
		},
	}
}

func dataSourceFastlyPackageHashRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	log.Printf("[DEBUG] Generating Package hash")

	pkg := "package.tar.gz"
	filename := d.Get("filename").(string)

	if filename != "" {
		pkg = filename
	} else {
		content := d.Get("content").(string)

		data, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return diag.Errorf("failed to decode base64 content: %s", err)
		}

		err = os.WriteFile(pkg, data, 0o644)
		if err != nil {
			return diag.Errorf("failed to write decoded base64 content to disk: %s", err)
		}

		defer os.Remove(pkg)
	}

	var (
		err error
		r   io.Reader
	)
	// G304 (CWE-22): Potential file inclusion via variable
	// #nosec
	r, err = os.Open(pkg)
	if err != nil {
		return diag.Errorf("failed to open package '%s': %s", pkg, err)
	}

	zr, err := gzip.NewReader(r)
	if err != nil {
		return diag.Errorf("failed to create a gzip reader: %s", err)
	}

	files, err := readFilesFromPackage(tar.NewReader(zr))
	if err != nil {
		return diag.Errorf("failed to read files within the package: %s", err)
	}

	hash, err := getFilesHash(files)
	if err != nil {
		return diag.Errorf("failed to generate hash from package files: %s", err)
	}

	d.SetId(hash)

	if err := d.Set("hash", hash); err != nil {
		return diag.Errorf("error setting package hash: %s", err)
	}

	return nil
}

// https://developer.fastly.com/learning/compute/#limitations-and-constraints
const maxPackageSize int64 = 100000000 // 100MB in bytes

// readFilesFromPackage reads all files within the provided package tar and
// generates a map data structure where the key is the filename and the value is
// the file contents.
func readFilesFromPackage(tr *tar.Reader) (map[string]*bytes.Buffer, error) {
	// Store the content of every file within the package.
	contents := make(map[string]*bytes.Buffer)

	// Track overall package size.
	var pkgSize int64

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Avoids G110: Potential DoS vulnerability via decompression bomb (gosec).
		pkgSize += hdr.Size
		if pkgSize > maxPackageSize {
			return nil, errors.New("package size exceeded 100MB limit")
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		contents[hdr.Name] = &bytes.Buffer{}

		_, err = io.CopyN(contents[hdr.Name], tr, hdr.Size)
		if err != nil {
			return nil, err
		}
	}

	return contents, nil
}

// getFilesHash returns a hash of all the filecontent in sorted filename order.
func getFilesHash(contents map[string]*bytes.Buffer) (string, error) {
	keys := make([]string, 0, len(contents))
	for k := range contents {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha512.New()
	for _, fname := range keys {
		if _, err := io.Copy(h, contents[fname]); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
