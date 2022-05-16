package scaleway

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
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
		lbAPI := lbSDK.NewZonedAPI(scwClient)

		l.Debugf("sweeper: destroying the lbs in (%s)", zone)
		listLBs, err := lbAPI.ListLBs(&lbSDK.ZonedAPIListLBsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing lbs in (%s) in sweeper: %s", zone, err)
		}

		for _, l := range listLBs.LBs {
			retryInterval := defaultWaitLBRetryInterval

			if DefaultWaitRetryInterval != nil {
				retryInterval = *DefaultWaitRetryInterval
			}

			_, err := lbAPI.WaitForLbInstances(&lbSDK.ZonedAPIWaitForLBInstancesRequest{
				Zone:          zone,
				LBID:          l.ID,
				Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
				RetryInterval: &retryInterval,
			}, scw.WithContext(context.Background()))
			if err != nil {
				return fmt.Errorf("error waiting for lb in sweeper: %s", err)
			}
			err = lbAPI.DeleteLB(&lbSDK.ZonedAPIDeleteLBRequest{
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

					resource scaleway_vpc_private_network pnLB01 {
						name = "pn-with-lb-static"
					}

					resource scaleway_lb lb01 {
					    ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-with-pn-static-2"
						type = "LB-S"
						release_ip = false
						private_network {
							private_network_id = scaleway_vpc_private_network.pnLB01.id
							static_config = ["172.16.0.100", "172.16.0.101"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbExists(tt, "scaleway_lb.lb01"),
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.0", "172.16.0.100"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.1", "172.16.0.101"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}

					resource scaleway_vpc_private_network pnLB01 {
						name = "pn-with-lb-to-add"
					}

					resource scaleway_vpc_private_network pnLB02 {
						name = "pn-with-lb-to-add"
					}

					resource scaleway_lb lb01 {
					    ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-with-static-to-update-with-two-pn-3"
						type = "LB-S"
						release_ip = false
						private_network {
							private_network_id = scaleway_vpc_private_network.pnLB01.id
							static_config = ["172.16.0.100", "172.16.0.101"]
						}

						private_network {
							private_network_id = scaleway_vpc_private_network.pnLB02.id
							static_config = ["172.16.0.105", "172.16.0.106"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.#", "2"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.0", "172.16.0.100"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.1", "172.16.0.101"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.1.static_config.0", "172.16.0.105"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.1.static_config.1", "172.16.0.106"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}

					resource scaleway_vpc_private_network pnLB01 {
						name = "pn-with-lb-to-add"
					}

					resource scaleway_vpc_private_network pnLB02 {
						name = "pn-with-lb-to-add"
					}

					resource scaleway_lb lb01 {
					    ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-with-static-to-update-with-two-pn-4"
						type = "LB-S"
						release_ip = false
						private_network {
							private_network_id = scaleway_vpc_private_network.pnLB01.id
							static_config = ["172.16.0.100", "172.16.0.101"]
						}

						private_network {
							private_network_id = scaleway_vpc_private_network.pnLB02.id
							static_config = ["172.16.0.105", "172.16.0.107"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.#", "2"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.0", "172.16.0.100"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.1", "172.16.0.101"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.1.static_config.0", "172.16.0.105"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.1.static_config.1", "172.16.0.107"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}

					resource scaleway_vpc_private_network pnLB01 {
						name = "pn-with-lb-detached"
					}

					resource scaleway_vpc_private_network pnLB02 {
						name = "pn-with-lb-detached"
					}

					resource scaleway_lb lb01 {
					    ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-with-only-one-pn-is-conserved-5"
						type = "LB-S"
						release_ip = false
						private_network {
							private_network_id = scaleway_vpc_private_network.pnLB01.id
							static_config = ["172.16.0.100", "172.16.0.101"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB02", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.0.static_config.0", "172.16.0.100"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.0.static_config.1", "172.16.0.101"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}

					resource scaleway_vpc_private_network pnLB01 {
						name = "pn-with-lb-detached"
					}

					resource scaleway_vpc_private_network pnLB02 {
						name = "pn-with-lb-detached"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB02", "name"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {
					}

					resource scaleway_vpc_private_network pnLB01 {
						name = "pn-with-lb-detached"
					}

					resource scaleway_vpc_private_network pnLB02 {
						name = "pn-with-lb-detached"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB02", "name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
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

func TestAccScalewayLbLb_WithSeveralPrivateNetworks(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayInstanceServerDestroy(tt),
			testAccCheckScalewayLbDestroy(tt),
			testAccCheckScalewayLbIPDestroy(tt),
			testAccCheckScalewayVPCGatewayNetworkDestroy(tt),
			testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
			testAccCheckScalewayVPCPublicGatewayDHCPDestroy(tt),
			testAccCheckScalewayVPCPublicGatewayDestroy(tt),
			testAccCheckScalewayVPCPublicGatewayIPDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_vpc_private_network" dhcp_network {
					name = "private network with a DHCP config"
				}

				resource scaleway_vpc_private_network static_network {
					name = "private network with static config"
				}


				resource "scaleway_vpc_public_gateway_ip" "main" {
				}

				resource "scaleway_vpc_public_gateway" "main" {
					name  = "tf-test-public-gw"
					type  = "VPC-GW-S"
					ip_id = scaleway_vpc_public_gateway_ip.main.id
				}
				
				resource "scaleway_vpc_public_gateway_dhcp" "main" {
					subnet = "10.0.0.0/24"
				}
				
				### VPC Gateway Network
				resource "scaleway_vpc_gateway_network" "main" {
					gateway_id         = scaleway_vpc_public_gateway.main.id
					private_network_id = scaleway_vpc_private_network.dhcp_network.id
					dhcp_id            = scaleway_vpc_public_gateway_dhcp.main.id
					cleanup_dhcp       = true
					enable_masquerade  = true
				}
				
				### Scaleway Instance
				resource "scaleway_instance_server" "main" {
					name        = "Scaleway Terraform Provider"
					type        = "DEV1-S"
					image       = "debian_bullseye"
					enable_ipv6 = false
				
					private_network {
						pn_id = scaleway_vpc_private_network.dhcp_network.id
					}

					depends_on = [scaleway_vpc_private_network.dhcp_network]
				}

				resource scaleway_lb_ip ip01 {
				}

				resource scaleway_lb lb01 {
					ip_id = scaleway_lb_ip.ip01.id
					name = "test-lb-with-private-network-configs"
					type = "LB-S"
				
					private_network {
						private_network_id = scaleway_vpc_private_network.static_network.id
						static_config = ["172.16.0.100", "172.16.0.101"]
					}
				
					private_network {
						private_network_id = scaleway_vpc_private_network.dhcp_network.id
						dhcp_config = true
					}
				
					depends_on = [scaleway_vpc_public_gateway.main]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbExists(tt, "scaleway_lb.lb01"),
					testAccCheckScalewayLbIPExists(tt, "scaleway_lb_ip.ip01"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.dhcp_network", "name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.static_network", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.0", "172.16.0.100"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.1", "172.16.0.101"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.status", lbSDK.PrivateNetworkStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.1.dhcp_config", "true"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.1.status", lbSDK.PrivateNetworkStatusReady.String()),
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

		_, err = lbAPI.GetLB(&lbSDK.ZonedAPIGetLBRequest{
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

			_, err = lbAPI.GetLB(&lbSDK.ZonedAPIGetLBRequest{
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
