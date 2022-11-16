package fastly

import (
	"context"
	"fmt"
	"log"
	"time"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func resourceFastlyTLSSubscriptionValidationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		subscription, err := conn.GetTLSSubscription(&gofastly.GetTLSSubscriptionInput{
			ID: d.Get("subscription_id").(string),
		})
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if subscription.State != subscriptionStateIssued {
			return resource.RetryableError(fmt.Errorf("expected subscription state to be %s but it was %s", subscriptionStateIssued, subscription.State))
		}

		err = diagToErr(resourceFastlyTLSSubscriptionValidationRead(ctx, d, meta))
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyTLSSubscriptionValidationRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing TLS Subscription Validation Configuration for (%s)", d.Id())

	conn := meta.(*APIClient).conn

	subscriptionID := d.Get("subscription_id").(string)
	subscription, err := conn.GetTLSSubscription(&gofastly.GetTLSSubscriptionInput{
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

	if subscription.State != subscriptionStateIssued {
		d.SetId("")
	} else {
		d.SetId(subscriptionID)
	}

	return nil
}

func resourceFastlyTLSSubscriptionValidationDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	// Virtual resource so doesn't need deleting
	return nil
}
