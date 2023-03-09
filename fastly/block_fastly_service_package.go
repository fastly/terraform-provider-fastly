package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
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
				"content": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The contents of the Wasm deployment package (e.g. could be provided using an input variable or via external data source output variable). Conflicts with `filename`. Exactly one of these two arguments must be specified",
				},
				"filename": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The path to the Wasm deployment package within your local filesystem. Conflicts with `content`. Exactly one of these two arguments must be specified",
				},
				// sha512 hash of the file
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
		input := &gofastly.UpdatePackageInput{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
		}

		// Schema guarantees one package block.
		pkg := v.([]any)[0].(map[string]any)

		if v := pkg["content"].(string); v != "" {
			input.PackageContent = []byte(v)
		}
		if v := pkg["filename"].(string); v != "" {
			input.PackagePath = v
		}

		err := updatePackage(conn, input)
		if err != nil {
			return fmt.Errorf("error modifying package %s: %s", d.Id(), err)
		}
	}

	return nil
}

type PkgType int64

const (
	PkgUndefined PkgType = iota
	PkgContent
	PkgFilename
)

// Read refreshes the attribute state against the Fastly API.
func (h *PackageServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	localState := d.Get(h.key).([]any)

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing package for (%s)", d.Id())
		remoteState, err := conn.GetPackage(&gofastly.GetPackageInput{
			ServiceID:      d.Id(),
			ServiceVersion: s.ActiveVersion.Number,
		})
		if err != nil {
			if err, ok := err.(*gofastly.HTTPError); ok && err.IsNotFound() {
				log.Printf("[WARN] No wasm Package found for (%s), version (%v): %v", d.Id(), s.ActiveVersion.Number, err)
				_ = d.Set(h.GetKey(), nil)
				return nil
			}
			return fmt.Errorf("error looking up Package for (%s), version (%v): %v", d.Id(), s.ActiveVersion.Number, err)
		}

		var (
			pkgData string
			pkgType PkgType
		)

		if v := d.Get("package.0.content").(string); v != "" {
			pkgData = v
			pkgType = PkgContent
		}
		if v := d.Get("package.0.filename").(string); v != "" {
			pkgData = v
			pkgType = PkgFilename
		}

		wp := flattenPackage(remoteState, pkgType, pkgData)
		if err := d.Set(h.GetKey(), wp); err != nil {
			log.Printf("[WARN] Error setting Package for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

func updatePackage(conn *gofastly.Client, i *gofastly.UpdatePackageInput) error {
	_, err := conn.UpdatePackage(i)
	return err
}

// flattenPackage models data into format suitable for saving to Terraform state.
func flattenPackage(remoteState *gofastly.Package, pkgType PkgType, pkgData string) []map[string]any {
	var result []map[string]any
	data := map[string]any{
		"source_code_hash": remoteState.Metadata.HashSum,
	}

	switch pkgType {
	case PkgContent:
		data["content"] = pkgData
		data["filename"] = ""
	case PkgFilename:
		data["content"] = ""
		data["filename"] = pkgData
	}

	// Convert Package to a map for saving to state.
	result = append(result, data)
	return result
}
