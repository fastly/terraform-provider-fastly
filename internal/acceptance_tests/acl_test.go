package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFastlyServiceCDNAuto_withACL(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withMultipleACLs(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName1 := fmt.Sprintf("acl_1_%s", acctest.RandString(10))
	aclName2 := fmt.Sprintf("acl_2_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithMultipleACLs(serviceName, domainName, aclName1, aclName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName1),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.1.name", aclName2),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.1.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.1.force_destroy", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withBackendAndACL(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithBackendAndACL(serviceName, domainName, backendName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withACLUpdate(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclNameUpdated),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withACL(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigComputeAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withMultipleACLs(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName1 := fmt.Sprintf("acl_1_%s", acctest.RandString(10))
	aclName2 := fmt.Sprintf("acl_2_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigComputeAutoWithMultipleACLs(serviceName, domainName, aclName1, aclName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.name", aclName1),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.1.name", aclName2),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "acl.1.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.1.force_destroy", "true"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withBackendAndACL(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigComputeAutoWithBackendAndACL(serviceName, domainName, backendName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withACLUpdate(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigComputeAutoWithACL(serviceName, domainName, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.name", aclNameUpdated),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withACLForceDestroy(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "false"),
				),
			},
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					AddACLEntry("fastly_service_cdn_auto.test"),
				),
			},
			{
				Config:      ConfigCDNAutoWithACL(serviceName, domainName, aclNameUpdated),
				ExpectError: regexp.MustCompile("cannot delete ACL"),
			},
			{
				Config: ConfigCDNAutoWithACLForceDestroy(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "true"),
				),
			},
			{
				Config: ConfigCDNAutoWithACLForceDestroy(serviceName, domainName, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclNameUpdated),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "true"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withACLForceDestroy(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.force_destroy", "false"),
				),
			},
			{
				Config: ConfigComputeAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					AddACLEntry("fastly_service_compute_auto.test"),
				),
			},
			{
				Config:      ConfigComputeAutoWithACL(serviceName, domainName, aclNameUpdated),
				ExpectError: regexp.MustCompile("cannot delete ACL"),
			},
			{
				Config: ConfigComputeAutoWithACLForceDestroy(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.force_destroy", "true"),
				),
			},
			{
				Config: ConfigComputeAutoWithACLForceDestroy(serviceName, domainName, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.name", aclNameUpdated),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "acl.0.force_destroy", "true"),
				),
			},
		},
	})
}
