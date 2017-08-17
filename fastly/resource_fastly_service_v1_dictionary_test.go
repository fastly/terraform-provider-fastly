package fastly

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

func TestResourceFastlyFlattenDictionary(t *testing.T) {
	cases := []struct {
		remote []*Dictionary
		local  []map[string]interface{}
	}{
		{
			remote: []*Dictionary{
				&Dictionary{
					dictionary: &gofastly.Dictionary{Name: "dictionary1", ID: "dict_id_1"},
					items:      []*gofastly.DictionaryItem{&gofastly.DictionaryItem{ItemKey: "key1", ItemValue: "val1"}, &gofastly.DictionaryItem{ItemKey: "key2", ItemValue: "val2"}},
				},
			},
			local: func() []map[string]interface{} {
				var ds []map[string]interface{}

				d := map[string]interface{}{
					"name": "dictionary1",
					"id":   "dict_id_1",
				}

				items := make(map[string]interface{})
				items["key1"] = "val1"
				items["key2"] = "val2"

				d["items"] = items

				ds = append(ds, d)
				return ds
			}(),
		},
	}

	for _, c := range cases {
		out := flattenDictionaries(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1_dictionary(t *testing.T) {
	var service gofastly.ServiceDetail
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(1))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccServiceV1Config_dictionary(domainName, backendName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, domainName, dictionaryName),
				),
			},
			resource.TestStep{
				Config: testAccServiceV1Config_dictionaryItems(domainName, backendName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionaryItems(&service, domainName, dictionaryName),
				),
			},
			resource.TestStep{
				Config: testAccServiceV1Config_dictionaryItemsAdd(domainName, backendName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionaryItemsAdd(&service, domainName, dictionaryName),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes_dictionary(service *gofastly.ServiceDetail, name, dictionaryName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		dictionariesList, err := conn.ListDictionaries(&gofastly.ListDictionariesInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionaries for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(dictionariesList) != 1 {
			return fmt.Errorf("Dictionary missing, expected: 1, got: %d", len(dictionariesList))
		}

		if dictionariesList[0].Name != dictionaryName {
			return fmt.Errorf("Dictionary name mismatch, expected: %s, got: %#v", dictionaryName, dictionariesList[0].Name)
		}

		return nil
	}
}

func testAccCheckFastlyServiceV1Attributes_dictionaryItems(service *gofastly.ServiceDetail, name, dictionaryName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		dictionariesList, err := conn.ListDictionaries(&gofastly.ListDictionariesInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionaries for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		dict := dictionariesList[0]
		itemsList, err := conn.ListDictionaryItems(&gofastly.ListDictionaryItemsInput{
			Service:    service.ID,
			Dictionary: dict.ID,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionary Items for (%s), dictionary (%v): %s", service.Name, dict.ID, err)
		}

		if len(itemsList) != 2 {
			return fmt.Errorf("[ERR] Number of Dictionary Items mismatch, expected: 2, got: %d for (%s)", len(itemsList), service.Name)
		}

		return nil
	}
}

func testAccCheckFastlyServiceV1Attributes_dictionaryItemsAdd(service *gofastly.ServiceDetail, name, dictionaryName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		dictionariesList, err := conn.ListDictionaries(&gofastly.ListDictionariesInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionaries for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		dict := dictionariesList[0]
		itemsList, err := conn.ListDictionaryItems(&gofastly.ListDictionaryItemsInput{
			Service:    service.ID,
			Dictionary: dict.ID,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionary Items for (%s), dictionary (%v): %s", service.Name, dict.ID, err)
		}

		if len(itemsList) != 3 {
			return fmt.Errorf("[ERR] Number of Dictionary Items mismatch, expected: 2, got: %d for (%s)", len(itemsList), service.Name)
		}

		return nil
	}
}

func testAccServiceV1Config_dictionary(domainName, backendName, dictionaryName string) string {

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

  dictionary {
	  name = "%s"
  }

  force_destroy = true
}`, domainName, domainName, backendName, dictionaryName)
}

func testAccServiceV1Config_dictionaryItems(domainName, backendName, dictionaryName string) string {

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

  dictionary {
	  name = "%s"
	  items = {
		  key1 = "item1"
		  key2 = "item2"
	  }
  }

  force_destroy = true
}`, domainName, domainName, backendName, dictionaryName)
}

func testAccServiceV1Config_dictionaryItemsAdd(domainName, backendName, dictionaryName string) string {

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
	
	  dictionary {
		  name = "%s"
		  items = {
			  key1 = "item1"
			  key2 = "item2"
			  key3 = "item3"
		  }
	  }
	
	  force_destroy = true
	}`, domainName, domainName, backendName, dictionaryName)
}
