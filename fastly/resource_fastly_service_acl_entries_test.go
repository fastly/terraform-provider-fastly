package fastly

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

func TestResourceFastlyFlattenAclEntries(t *testing.T) {
	cases := []struct {
		remote []*gofastly.ACLEntry
		local  []map[string]any
	}{
		{
			remote: []*gofastly.ACLEntry{
				{
					ServiceID: gofastly.ToPointer("service-id"),
					ACLID:     gofastly.ToPointer("1234567890"),
					IP:        gofastly.ToPointer("127.0.0.1"),
					Subnet:    gofastly.ToPointer(24),
					Negated:   gofastly.ToPointer(false),
					Comment:   gofastly.ToPointer("ACL Entry 1"),
				},
				{
					ServiceID: gofastly.ToPointer("service-id"),
					ACLID:     gofastly.ToPointer("0987654321"),
					IP:        gofastly.ToPointer("192.168.0.1"),
					Subnet:    gofastly.ToPointer(16),
					Negated:   gofastly.ToPointer(true),
					Comment:   gofastly.ToPointer("ACL Entry 2"),
				},
			},
			local: []map[string]any{
				{
					"ip":      "127.0.0.1",
					"subnet":  "24",
					"negated": false,
					"comment": "ACL Entry 1",
				},
				{
					"ip":      "192.168.0.1",
					"subnet":  "16",
					"negated": true,
					"comment": "ACL Entry 2",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenACLEntries(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceAclEntries_create(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("ACL %s", acctest.RandString(10))

	expectedRemoteEntries := []map[string]any{
		{
			"id":      "",
			"ip":      "127.0.0.1",
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 1",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, expectedRemoteEntries, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
				),
			},
			{
				ResourceName:            "fastly_service_acl_entries.entries",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manage_entries"},
			},
		},
	})
}

func TestAccFastlyServiceAclEntries_create_more_than_one_page(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("ACL %s", acctest.RandString(10))

	defaultPerPage := 100
	var expectedRemoteEntries []map[string]any
	for i := 1; i <= defaultPerPage*2; i++ {
		entry := map[string]any{
			"id":      "",
			"ip":      fmt.Sprintf("%d.0.0.1", i),
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 1",
		}
		expectedRemoteEntries = append(expectedRemoteEntries, entry)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, expectedRemoteEntries, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "200"),
				),
			},
			{
				ResourceName:            "fastly_service_acl_entries.entries",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manage_entries"},
			},
		},
	})
}

func TestAccFastlyServiceAclEntries_create_update(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("ACL %s", acctest.RandString(10))

	expectedRemoteEntries := []map[string]any{
		{
			"id":      "",
			"ip":      "127.0.0.1",
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 1",
		},
	}

	expectedRemoteEntriesAfterUpdate := []map[string]any{
		{
			"id":      "",
			"ip":      "127.0.0.2",
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 1",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, expectedRemoteEntries, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
				),
			},
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, expectedRemoteEntriesAfterUpdate, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, expectedRemoteEntriesAfterUpdate),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
				),
			},
			{
				ResourceName:            "fastly_service_acl_entries.entries",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manage_entries"},
			},
		},
	})
}

func TestAccFastlyServiceAclEntries_update_additional_fields(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := "ACL Test Update Negated Field"

	expectedRemoteEntries := []map[string]any{
		{
			"id":      "",
			"ip":      "127.0.0.1",
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 1",
		},
	}

	expectedRemoteEntriesAfterUpdate := []map[string]any{
		{
			"id":      "",
			"ip":      "127.0.0.1",
			"subnet":  "20",
			"negated": true,
			"comment": "ACL Entry 1 Updated",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, expectedRemoteEntries, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_acl_entries.entries", "entry.*", map[string]string{
						"ip":      "127.0.0.1",
						"subnet":  "24",
						"negated": "false",
						"comment": "ACL Entry 1",
					}),
				),
			},
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, expectedRemoteEntriesAfterUpdate, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, expectedRemoteEntriesAfterUpdate),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_acl_entries.entries", "entry.*", map[string]string{
						"ip":      "127.0.0.1",
						"subnet":  "20",
						"negated": "true",
						"comment": "ACL Entry 1 Updated",
					}),
				),
			},
		},
	})
}

func TestAccFastlyServiceAclEntries_delete(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("ACL %s", acctest.RandString(10))

	expectedRemoteEntries := []map[string]any{
		{
			"id":      "",
			"ip":      "127.0.0.1",
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 1",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, expectedRemoteEntries, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
				),
			},
			{
				Config: testAccServiceDictionaryItemsV1ConfigOneACLNoEntries(serviceName, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckNoResourceAttr("fastly_service_vcl.foo", "entry"),
				),
			},
		},
	})
}

