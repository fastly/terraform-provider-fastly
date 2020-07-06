package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type DirectorServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceDirector(sa ServiceAttributes) ServiceAttributeDefinition {
	return &DirectorServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:               "director",
			serviceAttributes: sa,
		},
	}
}

func (h *DirectorServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	od, nd := d.GetChange(h.GetKey())
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

func (h *DirectorServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	log.Printf("[DEBUG] Refreshing Directors for (%s)", d.Id())
	directorList, err := conn.ListDirectors(&gofastly.ListDirectorsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Directors for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	log.Printf("[DEBUG] Refreshing Backends for (%s)", d.Id())
	backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	})

	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	log.Printf("[DEBUG] Refreshing Director Backends for (%s)", d.Id())
	var directorBackendList []*gofastly.DirectorBackend

	for _, director := range directorList {
		for _, backend := range backendList {
			directorBackendGet, err := conn.GetDirectorBackend(&gofastly.GetDirectorBackendInput{
				Service:  d.Id(),
				Version:  s.ActiveVersion.Number,
				Director: director.Name,
				Backend:  backend.Name,
			})
			if err == nil {
				directorBackendList = append(directorBackendList, directorBackendGet)
			}
		}
	}

	dirl := flattenDirectors(directorList, directorBackendList)

	if err := d.Set(h.GetKey(), dirl); err != nil {
		log.Printf("[WARN] Error setting Directors for (%s): %s", d.Id(), err)
	}

	return nil
}

func (h *DirectorServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
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
	return nil
}

func flattenDirectors(directorList []*gofastly.Director, directorBackendList []*gofastly.DirectorBackend) []map[string]interface{} {
	var dl []map[string]interface{}
	for _, d := range directorList {
		// Convert Director to a map for saving to state.
		nd := map[string]interface{}{
			"name":     d.Name,
			"comment":  d.Comment,
			"shield":   d.Shield,
			"type":     d.Type,
			"quorum":   int(d.Quorum),
			"capacity": int(d.Capacity),
			"retries":  int(d.Retries),
		}

		var b []interface{}
		for _, db := range directorBackendList {
			if d.Name == db.Director {
				b = append(b, db.Backend)
			}
		}
		if len(b) > 0 {
			nd["backends"] = schema.NewSet(schema.HashString, b)
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nd {
			if v == "" {
				delete(nd, k)
			}
		}

		dl = append(dl, nd)
	}
	return dl
}
