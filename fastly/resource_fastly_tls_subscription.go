package fastly

import (
	"fmt"
	"time"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceFastlyTLSSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceFastlyTLSSubscriptionCreate,
		Read:   resourceFastlyTLSSubscriptionRead,
		Delete: resourceFastlyTLSSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough, // FIXME: might want to do something with adding a "subscription_validation" resource too
		},
		Schema: map[string]*schema.Schema{
			"domains": {
				Type:        schema.TypeSet,
				Description: "List of domains on which to enable TLS.",
				Required:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Set:         schema.HashString,
			},
			"certificate_authority": {
				Type:         schema.TypeString,
				Description:  "The entity that issues and certifies the TLS certificates for your subscription. Valid values are `lets-encrypt` or `globalsign`.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"lets-encrypt", "globalsign"}, false),
			},
			"configuration_id": {
				Type:        schema.TypeString,
				Description: "The ID of the set of TLS configuration options that apply to the enabled domains on this subscription.",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the subscription was created.",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the subscription was updated.",
				Computed:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "The current state of the subscription. The list of possible states are: `pending`, `processing`, `issued`, and `renewing`.",
				Computed:    true,
			},
			"tls_authorization_challenges": {
				Type:        schema.TypeSet,
				Description: "Data required to set up DNS to respond to domain ownership challenge.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"challenge_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_values": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
				Set: authorisationChallengesHash,
			},
		},
	}
}

func resourceFastlyTLSSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var configuration *fastly.TLSConfiguration
	if v, ok := d.GetOk("configuration_id"); ok {
		configuration = &fastly.TLSConfiguration{ID: v.(string)}
	}

	var domains []*fastly.TLSDomain
	for _, domain := range d.Get("domains").(*schema.Set).List() {
		domains = append(domains, &fastly.TLSDomain{ID: domain.(string)})
	}

	subscription, err := conn.CreateTLSSubscription(&fastly.CreateTLSSubscriptionInput{
		CertificateAuthority: d.Get("certificate_authority").(string),
		Configuration:        configuration,
		Domains:              domains,
	})
	if err != nil {
		return err
	}

	d.SetId(subscription.ID)

	return resourceFastlyTLSSubscriptionRead(d, meta)
}

func resourceFastlyTLSSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	include := "tls_authorizations"
	subscription, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
		ID:      d.Id(),
		Include: &include,
	})
	if err != nil {
		return err
	}

	var domains []string
	for _, domain := range subscription.Domains {
		domains = append(domains, domain.ID)
	}

	var authorisationChallenges []map[string]interface{}
	for _, challenge := range subscription.Authorizations[0].Challenges {
		authorisationChallenges = append(authorisationChallenges, map[string]interface{}{
			"challenge_type": challenge.Type,
			"record_type":    challenge.RecordType,
			"record_name":    challenge.RecordName,
			"record_values":  challenge.Values,
		})
	}

	err = d.Set("domains", domains)
	if err != nil {
		return err
	}
	err = d.Set("certificate_authority", subscription.CertificateAuthority)
	if err != nil {
		return err
	}
	err = d.Set("configuration_id", subscription.Configuration.ID)
	if err != nil {
		return err
	}
	err = d.Set("created_at", subscription.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	err = d.Set("updated_at", subscription.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	err = d.Set("state", subscription.State)
	if err != nil {
		return err
	}
	err = d.Set("tls_authorization_challenges", authorisationChallenges)
	if err != nil {
		return err
	}

	return nil
}

func resourceFastlyTLSSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	// Delete all activations on TLS Domains in the Subscription
	for _, domain := range d.Get("domains").(*schema.Set).List() {
		activations, err := conn.ListTLSActivations(&fastly.ListTLSActivationsInput{
			FilterTLSDomainID: domain.(string),
		})
		if err != nil {
			return err
		}

		for _, activation := range activations {
			if activation.Domain.ID != domain.(string) {
				return fmt.Errorf("Fastly API returned too many TLS activations for this domain (%s)", domain)
			}

			err = conn.DeleteTLSActivation(&fastly.DeleteTLSActivationInput{ID: activation.ID})
			if err != nil {
				return err
			}
		}
	}

	err := conn.DeleteTLSSubscription(&fastly.DeleteTLSSubscriptionInput{
		ID: d.Id(),
	})
	return err
}

func authorisationChallengesHash(value interface{}) int {
	m, ok := value.(map[string]interface{})
	if !ok {
		return 0
	}

	challengeType, ok := m["challenge_type"].(string)
	if ok {
		return hashcode.String(challengeType)
	}

	return 0
}
