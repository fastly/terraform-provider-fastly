package fastly

import (
	"context"
	"fmt"
	"log"

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
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The URL to download the Wasm deployment package.",
				},
				"filename": {
					Type:          schema.TypeString,
					Optional:      true,
					Description:   "The path to the Wasm deployment package within your local filesystem",
					ConflictsWith: []string{"url"},
				},
				// sha512 hash of the file
				"source_code_hash": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: `Used to trigger updates. Must be set to a SHA512 hash of the package file specified with the filename. The usual way to set this is filesha512("package.tar.gz") (Terraform 0.11.12 and later) or filesha512(file("package.tar.gz")) (Terraform 0.11.11 and earlier), where "package.tar.gz" is the local filename of the Wasm deployment package`,
				}, // TODO: How to handle `source_code_hash` for a url attribute?
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
		packageFilename := pkg["filename"].(string)

		// TODO: figure out how to switch between filename and url attributes.

		// TODO: https://stackoverflow.com/questions/11692860/how-can-i-efficiently-download-a-large-file-using-go

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

	// TODO: figure out what the structure for package.0.url is.

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
