package fastly

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

func resourceACLV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceACLV1Create,
		Read:   resourceACLV1Read,
		Update: resourceACLV1Update,
		Delete: resourceACLV1Delete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected SERVICE-ID/NAME", d.Id())
				}

				serviceID := idParts[0]
				aclName := idParts[1]

				conn := meta.(*FastlyClient).conn
				service, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
					ID: serviceID,
				})

				if err != nil {
					return nil, err
				}

				acl, err := conn.GetACL(&gofastly.GetACLInput{
					Name:    aclName,
					Service: serviceID,
					Version: service.ActiveVersion.Number,
				})

				if err != nil {
					return nil, err
				}

				d.Set("name", acl.Name)
				d.Set("service_id", acl.ServiceID)
				d.Set("version", acl.Version)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name for this ACL",
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the service",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Version of the service",
			},
			"activate": {
				Type:        schema.TypeBool,
				Description: "Conditionally prevents the Service from being activated",
				Default:     true,
				Optional:    true,
			},
		},
	}
}

func resourceACLV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn
	serviceID := d.Get("service_id").(string)

	service, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ID: d.Get("service_id").(string),
	})

	if err != nil {
		return err
	}

	clonedVersion, err := conn.CloneVersion(&gofastly.CloneVersionInput{
		Service: serviceID,
		Version: service.ActiveVersion.Number,
	})

	if err != nil {
		return err
	}

	versionNumber := clonedVersion.Number

	acl, err := conn.CreateACL(&gofastly.CreateACLInput{
		Name:    d.Get("name").(string),
		Service: serviceID,
		Version: versionNumber,
	})

	if err != nil {
		return err
	}

	valid, message, err := conn.ValidateVersion(&gofastly.ValidateVersionInput{
		Service: serviceID,
		Version: versionNumber,
	})

	if err != nil {
		return err
	}

	if !valid {
		return errors.New(message)
	}

	if d.Get("activate").(bool) {
		activateVersion(serviceID, versionNumber, meta)
	}

	d.SetId(acl.ID)
	d.Set("version", acl.Version)
	return resourceACLV1Read(d, meta)
}

func resourceACLV1Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)

	service, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ID: d.Get("service_id").(string),
	})

	if err != nil {
		return err
	}

	clonedVersion, err := conn.CloneVersion(&gofastly.CloneVersionInput{
		Service: serviceID,
		Version: service.ActiveVersion.Number,
	})

	if err != nil {
		return err
	}

	versionNumber := clonedVersion.Number

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		_, err = conn.UpdateACL(&gofastly.UpdateACLInput{
			Name:    oldName.(string),
			NewName: newName.(string),
			Service: serviceID,
			Version: versionNumber,
		})

		if err != nil {
			return err
		}

		valid, message, err := conn.ValidateVersion(&gofastly.ValidateVersionInput{
			Service: serviceID,
			Version: versionNumber,
		})

		if err != nil {
			return err
		}

		if !valid {
			return errors.New(message)
		}
	}

	if d.Get("activate").(bool) {
		activateVersion(serviceID, versionNumber, meta)
	}

	d.Set("version", versionNumber)

	return resourceACLV1Read(d, meta)
}

func resourceACLV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	acl, err := conn.GetACL(&gofastly.GetACLInput{
		Name:    d.Get("name").(string),
		Service: d.Get("service_id").(string),
		Version: d.Get("version").(int),
	})

	if err != nil {
		return err
	}

	d.SetId(acl.ID)
	d.Set("name", acl.Name)
	d.Set("service_id", acl.ServiceID)
	d.Set("version", acl.Version)

	return nil
}

func resourceACLV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)

	service, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ID: d.Get("service_id").(string),
	})

	if err != nil {
		return err
	}

	clonedVersion, err := conn.CloneVersion(&gofastly.CloneVersionInput{
		Service: serviceID,
		Version: service.ActiveVersion.Number,
	})

	if err != nil {
		return err
	}

	versionNumber := clonedVersion.Number

	err = conn.DeleteACL(&gofastly.DeleteACLInput{
		Name:    d.Get("name").(string),
		Service: d.Get("service_id").(string),
		Version: versionNumber,
	})

	if err != nil {
		return err
	}

	valid, message, err := conn.ValidateVersion(&gofastly.ValidateVersionInput{
		Service: serviceID,
		Version: versionNumber,
	})

	if err != nil {
		return err
	}

	if !valid {
		return errors.New(message)
	}

	if d.Get("activate").(bool) {
		activateVersion(serviceID, versionNumber, meta)
	}

	return nil
}

func activateVersion(serviceID string, version int, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	_, err := conn.ActivateVersion(&gofastly.ActivateVersionInput{
		Service: serviceID,
		Version: version,
	})

	if err != nil {
		return err
	}

	return nil
}
