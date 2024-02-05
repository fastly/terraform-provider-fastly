package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenSumologic(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Sumologic
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Sumologic{
				{
					Name:              gofastly.ToPointer("sumo collector"),
					URL:               gofastly.ToPointer("https://collectors.sumologic.com/receiver/1"),
					Format:            gofastly.ToPointer("log format"),
					FormatVersion:     gofastly.ToPointer(2),
					MessageType:       gofastly.ToPointer("classic"),
					ResponseCondition: gofastly.ToPointer("condition 1"),
				},
			},
			local: []map[string]any{
				{
					"name":               "sumo collector",
					"url":                "https://collectors.sumologic.com/receiver/1",
					"format":             "log format",
					"format_version":     2,
					"message_type":       "classic",
					"response_condition": "condition 1",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSumologics(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_sumologic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	s := gofastly.Sumologic{
		Name:          gofastly.ToPointer("sumologger"),
		URL:           gofastly.ToPointer("https://collectors.sumologic.com/receiver/1"),
		FormatVersion: gofastly.ToPointer(2),
		Format:        gofastly.ToPointer("my format"),
	}

	sn := gofastly.Sumologic{
		Name:          gofastly.ToPointer("sumologger"),
		URL:           gofastly.ToPointer("https://collectors.sumologic.com/receiver/1"),
		FormatVersion: gofastly.ToPointer(2),
		Format:        gofastly.ToPointer("my format new"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigSumologic(name, domainName, backendName, s),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesSumologic(&service, name, s, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_sumologic.#", "1"),
				),
			},
			{
				Config: testAccServiceVCLConfigSumologic(name, domainName, backendName, sn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesSumologic(&service, name, sn, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_sumologic.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_sumologic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	s := gofastly.Sumologic{
		Name: gofastly.ToPointer("sumologger"),
		URL:  gofastly.ToPointer("https://collectors.sumologic.com/receiver/1"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigSumologicCompute(name, domainName, backendName, s),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLAttributesSumologic(&service, name, s, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_sumologic.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLAttributesSumologic(service *gofastly.ServiceDetail, name string, sumologic gofastly.Sumologic, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		sumologicList, err := conn.ListSumologics(&gofastly.ListSumologicsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Sumologics for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(sumologicList) != 1 {
			return fmt.Errorf("sumologic missing, expected: 1, got: %d", len(sumologicList))
		}

		if gofastly.ToValue(sumologicList[0].Name) != gofastly.ToValue(sumologic.Name) {
			return fmt.Errorf("sumologic name mismatch, expected: %s, got: %#v", gofastly.ToValue(sumologic.Name), gofastly.ToValue(sumologicList[0].Name))
		}

		if serviceType == ServiceTypeVCL && gofastly.ToValue(sumologicList[0].Format) != gofastly.ToValue(sumologic.Format) {
			return fmt.Errorf("sumologic format mismatch, expected: %s, got: %#v", gofastly.ToValue(sumologic.Format), gofastly.ToValue(sumologicList[0].Format))
		}

		return nil
	}
}

func testAccServiceVCLConfigSumologicCompute(name, domainName, backendName string, sumologic gofastly.Sumologic) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  logging_sumologic {
    name = "%s"
    url = "%s"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, name, domainName, backendName, gofastly.ToValue(sumologic.Name), gofastly.ToValue(sumologic.URL))
}

func testAccServiceVCLConfigSumologic(name, domainName, backendName string, sumologic gofastly.Sumologic) string {
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

  logging_sumologic {
    name = "%s"
    url = "%s"
    format_version = %d
    format = "%s"
  }

  force_destroy = true
}`, name, domainName, backendName, gofastly.ToValue(sumologic.Name), gofastly.ToValue(sumologic.URL), gofastly.ToValue(sumologic.FormatVersion), gofastly.ToValue(sumologic.Format))
}
