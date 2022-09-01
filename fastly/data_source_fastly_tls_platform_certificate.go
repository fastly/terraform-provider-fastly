package fastly

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyTLSPlatformCertificate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyTLSPlatformCertificateRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Description:   "Unique ID assigned to certificate by Fastly. Conflicts with all the other filters.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"domains"},
			},
			"domains": {
				Type:          schema.TypeSet,
				Description:   "Domains that are listed in any certificate's Subject Alternative Names (SAN) list.",
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"id"},
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was created.",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was last updated.",
				Computed:    true,
			},
			"not_before": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate will become valid.",
				Computed:    true,
			},
			"not_after": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate will expire.",
				Computed:    true,
			},
			"replace": {
				Type:        schema.TypeBool,
				Description: "A recommendation from Fastly indicating the key associated with this certificate is in need of rotation.",
				Computed:    true,
			},
			"configuration_id": {
				Type:        schema.TypeString,
				Description: "ID of TLS configuration used to terminate TLS traffic.",
				Computed:    true,
			},
		},
	}
}

func dataSourceFastlyTLSPlatformCertificateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	var diags diag.Diagnostics

	var certificate *fastly.BulkCertificate

	if v, ok := d.GetOk("id"); ok {
		cert, err := conn.GetBulkCertificate(&fastly.GetBulkCertificateInput{
			ID: v.(string),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		certificate = cert
	} else {
		filters := getPlatformTLSCertificateFilters(d)

		certificates, err := listPlatformTLSCertificates(conn, filters...)
		if err != nil {
			return diag.FromErr(err)
		}

		if len(certificates) == 0 {
			return diag.Errorf("Your query returned no results. Please change your search criteria and try again.")
		}

		if len(certificates) > 1 {
			return diag.Errorf("Your query returned more than one result. Please change try a more specific search criteria and try again.")
		}

		certificate = certificates[0]
	}

	if certificate.Replace {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Fastly recommends that this certificate (%s) be replaced", certificate.ID),
		})
	}

	err := dataSourceFastlyTLSPlatformCertificateSetAttributes(certificate, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// PlatformTLSCertificatePredicate determines if a certificate should be filtered.
type PlatformTLSCertificatePredicate func(certificate *fastly.BulkCertificate) bool

func getPlatformTLSCertificateFilters(d *schema.ResourceData) []PlatformTLSCertificatePredicate {
	var filters []PlatformTLSCertificatePredicate

	if v, ok := d.GetOk("domains"); ok {
		filters = append(filters, func(c *fastly.BulkCertificate) bool {
			s := v.(*schema.Set)
			for _, domain := range c.Domains {
				if s.Contains(domain.ID) {
					return true
				}
			}
			return false
		})
	}

	return filters
}

func listPlatformTLSCertificates(conn *fastly.Client, filters ...PlatformTLSCertificatePredicate) ([]*fastly.BulkCertificate, error) {
	var certificates []*fastly.BulkCertificate
	pageNumber := 1
	for {
		list, err := conn.ListBulkCertificates(&fastly.ListBulkCertificatesInput{
			PageNumber: pageNumber,
			PageSize:   10,
		})
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			break
		}
		pageNumber++

		for _, certificate := range list {
			if filterPlatformTLSCertificate(certificate, filters) {
				certificates = append(certificates, certificate)
			}
		}
	}

	return certificates, nil
}

func dataSourceFastlyTLSPlatformCertificateSetAttributes(certificate *fastly.BulkCertificate, d *schema.ResourceData) error {
	var domains []string
	for _, domain := range certificate.Domains {
		domains = append(domains, domain.ID)
	}

	d.SetId(certificate.ID)
	if err := d.Set("created_at", certificate.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", certificate.UpdatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("not_before", certificate.NotBefore.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("not_after", certificate.NotAfter.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("replace", certificate.Replace); err != nil {
		return err
	}
	if err := d.Set("domains", domains); err != nil {
		return err
	}
	if err := d.Set("configuration_id", certificate.Configurations[0].ID); err != nil {
		return err
	}

	return nil
}

func filterPlatformTLSCertificate(config *fastly.BulkCertificate, filters []PlatformTLSCertificatePredicate) bool {
	for _, f := range filters {
		if !f(config) {
			return false
		}
	}
	return true
}
