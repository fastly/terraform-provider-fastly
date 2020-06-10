package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var snippetSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name to refer to this VCL snippet",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of init, recv, hit, miss, pass, fetch, error, deliver, log, none",
				ValidateFunc: validateSnippetType(),
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The contents of the VCL snippet",
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Determines ordering for multiple snippets. Lower priorities execute first. (Default: 100)",
			},
		},
	},
}

func processSnippet(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	// Note: as above with Gzip and S3 logging, we don't utilize the PUT
	// endpoint to update a VCL snippet, we simply destroy it and create a new one.
	oldSnippetVal, newSnippetVal := d.GetChange("snippet")
	if oldSnippetVal == nil {
		oldSnippetVal = new(schema.Set)
	}
	if newSnippetVal == nil {
		newSnippetVal = new(schema.Set)
	}

	oldSnippetSet := oldSnippetVal.(*schema.Set)
	newSnippetSet := newSnippetVal.(*schema.Set)

	remove := oldSnippetSet.Difference(newSnippetSet).List()
	add := newSnippetSet.Difference(oldSnippetSet).List()

	// Delete removed VCL Snippet configurations
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteSnippetInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly VCL Snippet Removal opts: %#v", opts)
		err := conn.DeleteSnippet(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new VCL Snippet configurations
	for _, dRaw := range add {
		opts, err := buildSnippet(dRaw.(map[string]interface{}))
		if err != nil {
			log.Printf("[DEBUG] Error building VCL Snippet: %s", err)
			return err
		}
		opts.Service = d.Id()
		opts.Version = latestVersion

		log.Printf("[DEBUG] Fastly VCL Snippet Addition opts: %#v", opts)
		_, err = conn.CreateSnippet(opts)
		if err != nil {
			return err
		}
	}
	return nil
}
