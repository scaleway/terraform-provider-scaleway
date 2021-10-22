package scaleway

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_lb", &resource.Sweeper{
		Name: "scaleway_lb",
		F:    testSweepLB,
	})
}

func testSweepLB(_ string) error {
	return sweepZones([]scw.Zone{scw.ZoneFrPar1, scw.ZoneNlAms1, scw.ZonePlWaw1}, func(scwClient *scw.Client, zone scw.Zone) error {
		lbAPI := lb.NewZonedAPI(scwClient)

		l.Debugf("sweeper: destroying the lbs in (%s)", zone)
		listLBs, err := lbAPI.ListLBs(&lb.ZonedAPIListLBsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing lbs in (%s) in sweeper: %s", zone, err)
		}

		for _, l := range listLBs.LBs {
			err := lbAPI.DeleteLB(&lb.ZonedAPIDeleteLBRequest{
				LBID:      l.ID,
				ReleaseIP: true,
				Zone:      zone,
			})
			if err != nil {
				return fmt.Errorf("error deleting lb in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayLbLb_WithIP(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}

					resource scaleway_lb lb01 {
					    ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "LB-S"
						release_ip = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbExists(tt, "scaleway_lb.lb01"),
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "name", "test-lb"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "release_ip", "true"),
					testCheckResourceAttrUUID("scaleway_lb.lb01", "ip_id"),
					testCheckResourceAttrIPv4("scaleway_lb.lb01", "ip_address"),
					resource.TestCheckResourceAttrPair("scaleway_lb.lb01", "ip_id", "scaleway_lb_ip.ip01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
				),
			},
		},
	})
}

func testAccCheckScalewayLbExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetLB(&lb.ZonedAPIGetLBRequest{
			LBID: ID,
			Zone: zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb" {
				continue
			}

			lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = lbAPI.GetLB(&lb.ZonedAPIGetLBRequest{
				Zone: zone,
				LBID: ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("load Balancer (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func TestLbUpgradeV1SchemaUpgradeFunc(t *testing.T) {
	v0Schema := map[string]interface{}{
		"id": "fr-par/22c61530-834c-4ab4-aa71-aaaa2ac9d45a",
	}
	v1Schema := map[string]interface{}{
		"id": "fr-par-1/22c61530-834c-4ab4-aa71-aaaa2ac9d45a",
	}

	actual, err := lbUpgradeV1SchemaUpgradeFunc(context.Background(), v0Schema, nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(v1Schema, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1Schema, actual)
	}
}
