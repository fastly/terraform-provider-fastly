package fastly

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v11/fastly"
)

func resourceFastlyTLSActivation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyTLSActivationCreate,
		ReadContext:   resourceFastlyTLSActivationRead,
		UpdateContext: resourceFastlyTLSActivationUpdate,
		DeleteContext: resourceFastlyTLSActivationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of certificate to use. Must have the `domain` specified in the certificate's Subject Alternative Names.",
			},
			"configuration_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "ID of TLS configuration to be used to terminate TLS traffic, or use the default one if missing.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time-stamp (GMT) when TLS was enabled.",
			},
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Domain to enable TLS on. Must be assigned to an existing Fastly Service.",
			},
			"mutual_authentication_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "An alphanumeric string identifying a mutual authentication.",
			},
		},
	}
}

func resourceFastlyTLSActivationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	var diags diag.Diagnostics

	var configuration *fastly.TLSConfiguration
	if v, ok := d.GetOk("configuration_id"); ok {
		configuration = &fastly.TLSConfiguration{ID: v.(string)}
	}

	activation, err := conn.CreateTLSActivation(ctx, &fastly.CreateTLSActivationInput{
		Certificate:   &fastly.CustomTLSCertificate{ID: d.Get("certificate_id").(string)},
		Configuration: configuration,
		Domain:        &fastly.TLSDomain{ID: d.Get("domain").(string)},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(activation.ID)

	// Setting the mutual_authentication_id is only possible through an update via PATCH.
	// See https://github.com/fastly/terraform-provider-fastly/issues/873
	mtlsID := d.Get("mutual_authentication_id").(string)
	if mtlsID != "" {
		diags = append(diags, resourceFastlyTLSActivationUpdate(ctx, d, meta)...)
	}

	return append(diags, resourceFastlyTLSActivationRead(ctx, d, meta)...)
}

func resourceFastlyTLSActivationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing TLS Activation Configuration for (%s)", d.Id())

	conn := meta.(*APIClient).conn

	activation, err := conn.GetTLSActivation(ctx, &fastly.GetTLSActivationInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if activation == nil {
		return diag.FromErr(errors.New("unexpected nil value for TLSActivation"))
	}

	if activation.Certificate != nil {
		err = d.Set("certificate_id", activation.Certificate.ID)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if activation.Configuration != nil {
		err = d.Set("configuration_id", activation.Configuration.ID)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if activation.Domain != nil {
		err = d.Set("domain", activation.Domain.ID)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if activation.CreatedAt != nil {
		err = d.Set("created_at", activation.CreatedAt.Format(time.RFC3339))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if activation.MutualAuthentication != nil {
		err = d.Set("mutual_authentication_id", activation.MutualAuthentication.ID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceFastlyTLSActivationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &fastly.UpdateTLSActivationInput{
		ID:          d.Id(),
		Certificate: &fastly.CustomTLSCertificate{ID: d.Get("certificate_id").(string)},
	}

	mtlsID := d.Get("mutual_authentication_id").(string)
	if mtlsID != "" {
		input.MutualAuthentication = &fastly.TLSMutualAuthentication{
			ID: mtlsID,
		}
	}

	_, err := conn.UpdateTLSActivation(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyTLSActivationRead(ctx, d, meta)
}

func resourceFastlyTLSActivationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteTLSActivation(ctx, &fastly.DeleteTLSActivationInput{
		ID: d.Id(),
	})
	if err != nil {
		if httpErr, ok := err.(*fastly.HTTPError); ok && httpErr.IsNotFound() {
			log.Printf("[WARN] Error deleting TLS activation (%s), not found. Was a TLS subscription enabled on the same domain?\n", d.Id())
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
