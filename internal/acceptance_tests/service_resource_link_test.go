package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceResourceLink_ACL(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceComputeWithACLResourceLink(serviceName, aclName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					resource.TestCheckResourceAttr("fastly_acl.acl", "name", aclName),
					resource.TestCheckResourceAttrSet("fastly_acl.acl", "id"),
					resource.TestCheckResourceAttr("fastly_service_resource_link.test", "name", linkName),
					resource.TestCheckResourceAttr("fastly_service_resource_link.test", "version", "1"),
					resource.TestCheckResourceAttrPair("fastly_service_resource_link.test", "resource_id", "fastly_acl.acl", "id"),
					resource.TestCheckResourceAttrSet("fastly_service_resource_link.test", "link_id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceResourceLink_ACLRename(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))
	linkNameRenamed := fmt.Sprintf("tf_test_link_renamed_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceComputeWithACLResourceLink(serviceName, aclName, linkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_resource_link.test", "name", linkName),
					resource.TestCheckResourceAttr("fastly_service_resource_link.test", "version", "1"),
				),
			},
			{
				// Renaming the alias is applied in place via UpdateResource; the underlying
				// ACL and the same service version are untouched.
				Config: ConfigServiceComputeWithACLResourceLink(serviceName, aclName, linkNameRenamed),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_resource_link.test", "name", linkNameRenamed),
					resource.TestCheckResourceAttr("fastly_service_resource_link.test", "version", "1"),
					resource.TestCheckResourceAttrPair("fastly_service_resource_link.test", "resource_id", "fastly_acl.acl", "id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceResourceLink_ACLRetarget(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	aclNameOther := fmt.Sprintf("tf_test_acl_other_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceComputeWithACLResourceLink(serviceName, aclName, linkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_acl.acl", "name", aclName),
					resource.TestCheckResourceAttrPair("fastly_service_resource_link.test", "resource_id", "fastly_acl.acl", "id"),
				),
			},
			{
				// Pointing the link at a different ACL forces replacement of the link
				// (resource_id is RequiresReplace); the original ACL is then destroyed
				// since nothing else references it.
				Config: ConfigServiceComputeWithACLResourceLink(serviceName, aclNameOther, linkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_acl.acl", "name", aclNameOther),
					resource.TestCheckResourceAttr("fastly_service_resource_link.test", "name", linkName),
					resource.TestCheckResourceAttrPair("fastly_service_resource_link.test", "resource_id", "fastly_acl.acl", "id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceResourceLink_ACLImport(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	var serviceID, version string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceComputeWithACLResourceLink(serviceName, aclName, linkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("fastly_service_resource_link.test", "link_id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_resource_link.test"]
						if !ok {
							return fmt.Errorf("resource link resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						version = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_resource_link.test",
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, version, linkName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
