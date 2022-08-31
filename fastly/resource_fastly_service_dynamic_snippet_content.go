package fastly

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServiceDynamicSnippetContent() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceDynamicSnippetCreate,
		ReadContext:   resourceServiceDynamicSnippetRead,
		UpdateContext: resourceServiceDynamicSnippetUpdate,
		DeleteContext: resourceServiceDynamicSnippetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceDynamicSnippetContentImport,
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the service that the dynamic snippet belongs to",
			},
			"snippet_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the dynamic snippet that the content belong to",
			},
			"manage_snippets": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Whether to reapply changes if the state of the snippets drifts, i.e. if snippets are managed externally",
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VCL code that specifies exactly what the snippet does",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.HasChange("snippet_id") && !d.Get("manage_snippets").(bool)
				},
			},
		},
	}
}

func resourceServiceDynamicSnippetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	snippetID := d.Get("snippet_id").(string)
	content := d.Get("content").(string)

	_, err := conn.UpdateDynamicSnippet(&gofastly.UpdateDynamicSnippetInput{
		ServiceID: serviceID,
		ID:        snippetID,
		Content:   gofastly.String(content),
	})

	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 409 {
			return diag.FromErr(err)
		}
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, snippetID))
	return resourceServiceDynamicSnippetRead(ctx, d, meta)
}

func resourceServiceDynamicSnippetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	snippetID := d.Get("snippet_id").(string)

	if d.HasChange("content") {

		content := d.Get("content").(string)

		_, err := conn.UpdateDynamicSnippet(&gofastly.UpdateDynamicSnippetInput{
			ServiceID: serviceID,
			ID:        snippetID,
			Content:   gofastly.String(content),
		})
		if err != nil {
			return diag.Errorf("Error updating dynamic snippet: service %s, snippet %s, %#v", serviceID, snippetID, err)
		}
	}

	return resourceServiceDynamicSnippetRead(ctx, d, meta)
}

func resourceServiceDynamicSnippetRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	snippetID := d.Get("snippet_id").(string)

	dynamicSnippet, err := conn.GetDynamicSnippet(&gofastly.GetDynamicSnippetInput{
		ServiceID: serviceID,
		ID:        snippetID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("content", dynamicSnippet.Content)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceDynamicSnippetDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Dynamic snippet content cannot be deleted. Removing from state only
	d.SetId("")
	return nil
}

func resourceServiceDynamicSnippetContentImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")

	if len(split) != 2 {
		return nil, fmt.Errorf("invalid id: %s. The ID should be in the format [service_id]/[snippet_id]", d.Id())
	}

	serviceID := split[0]
	snippetID := split[1]

	err := d.Set("service_id", serviceID)
	if err != nil {
		return nil, fmt.Errorf("error importing dynamic snippet content: service %s, dynamic snippet %s, %s", serviceID, snippetID, err)
	}

	err = d.Set("snippet_id", snippetID)
	if err != nil {
		return nil, fmt.Errorf("error importing dynamic snippet content: service %s, dynamic snippet %s, %s", serviceID, snippetID, err)
	}

	return []*schema.ResourceData{d}, nil
}
