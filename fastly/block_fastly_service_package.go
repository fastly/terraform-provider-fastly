package fastly

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// PackageServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type PackageServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServicePackage returns a new resource.
func NewServicePackage(sa ServiceMetadata) ServiceAttributeDefinition {
	return &PackageServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "package",
			serviceMetadata: sa,
		},
	}
}

// Register add the attribute to the resource schema.
func (h *PackageServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		Description: "The `package` block supports uploading or modifying Wasm packages for use in a Fastly Compute@Edge service. See Fastly's documentation on [Compute@Edge](https://developer.fastly.com/learning/compute/)",
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"url": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The URL to download the Wasm deployment package from.",
					ExactlyOneOf: []string{"url", "filename"},
				},
				"filename": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The path to the Wasm deployment package within your local filesystem",
					ExactlyOneOf: []string{"filename", "url"},
				},
				"source_code_hash": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: `Used to trigger updates. Must be set to a SHA512 hash of the package file specified with the filename. The usual way to set this is filesha512("package.tar.gz") (Terraform 0.11.12 and later) or filesha512(file("package.tar.gz")) (Terraform 0.11.11 and earlier), where "package.tar.gz" is the local filename of the Wasm deployment package`,
				},
			},
		},
	}
	return nil
}

// Process creates or updates the attribute against the Fastly API.
func (h *PackageServiceAttributeHandler) Process(_ context.Context, d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	if v, ok := d.GetOk(h.GetKey()); ok {
		// Schema guarantees one package block.
		pkg := v.([]interface{})[0].(map[string]interface{})

		packageURL := pkg["url"].(string)
		packageFilename := pkg["filename"].(string)

		if packageURL != "" {
			f, err := os.CreateTemp("", "package-*")
			if err != nil {
				return fmt.Errorf("unable to create a temporary file to copy package data into: %w", err)
			}
			defer os.Remove(f.Name())

			resp, err := http.Get(packageURL)
			if err != nil {
				return fmt.Errorf("unable to download package '%s': %w", packageURL, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("bad package response '%s': %s", packageURL, resp.Status)
			}

			_, err = io.Copy(f, resp.Body)
			if err != nil {
				return fmt.Errorf("unable copy package into temporary file: %w", err)
			}

			digest, err := fileSHA512(f.Name())
			if err != nil {
				return fmt.Errorf("unable to hash package content: %w", err)
			}
			d.Set("source_code_hash", digest)

			packageFilename = f.Name()
		}

		err := updatePackage(conn, &gofastly.UpdatePackageInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
			PackagePath:    packageFilename,
		})
		if err != nil {
			return fmt.Errorf("error modifying package %s: %s", d.Id(), err)
		}
	}

	return nil
}

// Read refreshes the attribute state against the Fastly API.
func (h *PackageServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing package for (%s)", d.Id())
	pkg, err := conn.GetPackage(&gofastly.GetPackageInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	})
	if err != nil {
		if err, ok := err.(*gofastly.HTTPError); ok && err.IsNotFound() {
			log.Printf("[WARN] No wasm Package found for (%s), version (%v): %v", d.Id(), s.ActiveVersion.Number, err)
			d.Set(h.GetKey(), nil)
			return nil
		}
		return fmt.Errorf("error looking up Package for (%s), version (%v): %v", d.Id(), s.ActiveVersion.Number, err)
	}

	// TODO: figure out what to do here? switch between url/filename like Process() method does?

	filename := d.Get("package.0.filename").(string)
	wp := flattenPackage(pkg, filename)
	if err := d.Set(h.GetKey(), wp); err != nil {
		log.Printf("[WARN] Error setting Package for (%s): %s", d.Id(), err)
	}

	return nil
}

func updatePackage(conn *gofastly.Client, i *gofastly.UpdatePackageInput) error {
	_, err := conn.UpdatePackage(i)
	return err
}

func flattenPackage(pkg *gofastly.Package, filename string) []map[string]interface{} {
	var pa []map[string]interface{}
	p := map[string]interface{}{
		"source_code_hash": pkg.Metadata.HashSum,
		"filename":         filename,
	}

	// Convert Package to a map for saving to state.
	pa = append(pa, p)
	return pa
}

// fileSHA512 reads the path content, hashes it with sha512 and hex encodes.
//
// The implementation is copied from HashiCorps internal implementation for the
// filesha512() function https://www.terraform.io/language/functions/filesha512:
//
// https://github.com/hashicorp/terraform/blob/31fc22a0d243a53f306eb41adb57b867aa170041/internal/lang/funcs/crypto.go#L236
// https://github.com/hashicorp/terraform/blob/31fc22a0d243a53f306eb41adb57b867aa170041/internal/lang/funcs/crypto.go#L214
func fileSHA512(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha512.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
