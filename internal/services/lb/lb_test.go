package lb_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb"
	lbchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/lb/testfuncs"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
	vpcgwchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpcgw/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func init() {
	resource.AddTestSweepers("scaleway_lb", &resource.Sweeper{
		Name: "scaleway_lb",
		F:    testSweepLB,
	})
}

func testSweepLB(_ string) error {
	return acctest.SweepZones([]scw.Zone{scw.ZoneFrPar1, scw.ZoneNlAms1, scw.ZonePlWaw1}, func(scwClient *scw.Client, zone scw.Zone) error {
		lbAPI := lbSDK.NewZonedAPI(scwClient)

		logging.L.Debugf("sweeper: destroying the lbs in (%s)", zone)
		listLBs, err := lbAPI.ListLBs(&lbSDK.ZonedAPIListLBsRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing lbs in (%s) in sweeper: %s", zone, err)
		}

		for _, l := range listLBs.LBs {
			retryInterval := lb.DefaultWaitLBRetryInterval

			if transport.DefaultWaitRetryInterval != nil {
				retryInterval = *transport.DefaultWaitRetryInterval
			}

			_, err := lbAPI.WaitForLbInstances(&lbSDK.ZonedAPIWaitForLBInstancesRequest{
				Zone:          zone,
				LBID:          l.ID,
				Timeout:       scw.TimeDurationPtr(instance.DefaultInstanceServerWaitTimeout),
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

func TestAccLbLb_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip main {
					}

					resource scaleway_lb main {
						ip_id = scaleway_lb_ip.main.id
						name = "test-lb-basic"
						description = "a description"
						type = "LB-S"
						tags = ["basic"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					acctest.CheckResourceAttrUUID("scaleway_lb.main", "id"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-basic"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "type", "LB-S"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "description", "a description"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "ssl_compatibility_level", lbSDK.SSLCompatibilityLevelSslCompatibilityLevelIntermediate.String()),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip main {
					}

					resource scaleway_lb main {
						ip_id = scaleway_lb_ip.main.id
						name = "test-lb-rename"
						description = "another description"
						type = "LB-S"
						tags = ["basic", "tag2"]
						ssl_compatibility_level = "ssl_compatibility_level_modern"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					acctest.CheckResourceAttrUUID("scaleway_lb.main", "id"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-rename"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "description", "another description"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "ssl_compatibility_level", lbSDK.SSLCompatibilityLevelSslCompatibilityLevelModern.String()),
				),
			},
		},
	})
}

func TestAccLbLb_Private(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb main {
						name = "test-lb-basic"
						description = "a description"
						type = "LB-S"
						tags = ["basic"]
						assign_flexible_ip = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					acctest.CheckResourceAttrUUID("scaleway_lb.main", "id"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "assign_flexible_ip", "false"),
					resource.TestCheckNoResourceAttr("scaleway_lb.main", "ip_address"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-basic"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "type", "LB-S"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "tags.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "description", "a description"),
				),
			},
			{
				Config: `
					resource scaleway_lb main {
						name = "test-lb-rename"
						description = "another description"
						type = "LB-S"
						tags = ["basic", "tag2"]
						assign_flexible_ip = false
						ssl_compatibility_level = "ssl_compatibility_level_modern"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					acctest.CheckResourceAttrUUID("scaleway_lb.main", "id"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "assign_flexible_ip", "false"),
					resource.TestCheckNoResourceAttr("scaleway_lb.main", "ip_address"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-rename"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "description", "another description"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "ssl_compatibility_level", lbSDK.SSLCompatibilityLevelSslCompatibilityLevelModern.String()),
				),
			},
		},
	})
}

func TestAccLbLb_Migrate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	lbID := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					### IP for LB IP
					resource scaleway_lb_ip main {
					}

					resource scaleway_lb main {
						ip_id = scaleway_lb_ip.main.id
						name = "test-lb-migration"
						type = "LB-S"
						tags = ["basic", "tag2"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["scaleway_lb.main"]
						if !ok {
							return fmt.Errorf("resource not found: %s", "scaleway_lb.main")
						}
						lbID = rs.Primary.ID
						return nil
					},
					resource.TestCheckResourceAttr("scaleway_lb.main", "type", "LB-S"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-migration"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "tags.#", "2"),
				),
			},
			{
				Config: `
					### IP for LB IP
					resource scaleway_lb_ip main {
					}

					resource scaleway_lb main {
						ip_id = scaleway_lb_ip.main.id
						name = "test-lb-migrate-lb-gp-m"
						type = "LB-GP-M"
						tags = ["migration"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["scaleway_lb.main"]
						if !ok {
							return fmt.Errorf("resource not found: %s", "scaleway_lb.main")
						}
						if rs.Primary.ID != lbID {
							return errors.New("LB id has changed")
						}
						return nil
					},
					resource.TestCheckResourceAttr("scaleway_lb.main", "type", "LB-GP-M"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-migrate-lb-gp-m"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "tags.#", "1"),
				),
			},
			{
				Config: `
					### IP for LB IP
					resource scaleway_lb_ip main {
					}

					resource scaleway_lb main {
						ip_id = scaleway_lb_ip.main.id
						name = "test-lb-migrate-lb-gp-m"
						type = "LB-GP-M"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "type", "LB-GP-M"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-migrate-lb-gp-m"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "tags.#", "0"),
				),
			},
			{
				Config: `
					### IP for LB IP
					resource scaleway_lb_ip main {
					}

					resource scaleway_lb main {
						ip_id = scaleway_lb_ip.main.id
						name = "test-lb-migrate-down"
						type = "LB-S"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "type", "LB-S"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-migrate-down"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "tags.#", "0"),
				),
			},
		},
	})
}

func TestAccLbLb_WithIP(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckLbDestroy(tt),
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
							static_config = ["172.16.0.100"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.lb01"),
					testAccCheckLbIPExists(tt, "scaleway_lb_ip.ip01"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.0", "172.16.0.100"),
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
							static_config = ["172.16.0.100"]
						}

						private_network {
							private_network_id = scaleway_vpc_private_network.pnLB02.id
							static_config = ["172.16.0.105"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_lb.lb01", "private_network.*", map[string]string{
						"static_config.0": "172.16.0.100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_lb.lb01", "private_network.*", map[string]string{
						"static_config.0": "172.16.0.105",
					}),
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
							static_config = ["172.16.0.100"]
						}

						private_network {
							private_network_id = scaleway_vpc_private_network.pnLB02.id
							static_config = ["172.16.0.107"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_lb.lb01", "private_network.*", map[string]string{
						"static_config.0": "172.16.0.100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_lb.lb01", "private_network.*", map[string]string{
						"static_config.0": "172.16.0.107",
					}),
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
							static_config = ["172.16.0.100"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB01", "name"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pnLB02", "name"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.#", "1"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01", "private_network.0.static_config.0", "172.16.0.100"),
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
					testAccCheckLbIPExists(tt, "scaleway_lb_ip.ip01"),
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
					testAccCheckLbIPExists(tt, "scaleway_lb_ip.ip01"),
				),
			},
		},
	})
}

