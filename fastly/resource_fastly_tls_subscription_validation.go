package fastly

import (
	"context"
	"fmt"
	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceFastlyTLSSubscriptionValidation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyTLSSubscriptionValidationCreate,
		ReadContext:   resourceFastlyTLSSubscriptionValidationRead,
		DeleteContext: resourceFastlyTLSSubscriptionValidationDelete,
		Schema: map[string]*schema.Schema{
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

func resourceFastlyTLSSubscriptionValidationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		subscription, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
			ID: d.Get("subscription_id").(string),
		})
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if subscription.State != subscriptionStateIssued {
			return resource.RetryableError(fmt.Errorf("Expected subscription state to be %s but it was %s", subscriptionStateIssued, subscription.State))
		}

		err = resourceFastlyTLSSubscriptionValidationRead(d, meta)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
}

func resourceFastlyTLSSubscriptionValidationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	subscriptionID := d.Get("subscription_id").(string)
	subscription, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
		ID: subscriptionID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if subscription.State != subscriptionStateIssued {
		d.SetId("")
	} else {
		d.SetId(subscriptionID)
	}

	return nil
}

func resourceFastlyTLSSubscriptionValidationDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Virtual resource so doesn't need deleting
	return nil
}
