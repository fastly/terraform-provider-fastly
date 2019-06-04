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

var flattenSumologicTests = []struct {
	name     string
	in       []*gofastly.Sumologic
	expected []map[string]interface{}
}{
	{
		name: "basic flatten",
		in: []*gofastly.Sumologic{
			{
				Name: "sumo collector", URL: "https://sumologic.com/collector/1",
				Format: "log format", FormatVersion: 2,
				MessageType: "classic", ResponseCondition: "condition 1",
			},
		},
		expected: []map[string]interface{}{
			{
				"name": "sumo collector", "url": "https://sumologic.com/collector/1",
				"format": "log format", "format_version": 2,
				"message_type": "classic", "response_condition": "condition 1",
			},
		},
	},
}

func TestResourceFastlyFlattenSumologic(t *testing.T) {

	for _, tt := range flattenSumologicTests {
		t.Run(tt.name, func(t *testing.T) {

			actual := flattenSumologics(tt.in)

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Error matching:\nexpected: %#v\ngot: %#v", tt.expected, actual)
			}
		})
	}
}

func TestAccFastlyServiceV1_sumologic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	s := gofastly.Sumologic{
		Name:          "sumologger",
		URL:           "https://collectors.sumologic.com/receiver/1",
		FormatVersion: 2,
		Format:        "my format",
	}

	sn := gofastly.Sumologic{
		Name:          "sumologger",
		URL:           "https://collectors.sumologic.com/receiver/1",
		FormatVersion: 2,
		Format:        "my format new",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1ConfigSumologic(name, domainName, backendName, s),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1AttributesSumologic(&service, name, s),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "sumologic.#", "1"),
				),
			},
			{
				Config: testAccServiceV1ConfigSumologic(name, domainName, backendName, sn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1AttributesSumologic(&service, name, sn),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "sumologic.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1AttributesSumologic(service *gofastly.ServiceDetail, name string, sumologic gofastly.Sumologic) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		sumologicList, err := conn.ListSumologics(&gofastly.ListSumologicsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Sumologics for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(sumologicList) != 1 {
			return fmt.Errorf("Sumologic missing, expected: 1, got: %d", len(sumologicList))
		}

		if sumologicList[0].Name != sumologic.Name {
			return fmt.Errorf("Sumologic name mismatch, expected: %s, got: %#v", sumologic.Name, sumologicList[0].Name)
		}

		if sumologicList[0].Format != sumologic.Format {
			return fmt.Errorf("Sumologic format mismatch, expected: %s, got: %#v", sumologic.Format, sumologicList[0].Format)
		}

		return nil
	}
}

func testAccServiceV1ConfigSumologic(name, domainName, backendName string, sumologic gofastly.Sumologic) string {

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

  sumologic {
    name = "%s"
    url = "%s"
    format_version = %d
    format = "%s"
  }

  force_destroy = true
}`, name, domainName, backendName, sumologic.Name, sumologic.URL, sumologic.FormatVersion, sumologic.Format)
}
