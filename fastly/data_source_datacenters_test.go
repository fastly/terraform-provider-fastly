package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceDatacenters(t *testing.T) {
	resourceName := "data.fastly_datacenters.some"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceDatacentersConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyDataSourceDatacenters(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "pops.0.code"),
					resource.TestCheckResourceAttrSet(resourceName, "pops.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "pops.0.group"),
					// NOTE: we don't validate pops.0.shield as not all pops within the
					// dataset has a shield value. We also can't rely on the data
					// staying consistent (e.g. if either the order of the data changes
					// or a pop that once reported a shield suddently stops reporting one,
					// then the test becomes flaky.
				),
			},
		},
	})
}

func testAccFastlyDataSourceDatacenters(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		var (
			popsSize int
			err      error
		)

		if popsSize, err = strconv.Atoi(a["pops.#"]); err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		datacenters, err := conn.AllDatacenters()
		if err != nil {
			return fmt.Errorf("[ERROR] error fetching datacenters: %s", err)
		}

		if popsSize != len(datacenters) {
			return fmt.Errorf("[ERROR] unexpected datacenters count (remote: %d, local: %d)", len(datacenters), popsSize)
		}

		return nil
	}
}

const testAccFastlyDataSourceDatacentersConfig = `
data "fastly_datacenters" "some" {
}
`
