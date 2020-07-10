package fastly

import (
	"errors"
	"fmt"
	"log"
	"time"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var fastlyNoServiceFoundErr = errors.New("No matching Fastly Service found")

const (
	// ServiceTypeVCL is the type for VCL services.
	ServiceTypeVCL = "vcl"
	// ServiceTypeCompute is the type for Compute services.
	ServiceTypeCompute = "wasm"
)

// ServiceAttributeDefinition provides an interface for service attributes.
// We compose a service resource out of attribute objects to allow us to construct both the VCL and Compute service
// resources from common components.
type ServiceAttributeDefinition interface {
	// Register add the attribute to the resource schema.
	Register(s *schema.Resource) error

	// Read refreshes the attribute state against the Fastly API.
	Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error

	// Process creates or updates the attribute against the Fastly API.
	Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error

	// HasChange returns whether the state of the attribute has changed against Terraform stored state.
	HasChange(d *schema.ResourceData) bool

	// MustProcess returns whether we must process the resource (usually HasChange==true but allowing exceptions).
	// For example: at present, the settings attributeHandler (block_fastly_service_v1_settings.go) must process when
	// default_ttl==0 and it is the initialVersion - as well as when default_ttl or default_host have changed.
	MustProcess(d *schema.ResourceData, initialVersion bool) bool
}

// ServiceMetadata provides a container to pass service attributes into an Attribute handler.
type ServiceMetadata struct {
	serviceType string
}

// DefaultServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type DefaultServiceAttributeHandler struct {
	key             string
	serviceMetadata ServiceMetadata
}

// GetKey is provided since most attributes will just use their private "key" for interacting with the service.
func (h *DefaultServiceAttributeHandler) GetKey() string {
	return h.key
}

// GetServiceMetadata is provided to allow internal methods to get the service Metadata
func (h *DefaultServiceAttributeHandler) GetServiceMetadata() ServiceMetadata {
	return h.serviceMetadata
}

// See interface definition for comments.
func (h *DefaultServiceAttributeHandler) HasChange(d *schema.ResourceData) bool {
	return d.HasChange(h.key)
}

// See interface definition for comments.
func (h *DefaultServiceAttributeHandler) MustProcess(d *schema.ResourceData, initialVersion bool) bool {
	return h.HasChange(d)
}

type VCLLoggingAttributes struct {
	format            string
	formatVersion     uint
	placement         string
	responseCondition string
}

// getVCLLoggingAttributes provides default values to Compute services for VCL only logging attributes
func (h *DefaultServiceAttributeHandler) getVCLLoggingAttributes(data map[string]interface{}) VCLLoggingAttributes {
	var vla = VCLLoggingAttributes{
		placement: "none",
	}
	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		if val, ok := data["format"]; ok {
			vla.format = val.(string)
		}
		if val, ok := data["format_version"]; ok {
			vla.formatVersion = uint(val.(int))
		}
		if val, ok := data["placement"]; ok {
			vla.placement = val.(string)
		}
		if val, ok := data["response_condition"]; ok {
			vla.responseCondition = val.(string)
		}
	}
	return vla
}

// ServiceDefinition defines the data model for service definitions
// There are two types of service: VCL and Compute. This interface specifies the data object from which service resources
// are constructed.
type ServiceDefinition interface {
	// GetType returns whether this is a VCL or Compute service.
	GetType() string

	// GetAttributeHandler returns the list of attributes/handlers supported by this service.
	GetAttributeHandler() []ServiceAttributeDefinition
}

// BaseServiceDefinition is the base implementation of the BaseServiceDefinition interface.
type BaseServiceDefinition struct {
	Attributes []ServiceAttributeDefinition
	Type       string
}

func (d *BaseServiceDefinition) GetType() string {
	return d.Type
}

func (d *BaseServiceDefinition) GetAttributeHandler() []ServiceAttributeDefinition {
	return d.Attributes
}

