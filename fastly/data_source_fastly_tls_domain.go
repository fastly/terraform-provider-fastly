package fastly

import (
	"fmt"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyTLSDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSDomainsRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "Domain name to look up activations, certificates and subscriptions for.",
				Required:    true,
			},
			"tls_activation_ids": {
				Type:        schema.TypeSet,
				Description: "IDs of the activations associated with the domain.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tls_certificate_ids": {
				Type:        schema.TypeSet,
				Description: "IDs of the certificates associated with the domain.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tls_subscription_ids": {
				Type:        schema.TypeSet,
				Description: "IDs of the subscriptions associated with the domain.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFastlyTLSDomainsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	domain, err := findTLSDomain(conn, d)
	if err != nil {
		return err
	}

	var activations []string
	for _, activation := range domain.Activations {
		activations = append(activations, activation.ID)
	}
	var certificates []string
	for _, certificate := range domain.Certificates {
		certificates = append(certificates, certificate.ID)
	}
	var subscriptions []string
	for _, subscription := range domain.Subscriptions {
		subscriptions = append(subscriptions, subscription.ID)
	}

	d.SetId(domain.ID)
	if err := d.Set("tls_activation_ids", activations); err != nil {
		return err
	}
	if err := d.Set("tls_certificate_ids", certificates); err != nil {
		return err
	}
	if err := d.Set("tls_subscription_ids", subscriptions); err != nil {
		return err
	}

	return nil
}

func findTLSDomain(conn *fastly.Client, d *schema.ResourceData) (*fastly.TLSDomain, error) {
	domain := d.Get("domain").(string)
	filter := func(d *fastly.TLSDomain) bool {
		return d.ID == domain
	}

	domains, err := listTLSDomains(conn, filter)
	if err != nil {
		return nil, err
	}

	if len(domains) == 0 {
		return nil, fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(domains) > 1 {
		return nil, fmt.Errorf("Your query returned more than one result. Please change to a more specific search criteria and try again.")
	}

	return domains[0], nil
}

type TLSDomainPredicate func(domain *fastly.TLSDomain) bool

func listTLSDomains(conn *fastly.Client, filters ...TLSDomainPredicate) ([]*fastly.TLSDomain, error) {
	var domains []*fastly.TLSDomain
	pageNumber := 1

	for {
		list, err := conn.ListTLSDomains(&fastly.ListTLSDomainsInput{
			PageNumber: pageNumber,
		})
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			break
		}
		pageNumber++

		for _, domain := range list {
			if filterTLSDomain(domain, filters) {
				domains = append(domains, domain)
			}
		}
	}
	return domains, nil
}

func filterTLSDomain(domain *fastly.TLSDomain, filters []TLSDomainPredicate) bool {
	for _, f := range filters {
		if !f(domain) {
			return false
		}
	}
	return true
}
