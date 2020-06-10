package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

func readDomain(conn *gofastly.Client, d *schema.ResourceData, s *gofastly.ServiceDetail) error {
	// TODO: update go-fastly to support an ActiveVersion struct, which contains
	// domain and backend info in the response. Here we do 2 additional queries
	// to find out that info
	log.Printf("[DEBUG] Refreshing Domains for (%s)", d.Id())
	domainList, err := conn.ListDomains(&gofastly.ListDomainsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Domains for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	// Refresh Domains
	dl := flattenDomains(domainList)

	if err := d.Set("domain", dl); err != nil {
		log.Printf("[WARN] Error setting Domains for (%s): %s", d.Id(), err)
	}
	return nil
}

func flattenDomains(list []*gofastly.Domain) []map[string]interface{} {
	dl := make([]map[string]interface{}, 0, len(list))

	for _, d := range list {
		dl = append(dl, map[string]interface{}{
			"name":    d.Name,
			"comment": d.Comment,
		})
	}

	return dl
}
