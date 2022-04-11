package scaleway

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_lb_ip", &resource.Sweeper{
		Name: "scaleway_lb_ip",
		F:    testSweepLBIP,
	})
}

func testSweepLBIP(_ string) error {
	return sweepZones([]scw.Zone{scw.ZoneFrPar1, scw.ZoneNlAms1, scw.ZonePlWaw1}, func(scwClient *scw.Client, zone scw.Zone) error {
		lbAPI := lbSDK.NewZonedAPI(scwClient)

		l.Debugf("sweeper: destroying the lb ips in zone (%s)", zone)
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
					resource scaleway_lb_ip ipZone {
						zone = "nl-ams-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ipZone"),
					testCheckResourceAttrIPv4("scaleway_lb_ip.ipZone", "ip_address"),
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
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
					testCheckResourceAttrIPv4("scaleway_lb_ip.ip01", "ip_address"),
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
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
					testCheckResourceAttrIPv4("scaleway_lb_ip.ip01", "ip_address"),
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
					testAccCheckScalewayLbExists(tt, "scaleway_lb.main"),
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
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

			lbID, lbExist := rs.Primary.Attributes["lb_id"]
			if lbExist && len(lbID) > 0 {
				retryInterval := defaultWaitLBRetryInterval

				if DefaultWaitRetryInterval != nil {
					retryInterval = *DefaultWaitRetryInterval
				}

				_, err := lbAPI.WaitForLbInstances(&lbSDK.ZonedAPIWaitForLBInstancesRequest{
					Zone:          zone,
					LBID:          ID,
					Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
					RetryInterval: &retryInterval,
				}, scw.WithContext(context.Background()))

				// Unexpected api error we return it
				if !is404Error(err) {
					return err
				}
			}

			err = resource.RetryContext(context.Background(), retryLbIPInterval, func() *resource.RetryError {
				_, errGet := lbAPI.GetIP(&lbSDK.ZonedAPIGetIPRequest{
					Zone: zone,
					IPID: ID,
				})
				if is403Error(errGet) {
					return resource.RetryableError(errGet)
				}

				return resource.NonRetryableError(errGet)
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
