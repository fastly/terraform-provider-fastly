package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFastlyServiceCompute_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceComputeBasic(serviceName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					resource.TestCheckResourceAttr("fastly_service_compute.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute.test", "comment", ""),
					resource.TestCheckResourceAttr("fastly_service_compute.test", "force_destroy", "true"),
					resource.TestCheckResourceAttr("fastly_service_compute.test", "reuse", "false"),
					resource.TestCheckResourceAttrSet("fastly_service_compute.test", "id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCompute_withComment(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceComputeWithComment(serviceName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					resource.TestCheckResourceAttr("fastly_service_compute.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute.test", "comment", "Managed by Terraform"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCompute_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	serviceNameUpdated := fmt.Sprintf("tf-test-updated-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceComputeBasic(serviceName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					resource.TestCheckResourceAttr("fastly_service_compute.test", "name", serviceName),
				),
			},
			{
				Config: ConfigServiceComputeBasic(serviceNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					resource.TestCheckResourceAttr("fastly_service_compute.test", "name", serviceNameUpdated),
				),
			},
		},
	})
}

func TestAccFastlyServiceCompute_import(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigServiceComputeBasic(serviceName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
				),
			},
			{
				ResourceName:            "fastly_service_compute.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse"},
			},
		},
	})
}
