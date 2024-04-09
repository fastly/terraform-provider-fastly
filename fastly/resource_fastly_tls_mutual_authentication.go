package fastly

import (
	"context"
	"fmt"
	"log"
	"sort"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"activation_id": {
				Type:        schema.TypeString,
				Description: "The ID of your TLS Activation object",
				Required:    true,
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

// There are two steps to setting up mTLS:
//
// 1. POST  /tls/mutual_authentications
// 2. PATCH /tls/activations/tls_activation_id
//
// The fastly_tls_activation data source can be used to acquire the Activation
// ID.
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

	mTLS, err := conn.CreateTLSMutualAuthentication(inputCreate)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(mTLS.ID)

	inputUpdate := &gofastly.UpdateTLSActivationInput{
		ID:                   d.Get("activation_id").(string),
		MutualAuthentication: &gofastly.TLSMutualAuthentication{ID: mTLS.ID},
	}
	log.Printf("[DEBUG] CREATE: Update TLS Activation input: %#v", inputUpdate)
	_, err = conn.UpdateTLSActivation(inputUpdate)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyTLSMutualAuthenticationRead(ctx, d, meta)
}

func resourceFastlyTLSMutualAuthenticationRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.GetTLSMutualAuthenticationInput{
		ID: d.Id(),
	}

	if v, ok := d.GetOk("include"); ok {
		input.Include = v.(string)
	}

	log.Printf("[DEBUG] REFRESH: TLS Mutual Authentication input: %#v", input)

	tma, err := conn.GetTLSMutualAuthentication(input)
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

// There are two steps to setting up mTLS:
//
// 1. POST  /tls/mutual_authentications
// 2. PATCH /tls/activations/tls_activation_id
//
// Once mTLS is set up and the Activation object updated with the mTLS object
// ID, then for the resource's UPDATE operation we need to allow the user to
// change the Activation ID.
func resourceFastlyTLSMutualAuthenticationUpdate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.UpdateTLSMutualAuthenticationInput{
		ID:         d.Id(),
		CertBundle: d.Get("cert_bundle").(string),
	}

	if d.HasChange("enforced") {
		input.Enforced = d.Get("enforced").(bool)
	}
	if d.HasChange("name") {
		input.Name = d.Get("name").(string)
	}

	log.Printf("[DEBUG] UPDATE: TLS Mutual Authentication input: %#v", input)

	_, err := conn.UpdateTLSMutualAuthentication(input)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("activation_id") {
		inputUpdate := &gofastly.UpdateTLSActivationInput{
			ID:                   d.Get("activation_id").(string),
			MutualAuthentication: &gofastly.TLSMutualAuthentication{ID: d.Id()},
		}
		log.Printf("[DEBUG] UPDATE: TLS Activation input: %#v", inputUpdate)
		_, err = conn.UpdateTLSActivation(inputUpdate)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceFastlyTLSMutualAuthenticationDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.DeleteTLSMutualAuthenticationInput{
		ID: d.Id(),
	}

	log.Printf("[DEBUG] DELETE: TLS Mutual Authentication input: %#v", input)

	err := conn.DeleteTLSMutualAuthentication(input)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
