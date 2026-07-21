package fastly

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v16/fastly"
)

func resourceFastlyTLSSubscriptionValidation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyTLSSubscriptionValidationCreate,
		ReadContext:   resourceFastlyTLSSubscriptionValidationRead,
		DeleteContext: resourceFastlyTLSSubscriptionValidationDelete,
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the certificate issued for the validated subscription. Only populated once the subscription reaches the `issued` state. Reference this from `fastly_tls_activation.certificate_id` to guarantee the activation is created after the certificate exists, within a single apply.",
			},
			"subscription_id": {
				Type:        schema.TypeString,
				Description: "The ID of the TLS Subscription that should be validated.",
				Required:    true,
				ForceNew:    true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
		},
	}
}

const (
	subscriptionStateIssued = "issued"
)

func resourceFastlyTLSSubscriptionValidationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		subscription, err := conn.GetTLSSubscription(ctx, &gofastly.GetTLSSubscriptionInput{
			ID: d.Get("subscription_id").(string),
		})
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if subscription.State != subscriptionStateIssued {
			return retry.RetryableError(fmt.Errorf("expected subscription state to be %s but it was %s", subscriptionStateIssued, subscription.State))
		}

		err = diagToErr(resourceFastlyTLSSubscriptionValidationRead(ctx, d, meta))
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyTLSSubscriptionValidationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing TLS Subscription Validation Configuration for (%s)", d.Id())

	conn := meta.(*APIClient).conn

	subscriptionID := d.Get("subscription_id").(string)
	subscription, err := conn.GetTLSSubscription(ctx, &gofastly.GetTLSSubscriptionInput{
		ID: subscriptionID,
	})
	if err, ok := err.(*gofastly.HTTPError); ok && err.IsNotFound() {
		id := d.Id()
		d.SetId("")
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("TLS subscription (%s) not found - removing from state", id),
				AttributePath: cty.Path{cty.GetAttrStep{Name: id}},
			},
		}
	} else if err != nil {
		return diag.FromErr(err)
	}

	// NOTE: there must be only one certificate id included per subscription.
	// A subscription with a certificate may be in the "issued" or "renewing"
	// state; both are valid, so validity is keyed off the certificate's
	// presence rather than the state (avoids destroy/recreate churn during
	// renewals).
	certificateID := ""
	if len(subscription.Certificates) > 0 {
		certificateID = subscription.Certificates[0].ID
	}

	if certificateID == "" {
		d.SetId("")
	} else {
		d.SetId(subscriptionID)
		if err := d.Set("certificate_id", certificateID); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceFastlyTLSSubscriptionValidationDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	// Virtual resource so doesn't need deleting
	return nil
}
