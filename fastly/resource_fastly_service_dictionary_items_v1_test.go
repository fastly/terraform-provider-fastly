package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceFastlyFlattenDictionaryItems(t *testing.T) {
	cases := []struct {
		remote []*gofastly.DictionaryItem
		local  map[string]string
	}{
		{
			remote: []*gofastly.DictionaryItem{
				{
					ServiceID:    "service-id",
					DictionaryID: "1234567890",
					ItemKey:      "key-1",
					ItemValue:    "value-1",
				},
				{
					ServiceID:    "service-id",
					DictionaryID: "1234567890",
					ItemKey:      "key-2",
					ItemValue:    "value-2",
				},
			},
			local: map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
			},
		},
	}

	for _, c := range cases {
		out := flattenDictionaryItems(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceDictionaryItemV1(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	dictName := fmt.Sprintf("dict %s", acctest.RandString(10))

	expectedItems := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDictionaryItemsV1Config(name, dictName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceDictionaryItemsV1Attributes(&service, name, dictName, expectedItems),
				),
			},
		},
	})
}

func TestAccFastlyServiceDictionaryItemV1_update_items(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	dictName := fmt.Sprintf("dict %s", acctest.RandString(10))

	expectedItems := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	expectedItemsAfterUpdate := map[string]string{
		"key2": "valuetwo",
		"key3": "value3",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDictionaryItemsV1Config(name, dictName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceDictionaryItemsV1Attributes(&service, name, dictName, expectedItems),
				),
			},
			{
				Config: testAccServiceDictionaryItemsV1ConfigUpdateItems(name, dictName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceDictionaryItemsV1Attributes(&service, name, dictName, expectedItemsAfterUpdate),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceDictionaryItemsV1Attributes(service *gofastly.ServiceDetail, name, dictName string, expectedItems map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		dict, err := conn.GetDictionary(&gofastly.GetDictionaryInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
			Name:    dictName,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionary records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		dictItems, err := conn.ListDictionaryItems(&gofastly.ListDictionaryItemsInput{
			Service:    service.ID,
			Dictionary: dict.ID,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionary Items records for (%s), dictionary (%s): %s", service.Name, dict.ID, err)
		}

		dictItemsMap := flattenDictionaryItems(dictItems)

		if !reflect.DeepEqual(dictItemsMap, expectedItems) {
			return fmt.Errorf("[ERR] Error matching:\nexpected: %#v\ngot: %#v", expectedItems, dictItemsMap)
		}

		return nil
	}
}


func testAccServiceDictionaryItemsV1Config(serviceName, dictName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
variable "mydict" {
	type = object({ name=string, items=map(string) })
	default = {
		name = "%s" 
		items = {
			key1: "value1"
			key2: "value2"
		}
	}
}

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

  dictionary {
	name       = var.mydict.name
	write_only = false
  }

  force_destroy = true
}

resource "fastly_service_dictionary_items_v1" "items" {
    service_id = "${fastly_service_v1.foo.id}"
    dictionary_id = "${{for s in fastly_service_v1.foo.dictionary : s.name => s.dictionary_id}[var.mydict.name]}"
    items = var.mydict.items
}`, dictName, serviceName, domainName, backendName)
}

func testAccServiceDictionaryItemsV1ConfigUpdateItems(serviceName, dictName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
variable "mydict" {
	type = object({ name=string, items=map(string) })
	default = {
		name = "%s" 
		items = {
			key2: "valuetwo"
			key3: "value3"
		}
	}
}

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

  dictionary {
	name       = var.mydict.name
	write_only = false
  }

  force_destroy = true
}

resource "fastly_service_dictionary_items_v1" "items" {
    service_id = "${fastly_service_v1.foo.id}"
    dictionary_id = "${{for s in fastly_service_v1.foo.dictionary : s.name => s.dictionary_id}[var.mydict.name]}"
    items = var.mydict.items
}`, dictName, serviceName, domainName, backendName)
}