// resourceService returns a Terraform resource schema for VCL or Compute.
func resourceService(serviceDef ServiceDefinition) *schema.Resource {
	s := &schema.Resource{
		Create: resourceCreate(serviceDef),
		Read:   resourceRead(serviceDef),
		Update: resourceUpdate(serviceDef),
		Delete: resourceDelete(serviceDef),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name for this Service",
			},

			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Managed by Terraform",
				Description: "A personal freeform descriptive note",
			},

			"version_comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A personal freeform descriptive note",
			},

			// Active Version represents the currently activated version in Fastly. In
			// Terraform, we abstract this number away from the users and manage
			// creating and activating. It's used internally, but also exported for
			// users to see.
			"active_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			// Cloned Version represents the latest cloned version by the provider. It
			// gets set whenever Terraform detects changes and clones the currently
			// activated version in order to modify it. Active Version and Cloned
			// Version can be different if the Activate field is set to false in order
			// to prevent the service from being activated. It is not used internally,
			// but it is exported for users to see after running `terraform apply`.
			"cloned_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"activate": {
				Type:        schema.TypeBool,
				Description: "Conditionally prevents the Service from being activated",
				Default:     true,
				Optional:    true,
			},

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}

	// This loops over all the attribute handlers in the service definition and calls Register.
	// Register adds schema attributes to the overall schema for the resource. This allows each AttributeHandler to
	// define it's own attributes while allowing the overall set to be composed.
	for _, a := range serviceDef.GetAttributeHandler() {
		a.Register(s) // Mutates s, adding handler-specific schema items to the list.
	}

	return s
}

// resourceCreate satisfies the Terraform resource schema Create "interface"
// while injecting the ServiceDefinition into the true Create functionality.
func resourceCreate(serviceDef ServiceDefinition) schema.CreateFunc {
	return func(data *schema.ResourceData, i interface{}) error {
		return resourceServiceCreate(data, i, serviceDef)
	}
}

// resourceRead satisfies the Terraform resource schema Read "interface"
// while injecting the ServiceDefinition into the true Read functionality.
func resourceRead(serviceDef ServiceDefinition) schema.ReadFunc {
	return func(data *schema.ResourceData, i interface{}) error {
		return resourceServiceRead(data, i, serviceDef)
	}
}

// resourceUpdate satisfies the Terraform resource schema Update "interface"
// while injecting the ServiceDefinition into the true Update functionality.
func resourceUpdate(serviceDef ServiceDefinition) schema.UpdateFunc {
	return func(data *schema.ResourceData, i interface{}) error {
		return resourceServiceUpdate(data, i, serviceDef)
	}
}

// resourceDelete satisfies the Terraform resource schema Delete "interface"
// while injecting the ServiceDefinition into the true Delete functionality.
func resourceDelete(serviceDef ServiceDefinition) schema.DeleteFunc {
	return func(data *schema.ResourceData, i interface{}) error {
		return resourceServiceDelete(data, i, serviceDef)
	}
}

// resourceServiceCreate provides service resource Create functionality.
func resourceServiceCreate(d *schema.ResourceData, meta interface{}, serviceDef ServiceDefinition) error {
	if err := validateVCLs(d); err != nil {
		return err
	}

	conn := meta.(*FastlyClient).conn
	service, err := conn.CreateService(&gofastly.CreateServiceInput{
		Name:    d.Get("name").(string),
		Comment: d.Get("comment").(string),
		Type:    serviceDef.GetType(),
	})

	if err != nil {
		return err
	}

	d.SetId(service.ID)
	return resourceServiceUpdate(d, meta, serviceDef)
}

