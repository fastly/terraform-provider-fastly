package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

func TestFastlyACLV1(t *testing.T) {
	var acl gofastly.ACL
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf_test_%s", acctest.RandString(10))
	serviceDomain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckACLV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testACLV1Config(serviceName, serviceDomain, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckACLV1Exists("fastly_service_v1.foo", "fastly_acl_v1.foo", &service, &acl, true),
					resource.TestCheckResourceAttr(
						"fastly_acl_v1.foo", "name", aclName),
					resource.TestCheckResourceAttr(
						"fastly_acl_v1.foo", "activate", "true"),
				),
			},
			{
				Config: testACLV1Config(serviceName, serviceDomain, aclName+"2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckACLV1Exists("fastly_service_v1.foo", "fastly_acl_v1.foo", &service, &acl, true),
					testAccCheckACLV1NameChanged(&acl, aclName),
					resource.TestCheckResourceAttr(
						"fastly_acl_v1.foo", "name", aclName+"2"),
				),
			},
		},
	})
}

func TestFastlyACLV1NotActivated(t *testing.T) {
	var acl gofastly.ACL
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf_test_%s", acctest.RandString(10))
	serviceDomain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckACLV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testACLV1ConfigNotActivated(serviceName, serviceDomain, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckACLV1Exists("fastly_service_v1.foo", "fastly_acl_v1.foo", &service, &acl, false),
					testAccCheckACLV1NotActivated(&service, &acl),
					resource.TestCheckResourceAttr(
						"fastly_acl_v1.foo", "name", aclName),
					resource.TestCheckResourceAttr(
						"fastly_acl_v1.foo", "activate", "false"),
				),
			},
			{
				Config: testACLV1ConfigNotActivated(serviceName, serviceDomain, aclName+"2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckACLV1Exists("fastly_service_v1.foo", "fastly_acl_v1.foo", &service, &acl, false),
					testAccCheckACLV1NameChanged(&acl, aclName),
					resource.TestCheckResourceAttr(
						"fastly_acl_v1.foo", "name", aclName+"2"),
				),
			},
		},
	})
}

func testAccCheckACLV1Exists(sn, n string, service *gofastly.ServiceDetail, acl *gofastly.ACL, activated bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		serviceResource, ok := s.RootModule().Resources[sn]

		if !ok {
			return fmt.Errorf("Service not found: %s", sn)
		}

		if serviceResource.Primary.ID == "" {
			return fmt.Errorf("No Service ID is set")
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		latestService, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ID: serviceResource.Primary.ID,
		})

		if err != nil {
			return err
		}

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("ACL not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ACL ID is set")
		}

		version := latestService.ActiveVersion.Number

		if !activated {
			version = version + 1
		}

		latestACL, err := conn.GetACL(&gofastly.GetACLInput{
			Service: latestService.ID,
			Version: version,
			Name:    rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		*service = *latestService
		*acl = *latestACL

		return nil
	}
}

func testAccCheckACLV1NotActivated(service *gofastly.ServiceDetail, acl *gofastly.ACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if acl.Version == service.ActiveVersion.Number {
			return fmt.Errorf("[WARN] ACL was activated - Service latest version %d, ACL version %d", service.ActiveVersion.Number, acl.Version)
		}

		return nil
	}
}

func testAccCheckACLV1NameChanged(acl *gofastly.ACL, oldACLName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if acl.Name == oldACLName {
			return fmt.Errorf("[WARN] ACL name has not been changed")
		}

		return nil
	}
}

func testAccCheckACLV1Destroy(s *terraform.State) error {
	service, ok := s.RootModule().Resources["fastly_service_v1.foo"]

	if !ok {
		return fmt.Errorf("Service not found: %s", "fastly_service_v1.foo")
	}

	if service.Primary.ID == "" {
		return fmt.Errorf("No Service ID is set")
	}

	conn := testAccProvider.Meta().(*FastlyClient).conn
	latestService, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
		ID: service.Primary.ID,
	})

	if err != nil {
		return fmt.Errorf("[WARN] Error listing services when deleting Fastly ACL (%s): %s", service.Primary.ID, err)
	}

	_, err = conn.GetACL(&gofastly.GetACLInput{
		Service: latestService.ID,
		Version: latestService.ActiveVersion.Number,
		Name:    "fastly_acl_v1.foo",
	})

	if err != nil {
		return nil
	}

	return fmt.Errorf("[WARN] Tried deleting ACL (%s), but it was still found: %s", "fastly_acl_v1.foo", err)
}

func testACLV1Config(serviceName, serviceDomain, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "127.0.0.1"
    name    = "amazon docs"
	}

	force_destroy = true
}

resource "fastly_acl_v1" "foo" {
  name       = "%s"
  service_id = "${fastly_service_v1.foo.id}"
}`, serviceName, serviceDomain, aclName)
}

func testACLV1ConfigNotActivated(serviceName, serviceDomain, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "127.0.0.1"
    name    = "amazon docs"
	}

	force_destroy = true
}

resource "fastly_acl_v1" "foo" {
  name       = "%s"
  service_id = "${fastly_service_v1.foo.id}"
  activate   = false
}`, serviceName, serviceDomain, aclName)
}
