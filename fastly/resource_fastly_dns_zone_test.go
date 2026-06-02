package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/dns/v1/dnszones"
)

func TestAccFastlyDNSZone_Basic(t *testing.T) {
	zoneName := fmt.Sprintf("%s.fastly-example.com.", acctest.RandString(10))
	createZone := dnszones.Zone{
		Name:        gofastly.ToPointer(zoneName),
		Description: gofastly.ToPointer("initial description"),
	}
	updateZone := dnszones.Zone{
		Name:        gofastly.ToPointer(zoneName),
		Description: gofastly.ToPointer("updated description"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDNSZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneConfig(createZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyDNSZoneRemoteState(createZone),
				),
			},
			{
				Config: testAccDNSZoneConfig(updateZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyDNSZoneRemoteState(updateZone),
				),
			},
			{
				ResourceName:      "fastly_dns_zone.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyDNSZone_WithXfrConfig(t *testing.T) {
	zoneName := fmt.Sprintf("%s.fastly-example.com.", acctest.RandString(10))
	createZone := dnszones.Zone{
		Name:        gofastly.ToPointer(zoneName),
		Description: gofastly.ToPointer("zone with xfr config"),
		XfrConfigInbound: &dnszones.XfrConfigInbound{
			Primaries: []dnszones.Primary{
				{
					Address:     gofastly.ToPointer("1.2.3.4"),
					Description: gofastly.ToPointer("primary server"),
				},
			},
		},
	}
	updateZone := dnszones.Zone{
		Name:        gofastly.ToPointer(zoneName),
		Description: gofastly.ToPointer("zone with updated xfr config"),
		XfrConfigInbound: &dnszones.XfrConfigInbound{
			Primaries: []dnszones.Primary{
				{
					Address:     gofastly.ToPointer("5.6.7.8"),
					Description: gofastly.ToPointer("updated primary server"),
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDNSZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneWithXfrConfig(createZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyDNSZoneRemoteState(createZone),
				),
			},
			{
				Config: testAccDNSZoneWithXfrConfig(updateZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyDNSZoneRemoteState(updateZone),
				),
			},
			{
				ResourceName:      "fastly_dns_zone.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckFastlyDNSZoneRemoteState(expected dnszones.Zone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["fastly_dns_zone.foo"]
		if !ok {
			return fmt.Errorf("resource not found: fastly_dns_zone.foo")
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		got, err := dnszones.Get(context.TODO(), conn, &dnszones.GetInput{
			ZoneID: gofastly.ToPointer(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("error fetching DNS zone (%s): %s", rs.Primary.ID, err)
		}

		if gofastly.ToValue(expected.Name) != gofastly.ToValue(got.Name) {
			return fmt.Errorf("bad name, expected (%s), got (%s)", gofastly.ToValue(expected.Name), gofastly.ToValue(got.Name))
		}
		if gofastly.ToValue(expected.Description) != gofastly.ToValue(got.Description) {
			return fmt.Errorf("bad description, expected (%s), got (%s)", gofastly.ToValue(expected.Description), gofastly.ToValue(got.Description))
		}

		if expected.XfrConfigInbound != nil {
			if got.XfrConfigInbound == nil {
				return fmt.Errorf("expected xfr_config_inbound to be set, got nil")
			}
			expectedPrimaries := expected.XfrConfigInbound.Primaries
			gotPrimaries := got.XfrConfigInbound.Primaries
			if len(expectedPrimaries) != len(gotPrimaries) {
				return fmt.Errorf("bad primaries count, expected (%d), got (%d)", len(expectedPrimaries), len(gotPrimaries))
			}
			for i := range expectedPrimaries {
				if gofastly.ToValue(expectedPrimaries[i].Address) != gofastly.ToValue(gotPrimaries[i].Address) {
					return fmt.Errorf("bad primary address at index %d, expected (%s), got (%s)", i, gofastly.ToValue(expectedPrimaries[i].Address), gofastly.ToValue(gotPrimaries[i].Address))
				}
				if gofastly.ToValue(expectedPrimaries[i].Description) != gofastly.ToValue(gotPrimaries[i].Description) {
					return fmt.Errorf("bad primary description at index %d, expected (%s), got (%s)", i, gofastly.ToValue(expectedPrimaries[i].Description), gofastly.ToValue(gotPrimaries[i].Description))
				}
			}
		}

		return nil
	}
}

func testAccCheckDNSZoneDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_dns_zone" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		_, err := dnszones.Get(context.TODO(), conn, &dnszones.GetInput{
			ZoneID: gofastly.ToPointer(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("tried deleting DNS zone (%s), but was still found", rs.Primary.ID)
		}
	}
	return nil
}

func TestAccFastlyDNSZone_ClearFields(t *testing.T) {
	zoneName := fmt.Sprintf("%s.fastly-example.com.", acctest.RandString(10))
	tsigKeyName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	createZone := dnszones.Zone{
		Name:        gofastly.ToPointer(zoneName),
		Description: gofastly.ToPointer("description to be cleared"),
		XfrConfigInbound: &dnszones.XfrConfigInbound{
			Primaries: []dnszones.Primary{
				{
					Address:     gofastly.ToPointer("1.2.3.4"),
					Description: gofastly.ToPointer("primary server"),
				},
			},
		},
	}
	// Clear description to "" and remove inbound_tsig_key_id.
	updateZone := dnszones.Zone{
		Name:        gofastly.ToPointer(zoneName),
		Description: gofastly.ToPointer(""),
		XfrConfigInbound: &dnszones.XfrConfigInbound{
			Primaries: []dnszones.Primary{
				{
					Address:     gofastly.ToPointer("1.2.3.4"),
					Description: gofastly.ToPointer("primary server"),
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDNSZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneWithTSIGConfig(createZone, tsigKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyDNSZoneRemoteState(createZone),
					resource.TestCheckResourceAttrSet("fastly_dns_zone.foo", "xfr_config_inbound.0.inbound_tsig_key_id"),
				),
			},
			// Remove the TSIG key reference from the zone before destroying the key.
			// The API rejects key deletion while it is still referenced by a zone.
			{
				Config: testAccDNSZoneWithTSIGConfigDetached(updateZone, tsigKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyDNSZoneRemoteState(updateZone),
				),
			},
			// Now the key is no longer referenced — drop it along with the zone.
			{
				Config: testAccDNSZoneWithXfrConfig(updateZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyDNSZoneRemoteState(updateZone),
				),
			},
			{
				ResourceName:      "fastly_dns_zone.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDNSZoneConfig(zone dnszones.Zone) string {
	return fmt.Sprintf(`
resource "fastly_dns_zone" "foo" {
  name        = "%s"
  description = "%s"
}`, gofastly.ToValue(zone.Name), gofastly.ToValue(zone.Description))
}

func testAccDNSZoneWithXfrConfig(zone dnszones.Zone) string {
	primary := zone.XfrConfigInbound.Primaries[0]
	return fmt.Sprintf(`
resource "fastly_dns_zone" "foo" {
  name        = "%s"
  description = "%s"

  xfr_config_inbound {
    primaries {
      address     = "%s"
      description = "%s"
    }
  }
}`, gofastly.ToValue(zone.Name), gofastly.ToValue(zone.Description), gofastly.ToValue(primary.Address), gofastly.ToValue(primary.Description))
}

func testAccDNSZoneWithTSIGConfigDetached(zone dnszones.Zone, tsigKeyName string) string {
	primary := zone.XfrConfigInbound.Primaries[0]
	return fmt.Sprintf(`
resource "fastly_tsig_key" "tsig" {
  name      = "%s"
  algorithm = "hmac-sha256"
  secret    = "dGVzdHNlY3JldA=="
}

resource "fastly_dns_zone" "foo" {
  name        = "%s"
  description = "%s"

  xfr_config_inbound {
    primaries {
      address     = "%s"
      description = "%s"
    }
  }
}`, tsigKeyName, gofastly.ToValue(zone.Name), gofastly.ToValue(zone.Description), gofastly.ToValue(primary.Address), gofastly.ToValue(primary.Description))
}

func testAccDNSZoneWithTSIGConfig(zone dnszones.Zone, tsigKeyName string) string {
	primary := zone.XfrConfigInbound.Primaries[0]
	return fmt.Sprintf(`
resource "fastly_tsig_key" "tsig" {
  name      = "%s"
  algorithm = "hmac-sha256"
  secret    = "dGVzdHNlY3JldA=="
}

resource "fastly_dns_zone" "foo" {
  name        = "%s"
  description = "%s"

  xfr_config_inbound {
    inbound_tsig_key_id = fastly_tsig_key.tsig.id

    primaries {
      address     = "%s"
      description = "%s"
    }
  }
}`, tsigKeyName, gofastly.ToValue(zone.Name), gofastly.ToValue(zone.Description), gofastly.ToValue(primary.Address), gofastly.ToValue(primary.Description))
}
