package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_lb_ip", &resource.Sweeper{
		Name: "scaleway_lb_ip",
		F:    testSweepLBIP,
	})
}

func testSweepLBIP(_ string) error {
	return sweepZones([]scw.Zone{scw.ZoneFrPar1, scw.ZoneNlAms1, scw.ZonePlWaw1}, func(scwClient *scw.Client, region scw.Zone) error {
		lbAPI := lb.NewZonedAPI(scwClient)

		l.Debugf("sweeper: destroying the lb ips in (%s)", region)
		listIPs, err := lbAPI.ListIPs(&lb.ZonedAPIListIPsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing lb ips in (%s) in sweeper: %s", region, err)
		}

		for _, ip := range listIPs.IPs {
			if ip.LBID == nil {
				err := lbAPI.ReleaseIP(&lb.ZonedAPIReleaseIPRequest{
					IPID: ip.ID,
				})
				if err != nil {
					return fmt.Errorf("error deleting lb ip in sweeper: %s", err)
				}
			}
		}

		return nil
	})
}

func TestAccScalewayLbIP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbIPDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
					testCheckResourceAttrIPv4("scaleway_lb_ip.ip01", "ip_address"),
					resource.TestCheckResourceAttrSet("scaleway_lb_ip.ip01", "reverse"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
						reverse = "myreverse.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
					testCheckResourceAttrIPv4("scaleway_lb_ip.ip01", "ip_address"),
					resource.TestCheckResourceAttr("scaleway_lb_ip.ip01", "reverse", "myreverse.com"),
				),
			},
		},
	})
}

func testAccCheckScalewayLbIPExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetIP(&lb.ZonedAPIGetIPRequest{
			IPID: ID,
			Zone: zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbIPDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb_ip" {
				continue
			}

			lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = lbAPI.GetIP(&lb.ZonedAPIGetIPRequest{
				Zone: zone,
				IPID: ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("IP (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
