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
					Computed:    true,
					Description: `Used to trigger updates. Is automatically set to a SHA512 hash of the package file.`,
				},
			},
		},
	}
	return nil
}

// Process creates or updates the attribute against the Fastly API.
func (h *PackageServiceAttributeHandler) Process(_ context.Context, d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	id := d.Id()
	log.Printf("[DEBUG] Processing package for Service ID %s", id)
	if v, ok := d.GetOk(h.GetKey()); ok {
		// Schema guarantees one package block.
		pkg := v.([]interface{})[0].(map[string]interface{})

		// NOTE: The schema.ResourceData now reflects the proposed diff plan.
		// This means the package data will show as if the diff has been applied.
		// e.g. if you removed the "filename" attribute, it may still show in the
		// statefile as having a value but here it will show as an empty string.
		packageURL := pkg["url"].(string)
		packageFilename := pkg["filename"].(string)

		// We don't overwrite packageFilename as we need to ensure the original
		// value can be persisted back to the state.
		packagePath := packageFilename

		if packageURL != "" {
			f, err := downloadPackage(packageURL)
			if err != nil {
				return err
			}

			filename := f.Name()
			defer os.Remove(filename)

			packagePath = filename
		}

		// NOTE: We can't prevent a package upload, even if the hash hasn't changed.
		// This is because we can't tell if HasChange() returned false because it
		// was the result of there being no change to the package's hashsum OR if
		// it was because this was the first time Process() was called, such as when
		// starting a new project.
		if d.HasChange("source_code_hash") {
			log.Print("[DEBUG] Package hash digest changed")
		}

		err := updatePackage(conn, &gofastly.UpdatePackageInput{
			ServiceID:      id,
			ServiceVersion: latestVersion,
			PackagePath:    packagePath,
		})
		if err != nil {
			return fmt.Errorf("error modifying package %s: %s", id, err)
		}

		// Once we've safely uploaded the package, we can then ensure the state file
		// is updated with the latest package hashsum.

		digest, err := fileSHA512(packagePath)
		if err != nil {
			return fmt.Errorf("unable to hash package content: %w", err)
		}
		log.Printf("[DEBUG] Package hash digest %s", digest)

		wp := flattenPackage(digest, packageFilename, packageURL)
		key := h.GetKey()
		if err := d.Set(key, wp); err != nil {
			log.Printf("[WARN] Error setting Package for (%s): %s", id, err)
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

	// The package schema dictates only one of `url` or `filename` can be set.
	// NOTE: d.Get() is pulling the data from the state file.
	// TODO: Interpolate h.GetKey() to avoid error if package resource changes.
	packageURL := d.Get("package.0.url").(string)
	packageFilename := d.Get("package.0.filename").(string)

	// This hash represents the 'active' package's hash (coming from the API)
	hashSum := pkg.Metadata.HashSum

	// We don't overwrite packageFilename as we need to ensure the original
	// value can be persisted back to the state.
	filename := packageFilename

	if packageURL != "" {
		// TODO: Check if hashes differ, only then let file persist to disk.
		// If the hashes don't differ, then Process() won't be called. If we allow
		// the file to persist to disk, because we know Process() will be called,
		// then we can have Process() handle the file cleanup. This will avoid the
		// current situation, which is both Read() and Process() downloading the
		// package.
		//
		// 🚨 Read() doesn't always get called before Process(). If this is the
		// first time that `terraform plan` is called (e.g. a new project) then
		// there is no state, and so Terraform will go straight to `Process()`.
		f, err := downloadPackage(packageURL)
		if err != nil {
			return err
		}

		filename = f.Name()
		defer os.Remove(filename)
	}

	localDigest, err := fileSHA512(filename)
	if err != nil {
		return fmt.Errorf("unable to hash package content: %w", err)
	}
	log.Printf("[DEBUG] Package hash digest %s", localDigest)

	// If the specified package file has a different hash to what the API
	// returns (e.g. the active package is not the same code) then we'll update
	// the state to use the new hash, otherwise the hash from the API is used.
	//
	// The potential change in the state's hash value will cause Terraform to
	// conclude there is a difference and will need to execute Process().
	if localDigest != hashSum {
		hashSum = localDigest
		log.Print("[DEBUG] Package hash does not match the API returned package hash")
	} else {
		log.Print("[DEBUG] Package hash matches the API returned package hash")
	}

	// Update the state file before comparing to the Terraform configuration file...

	wp := flattenPackage(hashSum, packageFilename, packageURL)
	key := h.GetKey()

	log.Printf("[DEBUG] Setting state for key '%s'", key)
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

// downloadPackage downloads a package from the endpoint defined in the "url"
// attribute of the user's Terraform configuration, and returns the file
// descriptor for the downloaded package.
func downloadPackage(packageURL string) (f *os.File, err error) {
	f, err = os.CreateTemp("", "package-*.tar.gz")
	if err != nil {
		return f, fmt.Errorf("unable to create a temporary file to copy package data into: %w", err)
	}
	log.Printf("[DEBUG] Temp Package file %s", f.Name())

	resp, err := http.Get(packageURL)
	if err != nil {
		return f, fmt.Errorf("unable to download package '%s': %w", packageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return f, fmt.Errorf("bad package response '%s': %s", packageURL, resp.Status)
	}
	log.Printf("[DEBUG] Downloaded Package file from %s", packageURL)

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return f, fmt.Errorf("unable copy package into temporary file: %w", err)
	}

	return f, err
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
