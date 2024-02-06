package fastly

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
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
		Optional:    true,
		Description: "The `package` block supports uploading or modifying Wasm packages for use in a Fastly Compute service (if omitted, ensure `activate = false` is set on `fastly_service_compute` to avoid service validation errors). See Fastly's documentation on [Compute](https://developer.fastly.com/learning/compute/)",
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"content": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The contents of the Wasm deployment package as a base64 encoded string (e.g. could be provided using an input variable or via external data source output variable). Conflicts with `filename`. Exactly one of these two arguments must be specified",
					ExactlyOneOf: []string{"package.0.content", "package.0.filename"},
				},
				"filename": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The path to the Wasm deployment package within your local filesystem. Conflicts with `content`. Exactly one of these two arguments must be specified",
					ExactlyOneOf: []string{"package.0.content", "package.0.filename"},
				},
				"source_code_hash": {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					Description:   "Used to trigger updates. Must be set to a SHA512 hash of all files (in sorted order) within the package. The usual way to set this is using the fastly_package_hash data source.",
					ConflictsWith: []string{"package.0.content"},
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
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return fmt.Errorf("error decoding base64 string for package %s: %s", d.Id(), err)
			}
			input.PackageContent = []byte(decoded)
		}
		if v := pkg["filename"].(string); v != "" {
			input.PackagePath = gofastly.ToPointer(v)
		}

		_, err := conn.UpdatePackage(input)
		if err != nil {
			return fmt.Errorf("error modifying package %s: %s", d.Id(), err)
		}
	}

	return nil
}

type PkgType int64

const (
	_ PkgType = iota
	PkgContent
	PkgFilename
)

// Read refreshes the attribute state against the Fastly API.
func (h *PackageServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	localState := d.Get(h.key).([]any)

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		if s.ActiveVersion == nil {
			return fmt.Errorf("error: no service ActiveVersion object")
		}
		serviceVersionNumber := gofastly.ToValue(s.ActiveVersion.Number)

		log.Printf("[DEBUG] Refreshing package for (%s)", d.Id())
		remoteState, err := conn.GetPackage(&gofastly.GetPackageInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersionNumber,
		})
		if err != nil {
			if err, ok := err.(*gofastly.HTTPError); ok && err.IsNotFound() {
				log.Printf("[WARN] No wasm Package found for (%s), version (%v): %v", d.Id(), serviceVersionNumber, err)
				_ = d.Set(h.GetKey(), nil)
				return nil
			}
			return fmt.Errorf("error looking up Package for (%s), version (%v): %v", d.Id(), serviceVersionNumber, err)
		}

		var (
			pkgData string
			pkgType PkgType
		)

		// The value is provided by the user's config as the API doesn't return it.
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

// flattenPackage models data into format suitable for saving to Terraform state.
func flattenPackage(remoteState *gofastly.Package, pkgType PkgType, pkgData string) []map[string]any {
	var result []map[string]any

	data := map[string]any{}

	if remoteState.Metadata != nil && remoteState.Metadata.FilesHash != nil {
		data["source_code_hash"] = *remoteState.Metadata.FilesHash
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
