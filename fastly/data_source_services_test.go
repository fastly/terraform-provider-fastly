package fastly

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyDataSourceServices_Config(t *testing.T) {
	resourceName := "data.fastly_services.some"
	serviceName := "fastly_service_vcl.example_service"

	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceServicesConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ids.*", serviceName, "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "details.*", map[string]string{
						"name":    "example_service",
						"comment": "example_comment",
						"type":    "vcl",
					}),
				),
			},
		},
	})
}

func testAccFastlyDataSourceServicesConfig() string {
	tf := `
resource "fastly_service_vcl" "example_service" {
	name    = "example_service"
	comment = "example_comment"

	domain {
		name = "%s.com"
	}

	force_destroy = true
}

data "fastly_services" "some" {
	depends_on = [ fastly_service_vcl.example_service ]
}
`

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf(tf, hex.EncodeToString(b))
}
