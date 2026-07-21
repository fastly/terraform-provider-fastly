package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFastlyServiceComputeAuto_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

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
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "force_destroy", "true"),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "id"),

					// Prove version 1 is bootstrapped and activated
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withBackend(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithBackend(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.address", "api.example.com"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.port", "443"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.use_ssl", "true"),
					// Verify version 1 is created and activated with backend
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withResourceLink(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	kvStoreName := fmt.Sprintf("tf-test-kv-%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndKVStoreDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithKVStoreResourceLink(serviceName, domainName, kvStoreName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_kvstore.store", "name", kvStoreName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.0.name", linkName),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_kvstore.store", "id"),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "resource_link.0.link_id"),
					// Verify version 1 is created and activated with the resource link
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_resourceLinkRename(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	kvStoreName := fmt.Sprintf("tf-test-kv-%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))
	linkNameRenamed := fmt.Sprintf("tf_test_link_renamed_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndKVStoreDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithKVStoreResourceLink(serviceName, domainName, kvStoreName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.0.name", linkName),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_kvstore.store", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
			{
				// Renaming the alias (same resource_id) is applied in place via UpdateResource,
				// but still requires a new service version like any other nested config change.
				Config: ConfigComputeAutoWithKVStoreResourceLink(serviceName, domainName, kvStoreName, linkNameRenamed),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.0.name", linkNameRenamed),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_kvstore.store", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_resourceLinkRetarget(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	kvStoreName1 := fmt.Sprintf("tf-test-kv-1-%s", acctest.RandString(10))
	kvStoreName2 := fmt.Sprintf("tf-test-kv-2-%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndKVStoreDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithKVStoreResourceLinkTarget(serviceName, domainName, kvStoreName1, kvStoreName2, linkName, "kv1"),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_kvstore.kv1", "name", kvStoreName1),
					resource.TestCheckResourceAttr("fastly_kvstore.kv2", "name", kvStoreName2),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_kvstore.kv1", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				// Pointing the same alias at a different resource_id can't be done via
				// UpdateResource (it only renames), so this exercises delete-old/create-new
				// within the same reconcile pass. Both KV Stores stay declared throughout, since
				// the Fastly API rejects deleting a KV Store in the same request that
				// unlinks it.
				Config: ConfigComputeAutoWithKVStoreResourceLinkTarget(serviceName, domainName, kvStoreName1, kvStoreName2, linkName, "kv2"),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttrSet("fastly_kvstore.kv1", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.0.name", linkName),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_kvstore.kv2", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_resourceLinkRemove(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	kvStoreName := fmt.Sprintf("tf-test-kv-%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndKVStoreDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithKVStoreResourceLink(serviceName, domainName, kvStoreName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				// Removing the block entirely from an existing service deletes the link
				// in-place on a newly cloned/activated version, without touching the service.
				// The KV Store stays declared (unlinked) here rather than disappearing in this
				// same apply, since the Fastly API rejects deleting a KV Store in the same
				// request that unlinks it.
				Config: ConfigComputeAutoWithStandaloneKVStore(serviceName, domainName, kvStoreName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttrSet("fastly_kvstore.store", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
			{
				// Now that the KV Store has been unlinked and that change has settled in its
				// own apply, it can be safely deleted.
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_resourceLinkImport(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	kvStoreName := fmt.Sprintf("tf-test-kv-%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndKVStoreDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithKVStoreResourceLink(serviceName, domainName, kvStoreName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
				),
			},
			{
				ResourceName:            "fastly_service_compute_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "package", "reuse"},
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withACLResourceLink(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndACLDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithACLResourceLink(serviceName, domainName, aclName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_acl.acl", "name", aclName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.0.name", linkName),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_acl.acl", "id"),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "resource_link.0.link_id"),
					// Verify version 1 is created and activated with the resource link
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_ACLResourceLinkRename(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))
	linkNameRenamed := fmt.Sprintf("tf_test_link_renamed_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndACLDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithACLResourceLink(serviceName, domainName, aclName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.0.name", linkName),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_acl.acl", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
			{
				// Renaming the alias (same resource_id) is applied in place via UpdateResource,
				// but still requires a new service version like any other nested config change.
				Config: ConfigComputeAutoWithACLResourceLink(serviceName, domainName, aclName, linkNameRenamed),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.0.name", linkNameRenamed),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_acl.acl", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_ACLResourceLinkRetarget(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName1 := fmt.Sprintf("tf_test_acl_1_%s", acctest.RandString(10))
	aclName2 := fmt.Sprintf("tf_test_acl_2_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndACLDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithACLResourceLinkTarget(serviceName, domainName, aclName1, aclName2, linkName, "acl1"),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_acl.acl1", "name", aclName1),
					resource.TestCheckResourceAttr("fastly_acl.acl2", "name", aclName2),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_acl.acl1", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				// Pointing the same alias at a different ACL can't be done via
				// UpdateResource (it only renames), so this exercises delete-old/create-new
				// within the same reconcile pass. Both ACLs stay declared throughout, since
				// the Fastly API rejects deleting an ACL in the same request that
				// unlinks it.
				Config: ConfigComputeAutoWithACLResourceLinkTarget(serviceName, domainName, aclName1, aclName2, linkName, "acl2"),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttrSet("fastly_acl.acl1", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.0.name", linkName),
					resource.TestCheckResourceAttrPair("fastly_service_compute_auto.test", "resource_link.0.resource_id", "fastly_acl.acl2", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_ACLResourceLinkRemove(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndACLDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithACLResourceLink(serviceName, domainName, aclName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				// Removing the block entirely from an existing service deletes the link
				// in-place on a newly cloned/activated version, without touching the service.
				// The ACL stays declared (unlinked) here rather than disappearing in this same
				// apply, since the Fastly API rejects deleting an ACL in the same
				// request that unlinks it.
				Config: ConfigComputeAutoWithStandaloneACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttrSet("fastly_acl.acl", "id"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
			{
				// Now that the ACL has been unlinked and that change has settled in its own
				// apply, it can be safely deleted.
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_ACLResourceLinkImport(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	linkName := fmt.Sprintf("tf_test_link_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceAndACLDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithACLResourceLink(serviceName, domainName, aclName, linkName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "resource_link.#", "1"),
				),
			},
			{
				ResourceName:            "fastly_service_compute_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "package", "reuse"},
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	serviceNameUpdated := fmt.Sprintf("tf-test-updated-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

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
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainName),
					// Initial version should be 1
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigComputeAutoBasic(serviceNameUpdated, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceNameUpdated),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainName),
					// Service name update does not create a new version (service-level attribute)
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_multipleBackends(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoMultipleBackends(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.name", "backend-primary"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.1.name", "backend-secondary"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_preservesBackendAndDomainOrder(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainBName := fmt.Sprintf("b-%s.example.com", acctest.RandString(10))
	domainAName := fmt.Sprintf("a-%s.example.com", acctest.RandString(10))
	config := ConfigComputeAutoUnsortedBackendAndDomainBlocks(serviceName, domainBName, domainAName)
	reversedConfig := ConfigComputeAutoReversedBackendAndDomainBlocks(serviceName, domainBName, domainAName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.name", "b"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.address", "b.example.com"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.1.name", "a"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.1.address", "a.example.com"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainBName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.1.name", domainAName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
			{
				Config: reversedConfig,
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.name", "a"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.address", "a.example.com"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.1.name", "b"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.1.address", "b.example.com"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainAName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.1.name", domainBName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
			{
				Config:   reversedConfig,
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_import(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
				),
			},
			{
				ResourceName:            "fastly_service_compute_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "package", "reuse"},
			},
		},
	})
}
