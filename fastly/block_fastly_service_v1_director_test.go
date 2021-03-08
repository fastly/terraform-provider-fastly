package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenDirectors(t *testing.T) {
	cases := []struct {
		remote_director        []*gofastly.Director
		remote_directorbackend []*gofastly.DirectorBackend

		local []map[string]interface{}
	}{
		{
			remote_director: []*gofastly.Director{
				{
					Name:     "somedirector",
					Type:     3,
					Quorum:   75,
					Capacity: 25,
					Retries:  10,
				},
			},
			remote_directorbackend: []*gofastly.DirectorBackend{
				{
					Director: "somedirector",
					Backend:  "somebackend",
				},
			},
			local: []map[string]interface{}{
				{
					"name":     "somedirector",
					"type":     3,
					"quorum":   75,
					"capacity": 25,
					"retries":  10,
					"backends": schema.NewSet(schema.HashString, []interface{}{"somebackend"}),
				},
			},
		},
		{
			remote_director: []*gofastly.Director{
				{
					Name: "somedirector",
				},
				{
					Name: "someotherdirector",
				},
			},
			remote_directorbackend: []*gofastly.DirectorBackend{
				{
					Director: "somedirector",
					Backend:  "somebackend",
				},
				{
					Director: "somedirector",
					Backend:  "someotherbackend",
				},
				{
					Director: "someotherdirector",
					Backend:  "somebackend",
				},
				{
					Director: "someotherdirector",
					Backend:  "someotherbackend",
				},
			},
			local: []map[string]interface{}{
				{
					"name":     "somedirector",
					"backends": schema.NewSet(schema.HashString, []interface{}{"somebackend", "someotherbackend"}),
				},
				{
					"name":     "someotherdirector",
					"backends": schema.NewSet(schema.HashString, []interface{}{"somebackend", "someotherbackend"}),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDirectors(c.remote_director, c.remote_directorbackend)
		// loop, because deepequal wont work with our sets
		expectedCount := len(c.local)
		var found int
		for _, o := range out {
			for _, l := range c.local {
				if o["name"].(string) == l["name"].(string) {
					found++
					if o["backends"] == nil && l["backends"] != nil {
						t.Fatalf("output backends are nil, local are not")
					}

					if o["backends"] != nil {
						oex := o["backends"].(*schema.Set)
						lex := l["backends"].(*schema.Set)
						if !oex.Equal(lex) {
							t.Fatalf("Backends don't match, expected: %#v, got: %#v", lex, oex)
						}
					}
				}
			}
		}

		if found != expectedCount {
			t.Fatalf("Found and expected mismatch: %d / %d", found, expectedCount)
		}
	}
}

func TestAccFastlyServiceV1_directors_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	// Director + Backend 1
	directorDeveloper := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_developer",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
	}
	directorBackendDeveloper := gofastly.DirectorBackend{
		Director: "director_developer",
		Backend:  "developer",
	}

	// Director + Backend 2
	directorDeveloperUpdated := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_developer",
		Type:           4,
		Quorum:         30,
		Capacity:       25,
		Retries:        10,
	}
	directorBackendDeveloperUpdated := gofastly.DirectorBackend{
		Director: "director_developer",
		Backend:  "developer_updated",
	}

	// Director + Backend 3
	directorApps := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_apps",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
	}
	directorBackendApps := gofastly.DirectorBackend{
		Director: "director_apps",
		Backend:  "apps",
	}

	// Director + Backend 4
	directorWWWDemo := gofastly.Director{
		ServiceVersion: 1,
		Name:           "director_www_demo",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
	}
	directorBackendWWW := gofastly.DirectorBackend{
		Director: "director_www_demo",
		Backend:  "www",
	}
	directorBackendDemo := gofastly.DirectorBackend{
		Director: "director_www_demo",
		Backend:  "demo",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1DirectorsConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DirectorsAttributes(
						&service,
						[]*gofastly.Director{&directorDeveloper, &directorApps},
						[]*gofastly.DirectorBackend{&directorBackendDeveloper, &directorBackendApps}),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "director.#", "2"),
				),
			},

			{
				Config: testAccServiceV1DirectorsConfig_update(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DirectorsAttributes(
						&service,
						[]*gofastly.Director{&directorDeveloperUpdated, &directorApps, &directorWWWDemo},
						[]*gofastly.DirectorBackend{&directorBackendDeveloperUpdated, &directorBackendApps, &directorBackendWWW, &directorBackendDemo}),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "director.#", "3"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1DirectorsAttributes(service *gofastly.ServiceDetail, directors []*gofastly.Director, director_backends []*gofastly.DirectorBackend) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*FastlyClient).conn
		directorList, err := conn.ListDirectors(&gofastly.ListDirectorsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Directors for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(directorList) != len(directors) {
			return fmt.Errorf("Director count mismatch, expected (%d), got (%d)", len(directors), len(directorList))
		}

		var found int
		for _, d := range directors {
			for _, ld := range directorList {
				if d.Name == ld.Name {
					// we don't know these things ahead of time, so populate them now
					d.ServiceID = service.ID
					d.ServiceVersion = service.ActiveVersion.Number
					ld.CreatedAt = nil
					ld.UpdatedAt = nil
					if !reflect.DeepEqual(d, ld) {
						return fmt.Errorf("Bad match Director match, expected (%#v), got (%#v)", d, ld)
					}
					found++
				}
			}
		}

		if found != len(directors) {
			return fmt.Errorf("Error matching Director rules")
		}

		return nil
	}
}

func testAccServiceV1DirectorsConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "developer.fastly.com"
    name    = "developer"
  }

  backend {
    address = "apps.fastly.com"
    name    = "apps"
    weight  = 1
  }

  director {
    name = "director_developer"
    type = 3
    backends = [ "developer" ]
  }

  director {
    name = "director_apps"
    type = 3
    backends = [ "apps" ]
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceV1DirectorsConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "developer.fastly.com"
    name    = "developer_updated"
  }

  backend {
    address = "apps.fastly.com"
    name    = "apps"
    weight  = 9
  }

  backend {
    address = "www.fastly.com"
    name    = "www"
  }

  backend {
    address = "www.fastlydemo.net"
    name    = "demo"
  }

  director {
    name = "director_developer"
    type = 4
    quorum = 30
    retries = 10
    capacity = 25
    backends = [ "developer_updated" ]
  }

  director {
    name = "director_apps"
    type = 3
    backends = [ "apps" ]
  }

  director {
    name = "director_www_demo"
    type = 3
    backends = [ "www", "demo" ]
  }

  force_destroy = true
}`, name, domain)
}
