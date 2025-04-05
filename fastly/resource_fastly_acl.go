package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyACLCreate,
		ReadContext:   resourceFastlyACLRead,
		UpdateContext: resourceFastlyACLUpdate,
		DeleteContext: resourceFastlyACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"acl_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the ACL",
			},
			"force_destroy": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Allow the ACL to be deleted, even if it contains entries. Defaults to false.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name to identify the ACL. It is important to note that changing this attribute will delete and recreate the ACL, and discard the current entries.",
				ForceNew:    true,
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the service that the ACL belongs to",
				ForceNew:    true,
			},
		},
	}
}

func resourceFastlyACLCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)

	// Get the latest active service version
	latestVersion, err := getLatestServiceVersion(serviceID, conn)
	if err != nil {
		return diag.FromErr(err)
	}

	input := &gofastly.CreateACLInput{
		ServiceID:      serviceID,
		ServiceVersion: latestVersion,
		Name:           gofastly.ToPointer(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] CREATE: ACL input: %#v", input)

	acl, err := conn.CreateACL(input)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the ACL ID
	if acl.ACLID != nil {
		d.Set("acl_id", acl.ACLID)
	}

	// Set the resource ID to a composite of service_id/acl_id
	d.SetId(fmt.Sprintf("%s/%s", serviceID, *acl.ACLID))

	return nil
}

func resourceFastlyACLRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	// Extract service_id and acl_id from the composite ID
	serviceID, aclID, err := parseACLID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Set service_id in the state if it's not already set
	if d.Get("service_id").(string) == "" {
		d.Set("service_id", serviceID)
	}

	// Get the latest active service version
	latestVersion, err := getLatestServiceVersion(serviceID, conn)
	if err != nil {
		return diag.FromErr(err)
	}

	// Try to find the ACL by name in the latest version
	acls, err := conn.ListACLs(&gofastly.ListACLsInput{
		ServiceID:      serviceID,
		ServiceVersion: latestVersion,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var acl *gofastly.ACL
	for _, a := range acls {
		if a.ACLID != nil && *a.ACLID == aclID {
			acl = a
			break
		}
	}

	if acl == nil {
		log.Printf("[WARN] No ACL found with ID '%s'", aclID)
		d.SetId("")
		return nil
	}

	d.Set("acl_id", aclID)

	if acl.Name != nil {
		d.Set("name", acl.Name)
	}

	return nil
}

// UPDATE operation is not supported for ACLs - any change in the name will result in a delete and recreate
func resourceFastlyACLUpdate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func resourceFastlyACLDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID, aclID, err := parseACLID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if !d.Get("force_destroy").(bool) {
		mayDelete, err := isACLEmpty(serviceID, aclID, conn)
		if err != nil {
			return diag.FromErr(err)
		}

		if !mayDelete {
			return diag.FromErr(fmt.Errorf("cannot delete ACL (%s), list is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", aclID))
		}
	} else {
		// If force_destroy is true, delete all entries first
		entries, err := getAllAclEntriesViaPaginator(conn, &gofastly.GetACLEntriesInput{
			ServiceID: serviceID,
			ACLID:     aclID,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing ACL entries during cleanup: %w", err))
		}

		// Delete entries in batches
		var batchEntries []*gofastly.BatchACLEntry
		for _, entry := range entries {
			batchEntries = append(batchEntries, &gofastly.BatchACLEntry{
				Operation: gofastly.ToPointer(gofastly.DeleteBatchOperation),
				EntryID:   entry.EntryID,
			})
		}

		if len(batchEntries) > 0 {
			err = executeBatchACLOperations(conn, serviceID, aclID, batchEntries)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error deleting ACL entries during cleanup: %w", err))
			}
		}
	}

	// Get the latest active service version
	latestVersion, err := getLatestServiceVersion(serviceID, conn)
	if err != nil {
		return diag.FromErr(err)
	}

	// Find the ACL by ID to get its name
	acls, err := conn.ListACLs(&gofastly.ListACLsInput{
		ServiceID:      serviceID,
		ServiceVersion: latestVersion,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var aclName string
	for _, a := range acls {
		if a.ACLID != nil && *a.ACLID == aclID {
			aclName = *a.Name
			break
		}
	}

	if aclName == "" {
		// ACL not found, it might have been deleted already
		d.SetId("")
		return nil
	}

	input := gofastly.DeleteACLInput{
		ServiceID:      serviceID,
		ServiceVersion: latestVersion,
		Name:           aclName,
	}

	log.Printf("[DEBUG] DELETE: ACL input: %#v", input)

	err = conn.DeleteACL(&input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.StatusCode == 404 {
			// ACL was already deleted
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Helper function to get the latest active service version
func getLatestServiceVersion(serviceID string, conn *gofastly.Client) (int, error) {
	service, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ServiceID: serviceID,
	})
	if err != nil {
		return 0, err
	}

	// Find the latest active version
	var latestVersion int
	for _, version := range service.Versions {
		if version.Active != nil && *version.Active && version.Number != nil {
			if *version.Number > latestVersion {
				latestVersion = *version.Number
			}
		}
	}

	if latestVersion == 0 {
		// If no active version, use the latest version
		for _, version := range service.Versions {
			if version.Number != nil && *version.Number > latestVersion {
				latestVersion = *version.Number
			}
		}
	}

	if latestVersion == 0 {
		return 0, fmt.Errorf("no active version found for service %s", serviceID)
	}

	return latestVersion, nil
}

// Helper function to parse the composite ID (service_id/acl_id)
func parseACLID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid ACL ID format. Expected format: service_id/acl_id")
	}
	return parts[0], parts[1], nil
}
