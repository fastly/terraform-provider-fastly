package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

	createdDir1 := gofastly.Director{
		ServiceVersion: 1,
		Name:           "mydirector",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
	}
	createdDb1 := gofastly.DirectorBackend{
		Director: "mydirector",
		Backend:  "origin old",
	}

	updatedDir1 := gofastly.Director{
		ServiceVersion: 1,
		Name:           "mydirector",
		Type:           4,
		Quorum:         30,
		Capacity:       25,
		Retries:        10,
	}
	updatedDb1 := gofastly.DirectorBackend{
		Director: "mydirector",
		Backend:  "origin new",
	}

	createdDir2 := gofastly.Director{
		ServiceVersion: 1,
		Name:           "unchangeddirector",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
	}
	createdDb2 := gofastly.DirectorBackend{
		Director: "unchangeddirector",
		Backend:  "origin apps",
	}

	// Updated director should be the same as the created ones
	updatedDir2 := createdDir2
	updatedDb2 := createdDb2

	updatedDir3 := gofastly.Director{
		ServiceVersion: 1,
		Name:           "myotherdirector",
		Type:           3,
		Quorum:         75,
		Capacity:       100,
		Retries:        5,
	}
	updatedDb3x := gofastly.DirectorBackend{
		Director: "myotherdirector",
		Backend:  "origin x",
	}
	updatedDb3y := gofastly.DirectorBackend{
		Director: "myotherdirector",
		Backend:  "origin y",
	}

	resource.Test(t, resource.TestCase{
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
						[]*gofastly.Director{&createdDir1, &createdDir2},
						[]*gofastly.DirectorBackend{&createdDb1, &createdDb2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "director.#", "2"),
				),
			},

			{
				Config: testAccServiceV1DirectorsConfig_update(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DirectorsAttributes(
						&service,
						[]*gofastly.Director{&updatedDir1, &updatedDir2, &updatedDir3},
						[]*gofastly.DirectorBackend{&updatedDb1, &updatedDb2, &updatedDb3x, &updatedDb3y}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "director.#", "3"),
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
    name    = "origin old"
  }

  backend {
    address = "apps.fastly.com"
    name    = "origin apps"
    weight  = 1
  }

  director {
    name = "mydirector"
    type = 3
    backends = [ "origin old" ]
  }

  director {
    name = "unchangeddirector"
    type = 3
    backends = [ "origin apps" ]
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
    name    = "origin new"
  }

  backend {
    address = "apps.fastly.com"
    name    = "origin apps"
    weight  = 9
  }

  backend {
    address = "www.fastly.com"
    name    = "origin x"
  }

  backend {
    address = "www.fastlydemo.net"
    name    = "origin y"
  }

  director {
    name = "mydirector"
    type = 4
    quorum = 30
    retries = 10
    capacity = 25
    backends = [ "origin new" ]
  }

  director {
    name = "unchangeddirector"
    type = 3
    backends = [ "origin apps" ]
  }

  director {
    name = "myotherdirector"
    type = 3
    backends = [ "origin x", "origin y" ]
  }

  force_destroy = true
}`, name, domain)
}
