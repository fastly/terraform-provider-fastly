package fastly

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenPoolServers(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Server
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Server{
				{
					ServiceID: "service-id",
					PoolID:    "1234567890",
					Address:   "127.0.0.1",
					Weight:    uint(100),
					MaxConn:   uint(200),
					Port:      uint(80),
					Disabled:  false,
					Comment:   "Server 1",
				},
				{
					ServiceID:    "service-id",
					PoolID:       "0987654321",
					Address:      "192.168.0.1",
					Weight:       uint(50),
					MaxConn:      uint(400),
					Port:         uint(88),
					OverrideHost: "origin.notexample.fastly",
					Disabled:     true,
					Comment:      "Server 2",
				},
			},
			local: []map[string]interface{}{
				{
					"address":  "127.0.0.1",
					"weight":   uint(100),
					"max_conn": uint(200),
					"port":     uint(80),
					"disabled": false,
					"comment":  "Server 1",
				},
				{
					"address":       "192.168.0.1",
					"weight":        uint(50),
					"max_conn":      uint(400),
					"port":          uint(88),
					"override_host": "origin.notexample.fastly",
					"disabled":      true,
					"comment":       "Server 2",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenPoolServers(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServicePoolServersV1_create(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	poolName := fmt.Sprintf("pool_%s", acctest.RandString(10))

	expectedRemoteEntries := []map[string]interface{}{
		{
			"id":            "",
			"address":       "127.0.0.1",
			"weight":        uint(100),
			"max_conn":      uint(200),
			"port":          uint(80),
			"override_host": "",
			"disabled":      false,
			"comment":       "Server 1",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePoolServersV1Config_one_pool_with_server(serviceName, poolName, expectedRemoteEntries),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServicePoolServersV1RemoteState(&service, serviceName, poolName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.server", "server.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServicePoolServersV1_update(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	poolName := fmt.Sprintf("Pool %s", acctest.RandString(10))

	expectedRemoteEntries := []map[string]interface{}{
		{
			"id":            "",
			"address":       "127.0.0.1",
			"weight":        100,
			"max_conn":      200,
			"port":          80,
			"override_host": "",
			"disabled":      false,
			"comment":       "Server 1",
		},
	}

	expectedRemoteEntriesAfterUpdate := []map[string]interface{}{
		{
			"id":            "",
			"address":       "127.0.0.1",
			"weight":        100,
			"max_conn":      200,
			"port":          80,
			"override_host": "",
			"disabled":      false,
			"comment":       "Server 1",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePoolServersV1Config_one_pool_with_server(serviceName, poolName, expectedRemoteEntries),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServicePoolServersV1RemoteState(&service, serviceName, poolName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.server", "server.#", "1"),
				),
			},
			{
				Config: testAccServicePoolServersV1Config_one_pool_with_server(serviceName, poolName, expectedRemoteEntriesAfterUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServicePoolServersV1RemoteState(&service, serviceName, poolName, expectedRemoteEntriesAfterUpdate),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.server", "server.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServicePoolServersV1_update_additional_fields(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	poolName := "Server Test Update Disabled Field"

	expectedRemoteEntries := []map[string]interface{}{
		{
			"id":            "",
			"address":       "127.0.0.1",
			"weight":        100,
			"max_conn":      200,
			"port":          80,
			"override_host": "",
			"disabled":      false,
			"comment":       "Server 1",
		},
	}

	expectedRemoteEntriesAfterUpdate := []map[string]interface{}{
		{
			"id":            "",
			"address":       "127.0.0.1",
			"weight":        100,
			"max_conn":      200,
			"port":          80,
			"override_host": "",
			"disabled":      true,
			"comment":       "Server 1 Updated",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePoolServersV1Config_one_pool_with_server(serviceName, poolName, expectedRemoteEntries),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServicePoolServersV1RemoteState(&service, serviceName, poolName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.2838444859.address", "127.0.0.1"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.2838444859.weight", "100"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.2838444859.max_conn", "200"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.2838444859.port", "80"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.2838444859.override_host", ""),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.2838444859.disabled", "false"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.2838444859.comment", "Server 1"),
				),
			},
			{
				Config: testAccServicePoolServersV1Config_one_pool_with_server(serviceName, poolName, expectedRemoteEntriesAfterUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServicePoolServersV1RemoteState(&service, serviceName, poolName, expectedRemoteEntriesAfterUpdate),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.1817859044.address", "127.0.0.1"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.1817859044.weight", "100"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.1817859044.max_conn", "200"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.1817859044.port", "80"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.1817859044.override_host", ""),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.1817859044.disabled", "true"),
					resource.TestCheckResourceAttr("fastly_service_pool_servers_v1.servers", "server.1817859044.comment", "Server 1 Updated"),
				),
			},
		},
	})
}

func TestAccFastlyServicePoolServersV1_delete(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	poolName := fmt.Sprintf("pool_%s", acctest.RandString(10))

	expectedRemoteEntries := []map[string]interface{}{
		{
			"id":            "",
			"address":       "127.0.0.1",
			"weight":        100,
			"max_conn":      200,
			"port":          80,
			"override_host": "",
			"disabled":      false,
			"comment":       "Server 1",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePoolServersV1Config_one_pool_with_server(serviceName, poolName, expectedRemoteEntries),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServicePoolServersV1RemoteState(&service, serviceName, poolName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries_v1.servers", "server.#", "1"),
				),
			},
			{
				Config: testAccServiceDictionaryItemsV1Config_one_pool_no_entries(serviceName, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					resource.TestCheckNoResourceAttr("fastly_service_v1.foo", "servers"),
				),
			},
		},
	})
}

func TestAccFastlyServicePoolServersV1_import(t *testing.T) {

	var service gofastly.ServiceDetail

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	poolName := fmt.Sprintf("pool %s", acctest.RandString(10))

	expectedRemoteEntries := []map[string]interface{}{
		{
			"id":            "",
			"address":       "127.0.0.1",
			"weight":        100,
			"max_conn":      200,
			"port":          80,
			"override_host": "",
			"disabled":      false,
			"comment":       "Server 1",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePoolServersV1Config_one_pool_with_server(name, poolName, expectedRemoteEntries),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
				),
			},
			{
				ResourceName:      "fastly_service_pool_servers_v1.server",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckFastlyServicePoolServersV1RemoteState(service *gofastly.ServiceDetail, serviceName, poolName string, expectedEntries []map[string]interface{}) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		if service.Name != serviceName {
			return fmt.Errorf("[ERR] Bad name, expected (%s), got (%s)", serviceName, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		var pool *gofastly.Pool
		pool, err := conn.GetPool(&gofastly.GetPoolInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
			Name:    poolName,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Pool records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		fmt.Errorf("[ERR] Something Pool name, expected (%s), got (%s)", poolName, pool.Name)

		server, err := conn.GetServer(&gofastly.GetServerInput{
			Service: service.ID,
			Pool:    pool.ID,
			Server:  "server-id-string-here",
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Server records for (%s - %s), version (%v): %s", service.Name, pool.Name, service.ActiveVersion.Number, err)
		}

		serverEntries, err := conn.ListServers(&gofastly.ListServersInput{
			Service: service.ID,
			Pool:    pool.ID,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Pool servers for (%s), Pool (%s - %s): %s", service.Name, pool.ID, server.ID, err)
		}

		flatPoolServers := flattenPoolServers(serverEntries)
		// Clear out the id values to allow a deep equal of the other attributes set in terraform.
		for _, val := range flatPoolServers {
			val["id"] = ""
		}

		sort.Slice(flatPoolServers, func(i, j int) bool {
			return flatPoolServers[i]["address"].(string) < flatPoolServers[j]["address"].(string)
		})

		sort.Slice(expectedEntries, func(i, j int) bool {
			return expectedEntries[i]["address"].(string) < expectedEntries[j]["address"].(string)
		})

		if !reflect.DeepEqual(flatPoolServers, expectedEntries) {
			return fmt.Errorf("[ERR] Error matching:\nexpected: %#v\ngot: %#v", expectedEntries, flatPoolServers)
		}

		return nil
	}
}

func testAccServiceDictionaryItemsV1Config_one_pool_no_entries(serviceName, poolName string) string {

	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}
  backend {
    address = "%s"
    name    = "tf -test backend"
  }
  pool {
	name       = "%s"
	type       = "hash"
  }
  force_destroy = true
}`, serviceName, domainName, backendName, poolName)
}

func testAccServicePoolServersV1Config_one_pool_with_server(serviceName, poolName string, serverList []map[string]interface{}) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	serverEntries := ""

	for _, val := range serverList {
		serverEntries += "server {\n"
		serverEntries += fmt.Sprintf("address = \"%s\"\n", val["address"])
		serverEntries += fmt.Sprintf("weight = %d\n", val["weight"])
		serverEntries += fmt.Sprintf("max_conn = %d\n", val["max_conn"])
		serverEntries += fmt.Sprintf("port = %d\n", val["port"])
		serverEntries += fmt.Sprintf("override_host = \"%s\"\n", val["override_host"])
		serverEntries += fmt.Sprintf("disabled = %t\n", val["disabled"])
		serverEntries += fmt.Sprintf("comment = \"%s\"\n", val["comment"])
		serverEntries += "}\n"
	}

	return fmt.Sprintf(`
variable "pool_name" {
	type = string
	default = "%s"
}

resource "fastly_service_v1" "foo" {
	name = "%s"
	domain {
		name    = "%s"
		comment = "tf-testing-domain"
	}
	backend {
		address = "%s"
		name    = "tf-testing-backend"
	}
	pool {
		name       = var.pool_name
	}
	force_destroy = true
}

resource "fastly_service_pool_servers_v1" "servers" {
	service_id = fastly_service_v1.foo.id
	# pool_id = {for s in fastly_service_v1.foo.pool : s.name => s.id}[var.pool_name]
         pool_id    = {for p in fastly_service_v1.foo.pool : p.name => p.pool_id}[var.pool_name]
	%s

}`, poolName, serviceName, domainName, backendName, serverEntries)
}