// resourceServiceUpdate provides service resource Update functionality.
func resourceServiceUpdate(d *schema.ResourceData, meta interface{}, serviceDef ServiceDefinition) error {
	if err := validateVCLs(d); err != nil {
		return err
	}

	conn := meta.(*FastlyClient).conn

	// Update Name and/or Comment. No new version is required for this.
	if d.HasChange("name") || d.HasChange("comment") {
		_, err := conn.UpdateService(&gofastly.UpdateServiceInput{
			ID:      d.Id(),
			Name:    d.Get("name").(string),
			Comment: d.Get("comment").(string),
		})
		if err != nil {
			return err
		}
	}

	// Once activated, Versions are locked and become immutable.
	// This loops over all AttributeHandlers calling HasChange. In this way each attribute handler can contribute
	// whether their current state and proposed changes mean a new version must be created.
	// So where changes are required, a new version must be created first, and updates posted to that
	// version. We only need one change to trigger this, so a break is OK.
	var needsChange bool
	for _, a := range serviceDef.GetAttributeHandler() {
		if a.HasChange(d) {
			needsChange = true
			break
		}
	}

	// Update the active version's comment. No new version is required for this.
	if d.HasChange("version_comment") && !needsChange {
		latestVersion := d.Get("active_version").(int)
		if latestVersion == 0 {
			// If the service was just created, there is an empty Version 1 available
			// that is unlocked and can be updated.
			latestVersion = 1
		}

		opts := gofastly.UpdateVersionInput{
			Service: d.Id(),
			Version: latestVersion,
			Comment: d.Get("version_comment").(string),
		}

		log.Printf("[DEBUG] Update Version opts: %#v", opts)
		_, err := conn.UpdateVersion(&opts)
		if err != nil {
			return err
		}
	}

	initialVersion := false

	if needsChange {
		latestVersion := d.Get("active_version").(int)
		if latestVersion == 0 {
			initialVersion = true
			// If the service was just created, there is an empty Version 1 available
			// that is unlocked and can be updated.
			latestVersion = 1
		} else {
			// Clone the latest version, giving us an unlocked version we can modify.
			log.Printf("[DEBUG] Creating clone of version (%d) for updates", latestVersion)
			newVersion, err := conn.CloneVersion(&gofastly.CloneVersionInput{
				Service: d.Id(),
				Version: latestVersion,
			})
			if err != nil {
				return err
			}

			// The new version number is named "Number", but it's actually a string.
			latestVersion = newVersion.Number
			d.Set("cloned_version", latestVersion)

			// New versions are not immediately found in the API, or are not
			// immediately mutable, so we need to sleep a few and let Fastly ready
			// itself. Typically, 7 seconds is enough.
			log.Print("[DEBUG] Sleeping 7 seconds to allow Fastly Version to be available")
			time.Sleep(7 * time.Second)

			// Update the cloned version's comment.
			if d.Get("version_comment").(string) != "" {
				opts := gofastly.UpdateVersionInput{
					Service: d.Id(),
					Version: latestVersion,
					Comment: d.Get("version_comment").(string),
				}

				log.Printf("[DEBUG] Update Version opts: %#v", opts)
				_, err := conn.UpdateVersion(&opts)
				if err != nil {
					return err
				}
			}
		}

		// This delegates the bulk of processing to attribute handlers which manage state
		// for their own attributes.
		for _, a := range serviceDef.GetAttributeHandler() {
			if a.MustProcess(d, initialVersion) {
				if err := a.Process(d, latestVersion, conn); err != nil {
					return err
				}
			}
		}

		// Validate version.
		log.Printf("[DEBUG] Validating Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
		valid, msg, err := conn.ValidateVersion(&gofastly.ValidateVersionInput{
			Service: d.Id(),
			Version: latestVersion,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error checking validation: %s", err)
		}

		if !valid {
			return fmt.Errorf("[ERR] Invalid configuration for Fastly Service (%s): %s", d.Id(), msg)
		}

		shouldActivate := d.Get("activate").(bool)
		if shouldActivate {
			log.Printf("[DEBUG] Activating Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
			_, err = conn.ActivateVersion(&gofastly.ActivateVersionInput{
				Service: d.Id(),
				Version: latestVersion,
			})
			if err != nil {
				return fmt.Errorf("[ERR] Error activating version (%d): %s", latestVersion, err)
			}

			// Only if the version is valid and activated do we set the active_version.
			// This prevents us from getting stuck in cloning an invalid version.
			d.Set("active_version", latestVersion)
		} else {
			log.Printf("[INFO] Skipping activation of Fastly Service (%s), Version (%v)", d.Id(), latestVersion)
			log.Print("[INFO] The Terraform definition is explicitly specified to not activate the changes on Fastly")
			log.Printf("[INFO] Version (%v) has been pushed and validated", latestVersion)
			log.Printf("[INFO] Visit https://manage.fastly.com/configure/services/%s/versions/%v and activate it manually", d.Id(), latestVersion)
		}
	}

	return resourceServiceRead(d, meta, serviceDef)
}

// resourceServiceRead provides service resource Read functionality.
func resourceServiceRead(d *schema.ResourceData, meta interface{}, serviceDef ServiceDefinition) error {
	conn := meta.(*FastlyClient).conn

	// Find the Service. Discard the service because we need the ServiceDetails,
	// not just a Service record.
	_, err := findService(d.Id(), meta)
	if err != nil {
		switch err {
		case fastlyNoServiceFoundErr:
			log.Printf("[WARN] %s for ID (%s)", err, d.Id())
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	s, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	d.Set("name", s.Name)
	d.Set("comment", s.Comment)
	d.Set("version_comment", s.Version.Comment)
	d.Set("active_version", s.ActiveVersion.Number)

	// If CreateService succeeds, but initial updates to the Service fail, we'll
	// have an empty ActiveService version (no version is active, so we can't
	// query for information on it).
	if s.ActiveVersion.Number != 0 {

		// This delegates read to all the attribute handlers which can then manage reading state for
		// their own attributes.
		for _, a := range serviceDef.GetAttributeHandler() {
			if err := a.Read(d, s, conn); err != nil {
				return err
			}
		}

	} else {
		log.Printf("[DEBUG] Active Version for Service (%s) is empty, no state to refresh", d.Id())
	}

	return nil
}

// resourceServiceDelete provides service resource Delete functionality.
func resourceServiceDelete(d *schema.ResourceData, meta interface{}, serviceDef ServiceDefinition) error {
	conn := meta.(*FastlyClient).conn

	// Fastly will fail to delete any service with an Active Version.
	// If `force_destroy` is given, we deactivate the active version and then send
	// the DELETE call.
	if d.Get("force_destroy").(bool) {
		s, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ID: d.Id(),
		})

		if err != nil {
			return err
		}

		if s.ActiveVersion.Number != 0 {
			_, err := conn.DeactivateVersion(&gofastly.DeactivateVersionInput{
				Service: d.Id(),
				Version: s.ActiveVersion.Number,
			})
			if err != nil {
				return err
			}
		}
	}

	err := conn.DeleteService(&gofastly.DeleteServiceInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	_, err = findService(d.Id(), meta)
	if err != nil {
		switch err {
		// We expect no records to be found here.
		case fastlyNoServiceFoundErr:
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	// findService above returned something and nil error, but shouldn't have.
	return fmt.Errorf("[WARN] Tried deleting Service (%s), but was still found", d.Id())

}

// findService finds a Fastly Service via the ListServices endpoint, returning
// the Service if found.
//
// Fastly API does not include any "deleted_at" type parameter to indicate
// that a Service has been deleted. GET requests to a deleted Service will
// return 200 OK and have the full output of the Service for an unknown time
// (days, in my testing). In order to determine if a Service is deleted, we
// need to hit /service and loop the returned Services, searching for the one
// in question. This endpoint only returns active or "alive" services. If the
// Service is not included, then it's "gone".
//
// Returns a fastlyNoServiceFoundErr error if the Service is not found in the
// ListServices response.
func findService(id string, meta interface{}) (*gofastly.Service, error) {
	conn := meta.(*FastlyClient).conn

	l, err := conn.ListServices(&gofastly.ListServicesInput{})
	if err != nil {
		return nil, fmt.Errorf("[WARN] Error listing services (%s): %s", id, err)
	}

	for _, s := range l {
		if s.ID == id {
			log.Printf("[DEBUG] Found Service (%s)", id)
			return s, nil
		}
	}

	return nil, fastlyNoServiceFoundErr
}