func TestAccFastlyServiceAclEntries_process_1001_entries(t *testing.T) {
	var service gofastly.ServiceDetail

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("acl %s", acctest.RandString(10))

	expectedBatchSize := gofastly.BatchModifyMaximumOperations + 1

	expectedRemoteEntries := make([]map[string]any, 0)

	ipPart3 := 0
	ipPart4 := 1
	for i := 0; i < expectedBatchSize; i++ {
		if ipPart4 > 254 {
			ipPart3++
			ipPart4 = 1
		}

		expectedRemoteEntries = append(expectedRemoteEntries, map[string]any{
			"id":      "",
			"ip":      fmt.Sprintf("127.0.%d.%d", ipPart3, ipPart4),
			"subnet":  "22",
			"negated": false,
			"comment": fmt.Sprintf("ACL Entry %d %d", ipPart3, ipPart4),
		})

		ipPart4++
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(name, aclName, expectedRemoteEntries, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, name, aclName, expectedRemoteEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", strconv.Itoa(expectedBatchSize)),
				),
			},
		},
	})
}

func TestAccFastlyServiceAclEntries_manage_entries_false(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("ACL %s", acctest.RandString(10))

	initialEntries := []map[string]any{
		{
			"id":      "",
			"ip":      "127.0.0.1",
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 1",
		},
	}

	updatedEntries := []map[string]any{
		{
			"id":      "",
			"ip":      "127.0.0.1",
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 1",
		},
		{
			"id":      "",
			"ip":      "127.0.0.2",
			"subnet":  "24",
			"negated": false,
			"comment": "ACL Entry 2",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, initialEntries, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, initialEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
				),
			},
			{
				Config: testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName, updatedEntries, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceACLEntriesRemoteState(&service, serviceName, aclName, initialEntries),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.entries", "entry.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceACLEntriesRemoteState(service *gofastly.ServiceDetail, serviceName, aclName string, expectedEntries []map[string]any) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != serviceName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", serviceName, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		acl, err := conn.GetACL(context.TODO(), &gofastly.GetACLInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
			Name:           aclName,
		})
		if err != nil {
			return fmt.Errorf("error looking up ACL records for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		aclEntries, err := getAllACLEntriesViaPaginator(context.TODO(), conn, &gofastly.GetACLEntriesInput{
			ServiceID: gofastly.ToValue(service.ServiceID),
			ACLID:     gofastly.ToValue(acl.ACLID),
		})
		if err != nil {
			return fmt.Errorf("error looking up ACL entry records for (%s), ACL (%s): %s", gofastly.ToValue(service.Name), gofastly.ToValue(acl.ACLID), err)
		}

		flatACLEntries := flattenACLEntries(aclEntries)
		// Clear out the id values to allow a deep equal of the other attributes set in terraform.
		for _, val := range flatACLEntries {
			val["id"] = ""
		}

		sort.Slice(flatACLEntries, func(i, j int) bool {
			return flatACLEntries[i]["ip"].(string) < flatACLEntries[j]["ip"].(string)
		})

		sort.Slice(expectedEntries, func(i, j int) bool {
			return expectedEntries[i]["ip"].(string) < expectedEntries[j]["ip"].(string)
		})

		if !reflect.DeepEqual(flatACLEntries, expectedEntries) {
			return fmt.Errorf("error matching:\nexpected: %#v\ngot: %#v", expectedEntries, flatACLEntries)
		}

		return nil
	}
}

func testAccServiceDictionaryItemsV1ConfigOneACLNoEntries(serviceName, aclName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}
  backend {
    address = "%s"
    name    = "tf -test backend"
  }
  acl {
	name       = "%s"
  }
  force_destroy = true
}`, serviceName, domainName, backendName, aclName)
}

func testAccServiceACLEntriesConfigOneACLWithEntries(serviceName, aclName string, aclEntriesList []map[string]any, manageEntries bool) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	aclEntries := ""

	for _, val := range aclEntriesList {
		aclEntries += "entry {\n"
		aclEntries += fmt.Sprintf("ip = \"%s\"\n", val["ip"])
		aclEntries += fmt.Sprintf("subnet = \"%s\"\n", val["subnet"])
		aclEntries += fmt.Sprintf("negated = %t\n", val["negated"])
		aclEntries += fmt.Sprintf("comment = \"%s\"\n", val["comment"])
		aclEntries += "}\n"
	}

	return fmt.Sprintf(`
variable "myacl_name" {
	type = string
	default = "%s"
}

resource "fastly_service_vcl" "foo" {
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
 resource "fastly_service_acl_entries" "entries" {
	service_id = fastly_service_vcl.foo.id
	acl_id = {for s in fastly_service_vcl.foo.acl : s.name => s.acl_id}[var.myacl_name]
	manage_entries = %t
	%s
}`, aclName, serviceName, domainName, backendName, manageEntries, aclEntries)
}
