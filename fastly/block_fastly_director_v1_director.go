package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var directorSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name to refer to this director",
			},
			"backends": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of backends associated with this director",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			// optional fields
			"capacity": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Load balancing weight for the backends",
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"shield": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Selected POP to serve as a 'shield' for origin servers.",
			},
			"quorum": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      75,
				Description:  "Percentage of capacity that needs to be up for the director itself to be considered up",
				ValidateFunc: validateDirectorQuorum(),
			},
			"type": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				Description:  "Type of load balance group to use. Integer, 1 to 4. Values: 1 (random), 3 (hash), 4 (client)",
				ValidateFunc: validateDirectorType(),
			},
			"retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "How many backends to search if it fails",
			},
		},
	},
}


func processDirector(d *schema.ResourceData, conn *gofastly.Client, latestVersion int) error {
	od, nd := d.GetChange("director")
	if od == nil {
		od = new(schema.Set)
	}
	if nd == nil {
		nd = new(schema.Set)
	}

	ods := od.(*schema.Set)
	nds := nd.(*schema.Set)

	removeDirector := ods.Difference(nds).List()
	addDirector := nds.Difference(ods).List()

	// DELETE old director configurations
	for _, dRaw := range removeDirector {
		df := dRaw.(map[string]interface{})
		opts := gofastly.DeleteDirectorInput{
			Service: d.Id(),
			Version: latestVersion,
			Name:    df["name"].(string),
		}

		log.Printf("[DEBUG] Director Removal opts: %#v", opts)
		err := conn.DeleteDirector(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// POST new/updated Director
	for _, dRaw := range addDirector {
		df := dRaw.(map[string]interface{})
		opts := gofastly.CreateDirectorInput{
			Service:  d.Id(),
			Version:  latestVersion,
			Name:     df["name"].(string),
			Comment:  df["comment"].(string),
			Shield:   df["shield"].(string),
			Capacity: uint(df["capacity"].(int)),
			Quorum:   uint(df["quorum"].(int)),
			Retries:  uint(df["retries"].(int)),
		}

		switch df["type"].(int) {
		case 1:
			opts.Type = gofastly.DirectorTypeRandom
		case 2:
			opts.Type = gofastly.DirectorTypeRoundRobin
		case 3:
			opts.Type = gofastly.DirectorTypeHash
		case 4:
			opts.Type = gofastly.DirectorTypeClient
		}

		log.Printf("[DEBUG] Director Create opts: %#v", opts)
		_, err := conn.CreateDirector(&opts)
		if err != nil {
			return err
		}

		if v, ok := df["backends"]; ok {
			if len(v.(*schema.Set).List()) > 0 {
				for _, b := range v.(*schema.Set).List() {
					opts := gofastly.CreateDirectorBackendInput{
						Service:  d.Id(),
						Version:  latestVersion,
						Director: df["name"].(string),
						Backend:  b.(string),
					}

					log.Printf("[DEBUG] Director Backend Create opts: %#v", opts)
					_, err := conn.CreateDirectorBackend(&opts)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}