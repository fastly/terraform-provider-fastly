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
		// schema.TypeList has an issue with ConflictsWith:
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/71
		Type:        schema.TypeList,
		Required:    true,
		Description: "The `package` block supports uploading or modifying Wasm packages for use in a Fastly Compute@Edge service. See Fastly's documentation on [Compute@Edge](https://developer.fastly.com/learning/compute/)",
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"url": {
					Type:          schema.TypeString,
					Optional:      true,
					Description:   "The URL to download the Wasm deployment package from.",
					ConflictsWith: []string{"package.0.filename"},
				},
				"filename": {
					Type:          schema.TypeString,
					Optional:      true,
					Description:   "The path to the Wasm deployment package within your local filesystem",
					ConflictsWith: []string{"package.0.url"},
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

		// NOTE: The schema.ResourceData now reflects the proposed diff plan.
		// This means the package data will show as if the diff has been applied.
		// e.g. if you removed the "filename" attribute, it may still show in the
		// statefile as having a value but here it will show as an empty string.
		packageURL := pkg["url"].(string)
		packageFilename := pkg["filename"].(string)

		if packageURL != "" {
			f, err := os.CreateTemp("", "package-*.tar.gz")
			if err != nil {
				return fmt.Errorf("unable to create a temporary file to copy package data into: %w", err)
			}
			log.Printf("[DEBUG] Temp Package file %s", f.Name())
			defer os.Remove(f.Name())

			resp, err := http.Get(packageURL)
			if err != nil {
				return fmt.Errorf("unable to download package '%s': %w", packageURL, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("bad package response '%s': %s", packageURL, resp.Status)
			}
			log.Printf("[DEBUG] Downloaded Package file from %s", packageURL)

			_, err = io.Copy(f, resp.Body)
			if err != nil {
				return fmt.Errorf("unable copy package into temporary file: %w", err)
			}

			digest, err := fileSHA512(f.Name())
			if err != nil {
				return fmt.Errorf("unable to hash package content: %w", err)
			}
			log.Printf("[DEBUG] Package hash digest %s", digest)
			sch := d.Get("source_code_hash")
			if sch != nil {
				v := sch.(string)
				fmt.Printf("current source_code_hash: %+v\n", v)
			}
			wp := flattenPackage(digest, packageFilename, packageURL)
			key := h.GetKey()
			if err := d.Set(key, wp); err != nil {
				log.Printf("[WARN] Error setting Package for (%s): %s", d.Id(), err)
			}

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
	id := d.Id()
	log.Printf("[DEBUG] Refreshing package for Service ID %s", id)
	pkg, err := conn.GetPackage(&gofastly.GetPackageInput{
		ServiceID:      id,
		ServiceVersion: s.ActiveVersion.Number,
	})
	if err != nil {
		if err, ok := err.(*gofastly.HTTPError); ok && err.IsNotFound() {
			log.Printf("[WARN] No wasm Package found for (%s), version (%v): %v", id, s.ActiveVersion.Number, err)
			d.Set(h.GetKey(), nil)
			return nil
		}
		return fmt.Errorf("error looking up Package for (%s), version (%v): %v", id, s.ActiveVersion.Number, err)
	}

	// d.Get() is pulling the data from the state file.
	// TODO: figure out why we don't use GetKey and am using hardcoded 'package'
	// name because if that key ever changes the code will break. we should
	// interpolate the value instead.
	packageURL := d.Get("package.0.url").(string)
	filename := d.Get("package.0.filename").(string)
	wp := flattenPackage(pkg.Metadata.HashSum, filename, packageURL)
	key := h.GetKey()
	if err := d.Set(key, wp); err != nil {
		log.Printf("[WARN] Error setting Package for (%s): %s", id, err)
	}

	return nil
}

func updatePackage(conn *gofastly.Client, i *gofastly.UpdatePackageInput) error {
	_, err := conn.UpdatePackage(i)
	return err
}

func flattenPackage(hashSum, filename, packageURL string) []map[string]interface{} {
	var pa []map[string]interface{}
	p := map[string]interface{}{
		"source_code_hash": hashSum,
		"filename":         filename,
		"url":              packageURL,
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
