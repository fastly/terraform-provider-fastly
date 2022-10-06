package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenVCLs(t *testing.T) {
	cases := []struct {
		remote []*gofastly.VCL
		local  []map[string]any
	}{
		{
			remote: []*gofastly.VCL{
				{
					Name:    "myVCL",
					Content: "<<EOF somecontent EOF",
					Main:    true,
				},
			},
			local: []map[string]any{
				{
					"name":    "myVCL",
					"content": "<<EOF somecontent EOF",
					"main":    true,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenVCLs(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCLBlockBasic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLVCLConfig(name, domainName1, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLVCLAttributes(&service, name, 1),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "vcl.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLVCLConfigUpdate(name, domainName1, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLVCLAttributes(&service, name, 2),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "vcl.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLVCLAttributes(service *gofastly.ServiceDetail, name string, vclCount int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if service.Name != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		vclList, err := conn.ListVCLs(&gofastly.ListVCLsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up VCL for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(vclList) != vclCount {
			return fmt.Errorf("vcl count mismatch, expected (%d), got (%d)", vclCount, len(vclList))
		}

		return nil
	}
}

func testAccServiceVCLVCLConfig(name, domain, backendName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf-test backend"
  }

  vcl {
    name    = "my_custom_main_vcl"
    content = <<EOF
sub vcl_recv {
#FASTLY recv

    if (req.request != "HEAD" && req.request != "GET" && req.request != "FASTLYPURGE") {
      return(pass);
    }

    return(lookup);
}

backend amazondocs {
  .host = "127.0.0.1";
  .port = "80";
}
EOF
    main    = true
  }

  force_destroy = true
}`, name, domain, backendName)
}

func testAccServiceVCLVCLConfigUpdate(name, domain, backendName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf-test backend"
  }

  vcl {
    name    = "my_custom_main_vcl"
    content = <<EOF
sub vcl_recv {
#FASTLY recv

    if (req.request != "HEAD" && req.request != "GET" && req.request != "FASTLYPURGE") {
      return(pass);
    }

    return(lookup);
}

backend amazondocs {
  .host = "127.0.0.1";
  .port = "80";
}
EOF
    main    = true
  }

        vcl {
                name = "other_vcl"
                content = <<EOF
sub vcl_error {
#FASTLY error
}

backend amazondocs {
  .host = "127.0.0.1";
  .port = "80";
}
EOF
        }

  force_destroy = true
}`, name, domain, backendName)
}
