package fastly

import (
	"fmt"
	"time"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSCertificateRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Description:   "Unique ID assigned to certificate by Fastly",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name", "issued_to", "domains", "issuer"},
			},
			"name": {
				Type:          schema.TypeString,
				Description:   "Human-readable name used to identify the certificate. Defaults to the certificate's Common Name or first Subject Alternative Name entry.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
			},
			"issued_to": {
				Type:          schema.TypeString,
				Description:   "The hostname for which a certificate was issued.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
			},
			"domains": {
				Type:          schema.TypeSet,
				Description:   "Domains that are listed in any certificates' Subject Alternative Names (SAN) list.",
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"id"},
			},
			"issuer": {
				Type:          schema.TypeString,
				Description:   "The certificate authority that issued the certificate.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was created",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the certificate was last updated",
				Computed:    true,
			},
			"replace": {
				Type:        schema.TypeBool,
				Description: "A recommendation from Fastly indicating the key associated with this certificate is in need of rotation",
				Computed:    true,
			},
			"serial_number": {
				Type:        schema.TypeString,
				Description: "A value assigned by the issuer that is unique to a certificate",
				Computed:    true,
			},
			"signature_algorithm": {
				Type:        schema.TypeString,
				Description: "The algorithm used to sign the certificate",
				Computed:    true,
			},
		},
	}
}

func dataSourceFastlyTLSCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var certificate *fastly.CustomTLSCertificate

	if v, ok := d.GetOk("id"); ok {
		cert, err := conn.GetCustomTLSCertificate(&fastly.GetCustomTLSCertificateInput{
			ID: v.(string),
		})
		if err != nil {
			return err
		}

		certificate = cert
	} else {
		filters := getTLSCertificateFilters(d)

		certificates, err := listTLSCertificates(conn, filters...)
		if err != nil {
			return err
		}

		if len(certificates) == 0 {
			return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
		}

		if len(certificates) > 1 {
			return fmt.Errorf("Your query returned more than one result. Please change try a more specific search criteria and try again.")
		}

		certificate = certificates[0]
	}

	return dataSourceFastlyTLSCertificateSetAttributes(certificate, d)
}

type TLSCertificatePredicate func(*fastly.CustomTLSCertificate) bool

func getTLSCertificateFilters(d *schema.ResourceData) []TLSCertificatePredicate {
	var filters []TLSCertificatePredicate

	if v, ok := d.GetOk("name"); ok {
		filters = append(filters, func(c *fastly.CustomTLSCertificate) bool {
			return c.Name == v.(string)
		})
	}
	if v, ok := d.GetOk("issued_to"); ok {
		filters = append(filters, func(c *fastly.CustomTLSCertificate) bool {
			return c.IssuedTo == v.(string)
		})
	}
	if v, ok := d.GetOk("domains"); ok {
		filters = append(filters, func(c *fastly.CustomTLSCertificate) bool {
			s := v.(*schema.Set)
			for _, domain := range c.TLSDomains {
				if s.Contains(domain.ID) {
					return true
				}
			}
			return false
		})
	}
	if v, ok := d.GetOk("issuer"); ok {
		filters = append(filters, func(c *fastly.CustomTLSCertificate) bool {
			return c.Issuer == v.(string)
		})
	}

	return filters
}

func listTLSCertificates(conn *fastly.Client, filters ...TLSCertificatePredicate) ([]*fastly.CustomTLSCertificate, error) {
	var certificates []*fastly.CustomTLSCertificate
	pageNumber := 1
	for {
		list, err := conn.ListCustomTLSCertificates(&fastly.ListCustomTLSCertificatesInput{
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
			if filterTLSCertificate(certificate, filters) {
				certificates = append(certificates, certificate)
			}
		}
	}

	return certificates, nil
}

func dataSourceFastlyTLSCertificateSetAttributes(certificate *fastly.CustomTLSCertificate, d *schema.ResourceData) error {
	var domains []string
	for _, domain := range certificate.TLSDomains {
		domains = append(domains, domain.ID)
	}

	d.SetId(certificate.ID)
	if err := d.Set("name", certificate.Name); err != nil {
		return err
	}
	if err := d.Set("created_at", certificate.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", certificate.UpdatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("issued_to", certificate.IssuedTo); err != nil {
		return err
	}
	if err := d.Set("issuer", certificate.Issuer); err != nil {
		return err
	}
	if err := d.Set("replace", certificate.Replace); err != nil {
		return err
	}
	if err := d.Set("serial_number", certificate.SerialNumber); err != nil {
		return err
	}
	if err := d.Set("signature_algorithm", certificate.SignatureAlgorithm); err != nil {
		return err
	}
	if err := d.Set("domains", domains); err != nil {
		return err
	}

	return nil
}

func filterTLSCertificate(config *fastly.CustomTLSCertificate, filters []TLSCertificatePredicate) bool {
	for _, f := range filters {
		if !f(config) {
			return false
		}
	}
	return true
}
