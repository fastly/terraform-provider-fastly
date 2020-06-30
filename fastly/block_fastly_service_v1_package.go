package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type PackageServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServicePackage() ServiceAttributeDefinition {
	return &PackageServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key: "package",
		},
	}
}

func (h *PackageServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"filename": {
					Type:     schema.TypeString,
					Required: true,
				},
				// sha512 hash of the file
				"source_code_hash": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"source_code_size": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}
	return nil
}

func (h *PackageServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {

	if v, ok := d.GetOk(h.GetKey()); ok {
		// Schema guarantees one package block.
		Package := v.([]interface{})[0].(map[string]interface{})
		packageFilename := Package["filename"].(string)

		err := updatePackage(conn, &gofastly.UpdatePackageInput{
			Service:     d.Id(),
			Version:     latestVersion,
			PackagePath: packageFilename,
		})
		if err != nil {
			return fmt.Errorf("Error modifying package %s: %s", d.Id(), err)
		}
	}

	return nil
}

func (h *PackageServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing package for (%s)", d.Id())
	Package, err := conn.GetPackage(&gofastly.GetPackageInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Package for (%s), version (%v): %v", d.Id(), s.ActiveVersion.Number, err)
	}

	filename := d.Get("package.0.filename").(string)
	wp := flattenPackage(Package, filename)
	if err := d.Set(h.GetKey(), wp); err != nil {
		log.Printf("[WARN] Error setting Package for (%s): %s", d.Id(), err)
	}

	return nil
}

func updatePackage(conn *gofastly.Client, i *gofastly.UpdatePackageInput) error {
	_, err := conn.UpdatePackage(i)
	return err
}

func flattenPackage(Package *gofastly.Package, filename string) []map[string]interface{} {
	var pa []map[string]interface{}
	p := map[string]interface{}{
		"source_code_hash": Package.Metadata.HashSum,
		"source_code_size": Package.Metadata.Size,
		"filename":         filename,
	}

	// Convert Package to a map for saving to state.
	pa = append(pa, p)
	return pa
}
