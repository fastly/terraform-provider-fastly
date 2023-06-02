package fastly

import (
	"context"
	"log"
	"sort"

	"github.com/fastly/go-fastly/v8/fastly"
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
				Description: "Determines whether Mutual TLS will fail closed (enforced) or fail open. A true value will require a successful Mutual TLS handshake for the connection to continue and will fail closed if unsuccessful. A false value will fail open and allow the connection to proceed.",
				Optional:    true,
				Computed:    true,
			},
			"include": {
				Type:        schema.TypeString,
				Description: "Comma-separated list of related objects to include (e.g. `tls_activations` will provide you with the TLS domain names that are related to your Mutual TLS authentication).",
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

func resourceFastlyTLSMutualAuthenticationCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &fastly.CreateTLSMutualAuthenticationInput{
		CertBundle: d.Get("cert_bundle").(string),
	}

	if v, ok := d.GetOk("enforced"); ok {
		input.Enforced = v.(bool)
	}
	if v, ok := d.GetOk("name"); ok {
		input.Name = v.(string)
	}

	log.Printf("[DEBUG] CREATE: TLS Mutual Authentication input: %#v", input)

	output, err := conn.CreateTLSMutualAuthentication(input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output.ID)

	return nil
}

func resourceFastlyTLSMutualAuthenticationRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &fastly.GetTLSMutualAuthenticationInput{
		ID: d.Id(),
	}

	if v, ok := d.GetOk("include"); ok {
		input.Include = v.(string)
	}

	log.Printf("[DEBUG] REFRESH: TLS Mutual Authentication input: %#v", input)

	tma, err := conn.GetTLSMutualAuthentication(input)
	if err != nil {
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
	if len(tma.Activations) > 0 {
		for _, a := range tma.Activations {
			activations = append(activations, a.ID)
		}
		sort.Strings(activations)

		err := d.Set("tls_activations", activations)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceFastlyTLSMutualAuthenticationUpdate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &fastly.UpdateTLSMutualAuthenticationInput{
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
	return nil
}

func resourceFastlyTLSMutualAuthenticationDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &fastly.DeleteTLSMutualAuthenticationInput{
		ID: d.Id(),
	}

	log.Printf("[DEBUG] DELETE: TLS Mutual Authentication input: %#v", input)

	err := conn.DeleteTLSMutualAuthentication(input)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
