package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
)

func resourceServicePoolServersV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServicePoolServersV1Create,
		Read:   resourceServicePoolServersV1Read,
		Update: resourceServicePoolServersV1Update,
		Delete: resourceServicePoolServersV1Delete,
		Importer: &schema.ResourceImporter{
			State: resourceServicePoolServersV1Import,
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service Id",
			},

			"pool_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Pool Id",
			},
			"server": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Server Entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "",
							Computed:    true,
						},
						"weight": {
							Type:        schema.TypeInt,
							Description: "Weight (1-100) used to load balance this server against others. Optional. Defaults to '100' if not set.",
							Optional:    true,
						},
						"max_conn": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum number of connections. If the value is '0', it inherits the value from pool's max_conn_default. Optional. Defaults to '0' if not set.",
						},
						"port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Port number. Setting port 443 does not force TLS. Set use_tls in pool to force TLS. Optional. Defaults to '80' if not set.",
						},
						"address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A hostname, IPv4, or IPv6 address for the server. Required.",
						},
						"comment": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A personal freeform descriptive note",
						},
						"disabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allows servers to be enabled and disabled in a pool.",
						},
						"override_host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The hostname to override the Host header. Optional. Defaults to null meaning no override of the Host header if not set.",
						},
					},
				},
			},
		},
	}

}

func resourceServicePoolServersV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	poolID := d.Get("pool_id").(string)
	servers := d.Get("server").(*schema.Set)

	fmt.Printf("server===> %v", servers)

	for _, vRaw := range servers.List() {
		val := vRaw.(map[string]interface{})

		weight := uint(val["weight"].(int))
		fmt.Printf("%v", weight)
		max_conn := uint(val["max_conn"].(int))
		port := uint(val["port"].(int))
		comment := val["comment"].(string)
		disabled := val["disabled"].(bool)
		override_host := val["override_host"].(string)

		opts := gofastly.CreateServerInput{
			Service:      serviceID,
			Pool:         poolID,
			Weight:       &weight,
			MaxConn:      &max_conn,
			Port:         &port,
			Address:      val["address"].(string),
			Comment:      &comment,
			Disabled:     &disabled,
			OverrideHost: &override_host,
		}

		_, err := conn.CreateServer(&opts)
		if err != nil {
			return fmt.Errorf("Error creating Pool servers: service %s, Pool %s, %s", serviceID, poolID, err)
		}
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, poolID))
	return resourceServicePoolServersV1Read(d, meta)
}

func resourceServicePoolServersV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	poolID := d.Get("pool_id").(string)

	poolServers, err := conn.ListServers(&gofastly.ListServersInput{
		Service: serviceID,
		Pool:    poolID,
	})

	if err != nil {
		return err
	}

	d.Set("server", flattenPoolServers(poolServers))
	return nil
}

func resourceServicePoolServersV1Update(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	poolID := d.Get("pool_id").(string)

	if d.HasChange("server") {

		oe, ne := d.GetChange("server")
		if oe == nil {
			oe = new(schema.Set)
		}
		if ne == nil {
			ne = new(schema.Set)
		}

		oes := oe.(*schema.Set)
		nes := ne.(*schema.Set)

		removeServers := oes.Difference(nes).List()
		addServers := nes.Difference(oes).List()

		// DELETE old Server entries
		for _, vRaw := range removeServers {
			val := vRaw.(map[string]interface{})
			err := conn.DeleteServer(&gofastly.DeleteServerInput{
				Service: serviceID,
				Pool:    poolID,
				Server:  val["id"].(string),
			})
			if err != nil {
				return fmt.Errorf("Error deleting Pool Server: service %s, Pool %s, %s", serviceID, poolID, err)
			}
		}

		// POST new Server entry
		for _, vRaw := range addServers {
			val := vRaw.(map[string]interface{})

			weight := uint(val["weight"].(int))
			max_conn := uint(val["max_conn"].(int))
			port := uint(val["port"].(int))
			comment := val["comment"].(string)
			disabled := val["disabled"].(bool)
			override_host := val["override_host"].(string)

			opts := gofastly.CreateServerInput{
				Service:      serviceID,
				Pool:         poolID,
				Weight:       &weight,
				MaxConn:      &max_conn,
				Port:         &port,
				Address:      val["address"].(string),
				Comment:      &comment,
				Disabled:     &disabled,
				OverrideHost: &override_host,
			}

			_, err := conn.CreateServer(&opts)
			if err != nil {
				return fmt.Errorf("Error creating Pool servers: service %s, Pool %s, %s", serviceID, poolID, err)
			}
		}
	}

	return resourceServicePoolServersV1Read(d, meta)
}

func resourceServicePoolServersV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	poolID := d.Get("pool_id").(string)
	servers := d.Get("server").(*schema.Set)

	for _, vRaw := range servers.List() {
		val := vRaw.(map[string]interface{})

		err := conn.DeleteServer(&gofastly.DeleteServerInput{
			Service: serviceID,
			Pool:    poolID,
			Server:  val["id"].(string),
		})
		if err != nil {
			return fmt.Errorf("Error deleting Pool Server: service %s, Pool %s, %s", serviceID, poolID, err)
		}
	}

	d.SetId("")
	return nil
}

func flattenPoolServers(poolServersList []*gofastly.Server) []map[string]interface{} {

	var resultList []map[string]interface{}

	for _, currentPoolServer := range poolServersList {
		ps := map[string]interface{}{
			"id":            currentPoolServer.ID,
			"weight":        currentPoolServer.Weight,
			"max_conn":      currentPoolServer.MaxConn,
			"port":          currentPoolServer.Port,
			"address":       currentPoolServer.Address,
			"comment":       currentPoolServer.Comment,
			"disabled":      currentPoolServer.Disabled,
			"override_host": currentPoolServer.OverrideHost,
		}

		for k, v := range ps {
			if v == "" {
				delete(ps, k)
			}
		}

		resultList = append(resultList, ps)
	}

	return resultList
}

func resourceServicePoolServersV1Import(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")

	if len(split) != 2 {
		return nil, fmt.Errorf("Invalid id: %s. The ID should be in the format [service_id]/[pool_id]", d.Id())
	}

	serviceID := split[0]
	poolID := split[1]

	err := d.Set("service_id", serviceID)
	if err != nil {
		return nil, fmt.Errorf("Error importing Pool Servers: service %s, Pool %s, %s", serviceID, poolID, err)
	}

	err = d.Set("pool_id", poolID)
	if err != nil {
		return nil, fmt.Errorf("Error importing Pool Servers: service %s, Pool %s, %s", serviceID, poolID, err)
	}

	return []*schema.ResourceData{d}, nil
}