func TestAccLbLb_WithStaticIPCIDR(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb_ip" "ip01" {}

					resource "scaleway_vpc_private_network" "pn" {
						name = "pn-with-lb-static"
					}

					resource "scaleway_lb" "lb01" {
					    ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-with-pn-static-cidr"
						type = "LB-S"
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
							static_config = ["192.168.1.1/25"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.lb01"),
					testAccCheckLbIPExists(tt, "scaleway_lb_ip.ip01"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.static_config.0", "192.168.1.1/25"),
				),
			},
		},
	})
}

func TestAccLbLb_InvalidStaticConfig(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb_ip" "ip01" {}

					resource "scaleway_vpc_private_network" "pn" {
					  name = "pn-with-lb-to-static"
					}

					resource "scaleway_lb" "lb01" {
					  ip_id      = scaleway_lb_ip.ip01.id
					  name       = "test-lb-with-invalid_ip"
					  type       = "LB-S"
					  private_network {
						private_network_id = scaleway_vpc_private_network.pn.id
						static_config      = ["472.16.0.100/24"]
					  }
					}`,
				ExpectError: regexp.MustCompile("\".+\" is not a valid IP address or CIDR notation: .+"),
			},
		},
	})
}

func TestAccLbLb_WithPrivateNetworksOnDHCPConfig(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			instancechecks.IsServerDestroyed(tt),
			testAccCheckLbDestroy(tt),
			lbchecks.IsIPDestroyed(tt),
			vpcgwchecks.IsGatewayNetworkDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
			vpcgwchecks.IsDHCPDestroyed(tt),
			vpcgwchecks.IsGatewayDestroyed(tt),
			vpcgwchecks.IsIPDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
				### IP for Public Gateway
				resource "scaleway_vpc_public_gateway_ip" "main" {
				}

				### The Public Gateway with the Attached IP
				resource "scaleway_vpc_public_gateway" "main" {
					name  = "tf-test-public-gw"
					type  = "VPC-GW-S"
					ip_id = scaleway_vpc_public_gateway_ip.main.id
				}
				
				### Scaleway Private Network
				resource "scaleway_vpc_private_network" "main" {
					name = "private network with a DHCP config"
				}
				
				### DHCP Space of VPC
				resource "scaleway_vpc_public_gateway_dhcp" "main" {
					subnet = "10.0.0.0/24"
				}
				
				### VPC Gateway Network
				resource "scaleway_vpc_gateway_network" "main" {
					gateway_id         = scaleway_vpc_public_gateway.main.id
					private_network_id = scaleway_vpc_private_network.main.id
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
						pn_id = scaleway_vpc_private_network.main.id
					}
				}

				### IP for LB IP
				resource scaleway_lb_ip ip01 {
				}
				
				resource scaleway_lb lb01 {
					ip_id = scaleway_lb_ip.ip01.id
					name = "test-lb-with-private-network-configs"
					type = "LB-S"
				
					private_network {
						private_network_id = scaleway_vpc_private_network.main.id
						dhcp_config = true
					}
				
					depends_on = [scaleway_vpc_public_gateway.main]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.lb01"),
					testAccCheckLbIPExists(tt, "scaleway_lb_ip.ip01"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.main", "name"),
					resource.TestCheckResourceAttrPair(
						"scaleway_lb.lb01", "private_network.0.private_network_id",
						"scaleway_vpc_private_network.main", "id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_instance_server.main", "private_network.0.pn_id",
						"scaleway_vpc_private_network.main", "id"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.status", lbSDK.PrivateNetworkStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.dhcp_config", "true"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.status", lbSDK.PrivateNetworkStatusReady.String()),
				),
			},
		},
	})
}

func TestAccLbLb_WithoutPNConfig(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_lb_ip" "ip01" {}

					resource "scaleway_vpc_private_network" "pn" {
						name = "pn-with-lb-static"
					}

					resource "scaleway_lb" "lb01" {
					    ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-with-pn-static-cidr"
						type = "LB-S"
						private_network {
							private_network_id = scaleway_vpc_private_network.pn.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.lb01"),
					testAccCheckLbIPExists(tt, "scaleway_lb_ip.ip01"),
					resource.TestCheckResourceAttrPair(
						"scaleway_lb.lb01", "private_network.0.private_network_id",
						"scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.status", lbSDK.PrivateNetworkStatusReady.String()),
					resource.TestCheckResourceAttr("scaleway_lb.lb01",
						"private_network.0.dhcp_config", "true"),
				),
			},
		},
	})
}

func TestAccLbLb_DifferentLocalityIPID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckLbDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip main {
						zone = "fr-par-2"
					}

					resource scaleway_lb main {
						ip_id = scaleway_lb_ip.main.id
						name = "test-lb-basic"
						type = "LB-S"
						zone = "fr-par-1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbExists(tt, "scaleway_lb.main"),
					acctest.CheckResourceAttrUUID("scaleway_lb.main", "id"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "name", "test-lb-basic"),
					resource.TestCheckResourceAttr("scaleway_lb.main", "type", "LB-S"),
				),
				ExpectError: regexp.MustCompile("has different locality than the resource"),
			},
			{
				Config: `
					resource scaleway_lb_ip main {
						zone = "fr-par-2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLbIPExists(tt, "scaleway_lb_ip.main"),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip main {
						zone = "fr-par-2"
					}

					resource scaleway_lb main {
						ip_id = scaleway_lb_ip.main.id
						name = "test-lb-basic"
						type = "LB-S"
						zone = "fr-par-1"
					}
				`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("has different locality than the resource"),
			},
		},
	})
}

func testAccCheckLbExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, zone, ID, err := lb.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
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

func testAccCheckLbDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb" {
				continue
			}

			lbAPI, zone, ID, err := lb.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
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
			if !httperrors.Is404(err) {
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

	actual, err := lb.UpgradeStateV1Func(context.Background(), v0Schema, nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(v1Schema, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1Schema, actual)
	}
}
