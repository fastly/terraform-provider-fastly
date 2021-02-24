package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
				Config: testAccServiceV1Config_dictionary(name, dictName, backendName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, name, dictName, false),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_dictionary_write_only(t *testing.T) {
	var service gofastly.ServiceDetail
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
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, name, dictName, true),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_dictionary_update_name(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	dictName := fmt.Sprintf("dict %s", acctest.RandString(10))
	updatedDictName := fmt.Sprintf("new dict %s", acctest.RandString(10))

	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_dictionary(name, dictName, backendName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, name, dictName, false),
				),
			},
			{
				Config: testAccServiceV1Config_dictionary(name, updatedDictName, backendName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_dictionary(&service, name, updatedDictName, false),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes_dictionary(service *gofastly.ServiceDetail, name, dictName string, writeOnly bool) resource.TestCheckFunc {
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
