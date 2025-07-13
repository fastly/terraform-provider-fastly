package fastly

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

func resourceFastlyTLSMutualAuthentication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyTLSMutualAuthenticationCreate,
		ReadContext:   resourceFastlyTLSMutualAuthenticationRead,
		UpdateContext: resourceFastlyTLSMutualAuthenticationUpdate,
		DeleteContext: resourceFastlyTLSMutualAuthenticationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"activation_ids": {
				Type:        schema.TypeSet,
				Description: "List of TLS Activation IDs",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cert_bundle": {
				Type:        schema.TypeString,
				Description: "One or more certificates. Enter each individual certificate blob on a new line. Must be PEM-formatted.",
				Required:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Date and time in ISO 8601 format.",
				Computed:    true,
			},
			"enforced": {
				Type:        schema.TypeBool,
				Description: "Determines whether Mutual TLS will fail closed (enforced) or fail open. A true value will require a successful Mutual TLS handshake for the connection to continue and will fail closed if unsuccessful. A false value will fail open and allow the connection to proceed (if this attribute is not set we default to `false`).",
				Optional:    true,
				Computed:    true,
			},
			"include": {
				Type:        schema.TypeString,
				Description: "A comma-separated list used by the Terraform provider during a state refresh to return more data related to your mutual authentication from the Fastly API (permitted values: `tls_activations`).",
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "A custom name for your mutual authentication. If name is not supplied we will auto-generate one.",
				Optional:    true,
				Computed:    true,
			},
			"tls_activations": {
				Type:        schema.TypeList,
				Description: "List of alphanumeric strings identifying TLS activations.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Date and time in ISO 8601 format.",
				Computed:    true,
			},
		},
	}
}

func resourceFastlyTLSMutualAuthenticationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	inputCreate := &gofastly.CreateTLSMutualAuthenticationInput{
		CertBundle: d.Get("cert_bundle").(string),
	}

	if v, ok := d.GetOk("enforced"); ok {
		inputCreate.Enforced = v.(bool)
	}
	if v, ok := d.GetOk("name"); ok {
		inputCreate.Name = v.(string)
	}

	log.Printf("[DEBUG] CREATE: TLS Mutual Authentication input: %#v", inputCreate)

	mTLS, err := conn.CreateTLSMutualAuthentication(ctx, inputCreate)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(mTLS.ID)

	if v, ok := d.GetOk("activation_ids"); ok {
		activationIDs := v.(*schema.Set).List()
		for _, id := range activationIDs {
			inputUpdate := &gofastly.UpdateTLSActivationInput{
				ID:                   id.(string),
				MutualAuthentication: &gofastly.TLSMutualAuthentication{ID: mTLS.ID},
			}
			log.Printf("[DEBUG] CREATE: Update TLS Activation input: %#v", inputUpdate)
			_, err = conn.UpdateTLSActivation(ctx, inputUpdate)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceFastlyTLSMutualAuthenticationRead(ctx, d, meta)
}

func resourceFastlyTLSMutualAuthenticationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.GetTLSMutualAuthenticationInput{
		ID: d.Id(),
	}

	if v, ok := d.GetOk("include"); ok {
		input.Include = v.(string)
	}

	log.Printf("[DEBUG] REFRESH: TLS Mutual Authentication input: %#v", input)

	tma, err := conn.GetTLSMutualAuthentication(ctx, input)
	if err != nil {
		if err, ok := err.(*gofastly.HTTPError); ok && err.IsNotFound() {
			id := d.Id()
			d.SetId("")
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       fmt.Sprintf("TLS Mutual Authentication (%s) not found - removing from state", id),
					AttributePath: cty.Path{cty.GetAttrStep{Name: id}},
				},
			}
		}
		return diag.FromErr(err)
	}

	if err := d.Set("created_at", tma.CreatedAt.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enforced", tma.Enforced); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", tma.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", tma.UpdatedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	var activations []string
	for _, a := range tma.Activations {
		activations = append(activations, a.ID)
	}
	sort.Strings(activations)

	err = d.Set("tls_activations", activations)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyTLSMutualAuthenticationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.UpdateTLSMutualAuthenticationInput{
		ID:         d.Id(),
		CertBundle: d.Get("cert_bundle").(string),
	}

	// Since a boolean value is not 'optional', the input struct
	// must always contain the expected value of the 'enforced'
	// setting, whether it was changed or not
	input.Enforced = d.Get("enforced").(bool)

	if d.HasChange("name") {
		input.Name = d.Get("name").(string)
	}

	log.Printf("[DEBUG] UPDATE: TLS Mutual Authentication input: %#v", input)

	_, err := conn.UpdateTLSMutualAuthentication(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("activation_ids") {
		// First unset mTLS from the old TLS Activations.
		old, _ := d.GetChange("activation_ids")
		for _, id := range old.(*schema.Set).List() {
			input := &gofastly.UpdateTLSActivationInput{
				ID:                   id.(string),
				MutualAuthentication: &gofastly.TLSMutualAuthentication{ID: ""},
			}
			log.Printf("[DEBUG] UPDATE: TLS Activation input: %#v", input)
			_, _ = conn.UpdateTLSActivation(ctx, input)
		}

		// Once old Activations have mTLS unset, set mTLS on the new Activations.
		for _, id := range d.Get("activation_ids").(*schema.Set).List() {
			inputUpdate := &gofastly.UpdateTLSActivationInput{
				ID:                   id.(string),
				MutualAuthentication: &gofastly.TLSMutualAuthentication{ID: d.Id()},
			}
			log.Printf("[DEBUG] UPDATE: Update TLS Activation input: %#v", inputUpdate)
			_, err = conn.UpdateTLSActivation(ctx, inputUpdate)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

func resourceFastlyTLSMutualAuthenticationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	// IMPORTANT: You can't delete mTLS with active domains.
	// You must first disable the active domains.
	// To do that, you can set "" for the mTLS ID on each TLS Activation.
	activationIDs := d.Get("activation_ids").(*schema.Set).List()
	for _, id := range activationIDs {
		input := &gofastly.UpdateTLSActivationInput{
			ID:                   id.(string),
			MutualAuthentication: &gofastly.TLSMutualAuthentication{ID: ""},
		}
		log.Printf("[DEBUG] DELETE: TLS Activation input: %#v", input)
		_, err := conn.UpdateTLSActivation(ctx, input)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	input := &gofastly.DeleteTLSMutualAuthenticationInput{
		ID: d.Id(),
	}

	log.Printf("[DEBUG] DELETE: TLS Mutual Authentication input: %#v", input)

	err := conn.DeleteTLSMutualAuthentication(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
