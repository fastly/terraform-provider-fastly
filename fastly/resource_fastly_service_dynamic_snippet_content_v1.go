package fastly

import (
	"fmt"
	"strings"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceServiceDynamicSnippetContentV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceDynamicSnippetV1Create,
		Read:   resourceServiceDynamicSnippetV1Read,
		Update: resourceServiceDynamicSnippetV1Update,
		Delete: resourceServiceDynamicSnippetV1Delete,
		Importer: &schema.ResourceImporter{
			State: resourceServiceDynamicSnippetContentV1Import,
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service Id",
			},
			"snippet_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Snippet Id",
			},

			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The contents of the VCL dynamic snippet",
			},
		},
	}
}

func resourceServiceDynamicSnippetV1Create(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	snippetID := d.Get("snippet_id").(string)
	content := d.Get("content").(string)

	_, err := conn.UpdateDynamicSnippet(&gofastly.UpdateDynamicSnippetInput{
		Service: serviceID,
		ID:      snippetID,
		Content: content,
	})

	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 409 {
			return err
		}
	} else if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, snippetID))
	return resourceServiceDynamicSnippetV1Read(d, meta)
}

func resourceServiceDynamicSnippetV1Update(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	snippetID := d.Get("snippet_id").(string)

	if d.HasChange("content") {

		content := d.Get("content").(string)

		_, err := conn.UpdateDynamicSnippet(&gofastly.UpdateDynamicSnippetInput{
			Service: serviceID,
			ID:      snippetID,
			Content: content,
		})

		if err != nil {
			return fmt.Errorf("Error updating dynamic snippet: service %s, snippet %s, %#v", serviceID, snippetID, err)
		}
	}

	return resourceServiceDynamicSnippetV1Read(d, meta)
}

func resourceServiceDynamicSnippetV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	snippetID := d.Get("snippet_id").(string)

	dynamicSnippet, err := conn.GetDynamicSnippet(&gofastly.GetDynamicSnippetInput{
		Service: serviceID,
		ID:      snippetID,
	})
	if err != nil {
		return err
	}

	err = d.Set("content", dynamicSnippet.Content)
	if err != nil {
		return err
	}

	return nil
}

func resourceServiceDynamicSnippetV1Delete(d *schema.ResourceData, meta interface{}) error {
	// Dynamic snippet content cannot be deleted. Removing from state only
	d.SetId("")
	return nil
}

func resourceServiceDynamicSnippetContentV1Import(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")

	if len(split) != 2 {
		return nil, fmt.Errorf("Invalid id: %s. The ID should be in the format [service_id]/[snippet_id]", d.Id())
	}

	serviceID := split[0]
	snippetID := split[1]

	err := d.Set("service_id", serviceID)
	if err != nil {
		return nil, fmt.Errorf("Error importing dynamic snippet content: service %s, dynamic snippet %s, %s", serviceID, snippetID, err)
	}

	err = d.Set("snippet_id", snippetID)
	if err != nil {
		return nil, fmt.Errorf("Error importing dynamic snippet content: service %s, dynamic snippet %s, %s", serviceID, snippetID, err)
	}

	return []*schema.ResourceData{d}, nil
}
