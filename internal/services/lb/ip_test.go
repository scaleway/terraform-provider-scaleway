package lb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb"
	lbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb/testfuncs"
)

func init() {
	resource.AddTestSweepers("scaleway_lb_ip", &resource.Sweeper{
		Name: "scaleway_lb_ip",
		F:    testSweepIP,
	})
}

func testSweepIP(_ string) error {
	return acctest.SweepZones([]scw.Zone{scw.ZoneFrPar1, scw.ZoneNlAms1, scw.ZonePlWaw1}, func(scwClient *scw.Client, zone scw.Zone) error {
		lbAPI := lbSDK.NewZonedAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the lb ips in zone (%s)", zone)
		listIPs, err := lbAPI.ListIPs(&lbSDK.ZonedAPIListIPsRequest{Zone: zone}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing lb ips in (%s) in sweeper: %s", zone, err)
		}

		for _, ip := range listIPs.IPs {
			if ip.LBID == nil {
				err := lbAPI.ReleaseIP(&lbSDK.ZonedAPIReleaseIPRequest{
					Zone: zone,
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

func TestAccIP_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      lbchecks.IsIPDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip ipZone {
						zone = "nl-ams-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isIPPresent(tt, "scaleway_lb_ip.ipZone"),
					acctest.CheckResourceAttrIPv4("scaleway_lb_ip.ipZone", "ip_address"),
					resource.TestCheckResourceAttrSet("scaleway_lb_ip.ipZone", "reverse"),
					resource.TestCheckResourceAttr("scaleway_lb_ip.ipZone", "zone", "nl-ams-1"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isIPPresent(tt, "scaleway_lb_ip.ip01"),
					acctest.CheckResourceAttrIPv4("scaleway_lb_ip.ip01", "ip_address"),
					resource.TestCheckResourceAttrSet("scaleway_lb_ip.ip01", "reverse"),
					resource.TestCheckResourceAttr("scaleway_lb_ip.ip01", "zone", "fr-par-1"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
						reverse = "myreverse.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isIPPresent(tt, "scaleway_lb_ip.ip01"),
					acctest.CheckResourceAttrIPv4("scaleway_lb_ip.ip01", "ip_address"),
					resource.TestCheckResourceAttr("scaleway_lb_ip.ip01", "reverse", "myreverse.com"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
						reverse = "myreverse.com"
					}

					resource scaleway_lb main {
					    ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-with-release-ip"
						type = "LB-S"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isLbPresent(tt, "scaleway_lb.main"),
					isIPPresent(tt, "scaleway_lb_ip.ip01"),
				),
			},
		},
	})
}

func isIPPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, zone, ID, err := lb.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetIP(&lbSDK.ZonedAPIGetIPRequest{
			IPID: ID,
			Zone: zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
