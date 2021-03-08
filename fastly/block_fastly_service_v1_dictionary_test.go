package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenDictionary(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Dictionary
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Dictionary{
				{
					ID:        "1234567890",
					Name:      "dictionary-example",
					WriteOnly: false,
				},
			},
			local: []map[string]interface{}{
				{
					"dictionary_id": "1234567890",
					"name":          "dictionary-example",
					"write_only":    false,
				},
			},
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
	var dictionary gofastly.Dictionary
	name := acctest.RandomWithPrefix(testResourcePrefix)
	dictName := fmt.Sprintf("dict %s", acctest.RandString(10))
	updatedDictName := fmt.Sprintf("new dict %s", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	// Six part test:
	// 1. Create service with dictionary
	// 2. Rename the dictionary, should succeed because it is empty
	// 3. Keep dictionary the same and add an item to it
	// 4. Try to rename it, expect to fail with "dictionary not empty error"
	// 5. Without renaming, set force_destroy=true to skip the deletion check
	// 6. Try to rename again, expect to succeed
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_dictionary(name, dictName, backendName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, &dictionary, name, dictName, false),
				),
			},
			{
				Config: testAccServiceV1Config_dictionary(name, updatedDictName, backendName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, &dictionary, name, updatedDictName, false),
				),
			},
			{
				Config: testAccServiceV1Config_dictionary(name, updatedDictName, backendName, domainName),
				Check:  testAccAddDictionaryItems(&dictionary), // triggers side-effect of adding a Dictionary Item
			},
			{
				Config:      testAccServiceV1Config_dictionary(name, dictName, backendName, domainName),
				ExpectError: regexp.MustCompile("Cannot delete.*not empty.*"),
			},
			{
				Config: testAccServiceV1Config_dictionaryForceDestroy(name, updatedDictName, backendName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, &dictionary, name, updatedDictName, false),
				),
			},
			{
				Config: testAccServiceV1Config_dictionaryForceDestroy(name, dictName, backendName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, &dictionary, name, dictName, false),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_dictionary_write_only(t *testing.T) {
	var service gofastly.ServiceDetail
	var dictionary gofastly.Dictionary
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	dictName := fmt.Sprintf("dict %s", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_dictionary_write_only(name, dictName, backendName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, &dictionary, name, dictName, true),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes_dictionary(service *gofastly.ServiceDetail, dictionary *gofastly.Dictionary, name, dictName string, writeOnly bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		dict, err := conn.GetDictionary(&gofastly.GetDictionaryInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
			Name:           dictName,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Dictionary records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if dict.Name != dictName {
			return fmt.Errorf("Dictionary name mismatch, expected: %s, got: %#v", dictName, dict.Name)
		}

		if dict.WriteOnly != writeOnly {
			return fmt.Errorf("Dictionary write_only attribute mismatch, expected: %#v, got: %#v", writeOnly, dict.WriteOnly)
		}

		*dictionary = *dict

		return nil
	}
}

// testAccAddDictionaryItems doesn't technically check for anything despite returning a TestCheckFunc. Instead it is
// used for its side effect of adding a Dictionary Item
func testAccAddDictionaryItems(dictionary *gofastly.Dictionary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*FastlyClient).conn
		_, err := conn.CreateDictionaryItem(&gofastly.CreateDictionaryItemInput{
			ServiceID:    dictionary.ServiceID,
			DictionaryID: dictionary.ID,
			ItemKey:      "testKey",
			ItemValue:    "testItem",
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error adding item to dictionary (%s) on service (%s): %w", dictionary.ID, dictionary.ServiceID, err)
		}

		return nil
	}
}

func testAccServiceV1Config_dictionary(name, dictName, backendName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf-test backend"
  }

  dictionary {
    name = "%s"
  }

  force_destroy = true
}`, name, domainName, backendName, dictName)
}

func testAccServiceV1Config_dictionaryForceDestroy(name, dictName, backendName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf-test backend"
  }

  dictionary {
    name          = "%s"
    force_destroy = true
  }

  force_destroy = true
}`, name, domainName, backendName, dictName)
}

func testAccServiceV1Config_dictionary_write_only(name, dictName, backendName, domainName string) string {
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
    name       = "%s"
    write_only = true
  }

  force_destroy = true
}`, name, domainName, backendName, dictName)
}
