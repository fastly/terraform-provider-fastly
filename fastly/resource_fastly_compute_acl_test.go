package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/fastly/go-fastly/v12/fastly/computeacls"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

func TestAccFastlyComputeACL_validate(t *testing.T) {
	var (
		acl     computeacls.ComputeACL
		service gofastly.ServiceDetail
	)
	aclName := fmt.Sprintf("Compute ACL %s", acctest.RandString(10))
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	linkName := fmt.Sprintf("resource_link_%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeACLConfig(aclName, serviceName, domainName, linkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.example", &service),
					testAccCheckFastlyComputeACLRemoteState(&service, &acl, serviceName, aclName, linkName),
				),
			},
			{
				ResourceName:      "fastly_compute_acl.example",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "package.0.filename", "imported", "stage"},
			},
			{
				Config: testAccComputeACLConfigDeleteStep1(aclName, serviceName, domainName),
			},
			{
				Config: testAccComputeACLConfigDeleteStep2(serviceName, domainName),
			},
		},
	})
}

func testAccCheckFastlyComputeACLRemoteState(service *gofastly.ServiceDetail, computeACL *computeacls.ComputeACL, serviceName, aclName, linkName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != serviceName {
			return fmt.Errorf("bad name, expected (%s), got (%s)", serviceName, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		ctx := context.TODO()

		acls, err := computeacls.ListACLs(ctx, conn)
		if err != nil {
			return fmt.Errorf("error listing all Compute ACLs: %s", err)
		}

		var found bool
		for _, acl := range acls.Data {
			if acl.Name == aclName {
				found = true
				*computeACL = acl
				break
			}
		}
		if !found {
			return fmt.Errorf("error looking up the Compute ACL")
		}

		links, err := conn.ListResources(ctx, &gofastly.ListResourcesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.Version.Number),
		})
		if err != nil {
			return fmt.Errorf("error listing all resource links: %s", err)
		}

		found = false
		for _, link := range links {
			if gofastly.ToValue(link.Name) == linkName {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("error looking up the resource link")
		}

		return nil
	}
}

func testAccComputeACLConfig(aclName, serviceName, domainName, linkName string) string {
	return fmt.Sprintf(`
resource "fastly_compute_acl" "example" {
  name = "%s"
}

resource "fastly_service_compute" "example" {
  name = "%s"

  domain {
    name = "%s"
  }

  package {
    filename         = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  resource_link {
    name        = "%s"
    resource_id = fastly_compute_acl.example.id
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}
`, aclName, serviceName, domainName, linkName)
}

func testAccComputeACLConfigDeleteStep1(aclName, serviceName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_compute_acl" "example" {
  name = "%s"
}

resource "fastly_service_compute" "example" {
  name = "%s"

  domain {
    name = "%s"
  }

  package {
    filename         = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}
`, aclName, serviceName, domainName)
}

func testAccComputeACLConfigDeleteStep2(serviceName, domainName string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "example" {
  name = "%s"

  domain {
    name = "%s"
  }

  package {
    filename         = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}

data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}
`, serviceName, domainName)
}
