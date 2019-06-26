package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"reflect"
	"testing"
)

func TestResourceFastlyFlattenAclEntries(t *testing.T) {
	cases := []struct {
		remote []*gofastly.ACLEntry
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.ACLEntry{
				{
					ServiceID: "service-id",
					ACLID:     "1234567890",
					IP:        "127.0.0.1",
					Subnet:    "24",
					Negated:   false,
					Comment:   "ALC Entry 1",
				},
				{
					ServiceID: "service-id",
					ACLID:     "0987654321",
					IP:        "192.168.0.1",
					Subnet:    "16",
					Negated:   true,
					Comment:   "ALC Entry 2",
				},
			},
			local: []map[string]interface{}{
				{
					"ip":      "127.0.0.1",
					"subnet":  "24",
					"negated": false,
					"comment": "ALC Entry 1",
				},
				{
					"ip":      "192.168.0.1",
					"subnet":  "16",
					"negated": true,
					"comment": "ALC Entry 2",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenAclEntries(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceAclEntriesV1_create(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("ACL %s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAclEntriesV1Config_create(serviceName, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceDictionaryItemsV1RemoteState(&service, serviceName, aclName),
					resource.TestCheckResourceAttr("fastly_service_acl_entries_v1.entries", "entry.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceDictionaryItemsV1RemoteState(service *gofastly.ServiceDetail, serviceName, aclName string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		// TODO check against remote state through api.
		return nil
	}
}

func testAccServiceAclEntriesV1Config_create(serviceName, aclName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
variable "myacl_name" {
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
	acl {
		name       = var.myacl_name
	}
	force_destroy = true
}
 resource "fastly_service_acl_entries_v1" "entries" {
	service_id = "${fastly_service_v1.foo.id}"
	acl_id = "${{for s in fastly_service_v1.foo.acl : s.name => s.acl_id}[var.myacl_name]}"
	entry {
		ip = "127.0.0.1"
		subnet = "24"
		negated = false
		comment = "ALC Entry 1"
	}

	entry {
		ip = "192.168.0.1"
		subnet = "24"
		negated = false
		comment = "ALC Entry 2"
	}
}`, aclName, serviceName, domainName, backendName)
}
