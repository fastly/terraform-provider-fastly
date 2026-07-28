package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceCondition_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigConditionBasic(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "name", conditionName),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "type", "REQUEST"),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "statement", `req.url ~ "^/admin"`),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "priority", "10"),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_condition.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_condition.test", "id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCondition_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigConditionBasic(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_condition.test", "statement", `req.url ~ "^/admin"`),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "priority", "10"),
				),
			},
			{
				Config: ConfigConditionUpdated(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_condition.test", "statement", `req.url ~ "^/private"`),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "priority", "5"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCondition_typeChange(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigConditionBasic(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_condition.test", "type", "REQUEST"),
				),
			},
			{
				// The Fastly API doesn't support updating a condition's type via PUT, so the
				// resource's Update handler must delete and recreate it (see condition's
				// ops.Update). Confirm the resource still applies cleanly and the new type
				// sticks, rather than erroring or silently keeping the old type.
				Config: ConfigConditionTypeChanged(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_condition.test", "name", conditionName),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "type", "CACHE"),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "statement", "beresp.status == 200"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCondition_multiple(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	conditionName1 := fmt.Sprintf("condition-a-%s", acctest.RandString(10))
	conditionName2 := fmt.Sprintf("condition-b-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigConditionMultiple(serviceName, domainName, conditionName1, conditionName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_condition.test1", "name", conditionName1),
					resource.TestCheckResourceAttr("fastly_service_condition.test1", "type", "REQUEST"),
					resource.TestCheckResourceAttr("fastly_service_condition.test2", "name", conditionName2),
					resource.TestCheckResourceAttr("fastly_service_condition.test2", "type", "CACHE"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCondition_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigConditionForImport(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_condition.test", "name", conditionName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_condition.test"]
						if !ok {
							return fmt.Errorf("condition resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_condition.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, conditionName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
