package fastly

import (
	"fmt"
	"testing"
	"text/template"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceCompute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test1.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceComputeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputeConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "package.#", "1"),
				),
			},
			{
				ResourceName:      "fastly_service_compute.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy", "package.0.filename", "imported", "stage"},
			},
		},
	})
}

func TestAccFastlyServiceCompute_stage(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName1 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName2 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName3 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	type Config struct {
		Name     string
		Domain   string
		Backends []string
		Stage    bool
	}

	tmplText := `
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "{{ .Name }}"

  {{ if .Stage }}
  activate = false
  stage = true
  {{ end }}

  domain {
    name    = "{{ .Domain }}"
    comment = "tf-testing-domain"
  }

  {{ range .Backends }}
  backend {
    address = "{{ . }}"
    name    = "tf-test backend {{ . }}"
  }
  {{ end }}

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`

	tmpl, err := template.New("test").Parse(tmplText)
	if err != nil {
		t.Fatal(err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceComputeDestroy,
		Steps: []resource.TestStep{
			{
				Config: renderTestConfigTemplate(t, tmpl, Config{
					Name:     name,
					Domain:   domain,
					Backends: []string{backendName1},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceAttributesBackends(&service, name, []string{backendName1}, 1),
				),
			},

			{
				Config: renderTestConfigTemplate(t, tmpl, Config{
					Name:     name,
					Domain:   domain,
					Backends: []string{backendName1, backendName2},
					Stage:    true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceAttributesBackends(&service, name, []string{backendName1, backendName2}, 2),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "staged_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "backend.#", "2"),
				),
			},

			{
				Config: renderTestConfigTemplate(t, tmpl, Config{
					Name:     name,
					Domain:   domain,
					Backends: []string{backendName1, backendName2, backendName3},
					Stage:    true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceAttributesBackends(&service, name, []string{backendName1, backendName2, backendName3}, 2),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "staged_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "backend.#", "3"),
				),
			},
		},
	})
}

func testAccCheckServiceComputeDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_compute" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		l, err := conn.ListServices(&gofastly.ListServicesInput{})
		if err != nil {
			return fmt.Errorf("error listing services when deleting Fastly Service (%s): %s", rs.Primary.ID, err)
		}

		for _, s := range l {
			if gofastly.ToValue(s.ServiceID) == rs.Primary.ID {
				// service still found
				return fmt.Errorf("tried deleting Service (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccServiceComputeConfig(name, domain string) string {
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
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }
  force_destroy = true
  activate = false
}`, name, domain)
}
