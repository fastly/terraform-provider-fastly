package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var domainSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Required: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The domain that this Service will respond to",
			},

			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	},
}

func processDomain(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	od, nd := d.GetChange("domain")
	if od == nil {
		od = new(schema.Set)
	}
	if nd == nil {
		nd = new(schema.Set)
	}

	ods := od.(*schema.Set)
	nds := nd.(*schema.Set)

	remove := ods.Difference(nds).List()
	add := nds.Difference(ods).List()

	// Delete removed domains
	for _, dRaw := range remove {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteDomainInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Fastly Domain removal opts: %#v", opts)
		err := conn.DeleteDomain(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new Domains
	for _, dRaw := range add {
		df := dRaw.(map[string]interface{})
		opts := gofastly.CreateDomainInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		if v, ok := df["comment"]; ok {
			opts.Comment = v.(string)
		}

		log.Printf("[DEBUG] Fastly Domain Addition opts: %#v", opts)
		_, err := conn.CreateDomain(&opts)
		if err != nil {
			return err
		}
	}
	return nil
}
