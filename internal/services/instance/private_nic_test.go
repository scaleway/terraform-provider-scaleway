package instance_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

func TestAccPrivateNIC_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			isPrivateNICDestroyed(tt),
			vpcchecks.CheckVPCDestroy(tt),
			instancechecks.IsServerDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc main {
						name = "TestAccPrivateNIC_Basic"
					}

					resource scaleway_vpc_private_network pn {
						name = "TestAccScalewayInstancePrivateNIC_Basic"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "server" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
					}

					resource scaleway_instance_private_nic nic {
						server_id          = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.pn.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("scaleway_instance_private_nic.nic", "srn", regexp.MustCompile(`^srn://instance\..+/zones/.+/private-network-interfaces/.+$`)),
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "private_network_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "server_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "private_ips.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "private_ips.0.address"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "project_id", "scaleway_instance_server.server", "project_id"),
				),
			},
			{
				ResourceName:            "scaleway_instance_private_nic.nic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipam_ip_ids"},
			},
		},
	})
}

func TestAccPrivateNIC_Tags(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			isPrivateNICDestroyed(tt),
			vpcchecks.CheckVPCDestroy(tt),
			instancechecks.IsServerDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc main {
						name = "TestAccPrivateNIC_Tags"
					}

					resource scaleway_vpc_private_network pn {
						name = "TestAccScalewayInstancePrivateNIC_Tags"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "server" {
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						state = "stopped"
					}

					resource scaleway_instance_private_nic nic {
						server_id          = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.pn.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "private_network_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "server_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "project_id", "scaleway_instance_server.server", "project_id"),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "tags.#", "0"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "TestAccPrivateNIC_Tags"
					}

					resource scaleway_vpc_private_network pn {
						name = "TestAccScalewayInstancePrivateNIC_Tags"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "server" {
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						state = "stopped"
					}

					resource scaleway_instance_private_nic nic {
						server_id          = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.pn.id
						tags = ["tag1", "tag2"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "private_network_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "server_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "project_id", "scaleway_instance_server.server", "project_id"),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "tags.1", "tag2"),
				),
			},
			{
				ResourceName:            "scaleway_instance_private_nic.nic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipam_ip_ids"},
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "TestAccPrivateNIC_Tags"
					}

					resource scaleway_vpc_private_network pn {
						name = "TestAccScalewayInstancePrivateNIC_Tags"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "server" {
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						state = "stopped"
					}

					resource scaleway_instance_private_nic nic {
						server_id          = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.pn.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "private_network_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_private_nic.nic", "server_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "project_id", "scaleway_instance_server.server", "project_id"),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "tags.#", "0"),
				),
			},
		},
	})
}

func TestAccPrivateNIC_WithIPAM(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			isPrivateNICDestroyed(tt),
			vpcchecks.CheckVPCDestroy(tt),
			instancechecks.IsServerDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "vpc" {
					  name = "TestAccScalewayInstancePrivateNIC_IPAM"
					}
					
					resource "scaleway_vpc_private_network" "pn" {
					  name = "TestAccScalewayInstancePrivateNIC_IPAM"
					  ipv4_subnet {
						subnet = "172.16.64.0/22"
					  }
					  vpc_id = scaleway_vpc.vpc.id
					}
					
					resource "scaleway_ipam_ip" "ip" {
					  address = "172.16.64.7"
					  source {
						private_network_id = scaleway_vpc_private_network.pn.id
					  }
					}
					
					resource "scaleway_instance_server" "server" {
					  name = "TestAccScalewayInstancePrivateNIC_IPAM"
					  image = "ubuntu_focal"
					  type  = "PLAY2-MICRO"
					}
					
					resource "scaleway_instance_private_nic" "pnic" {
					  private_network_id = scaleway_vpc_private_network.pn.id
					  server_id          = scaleway_instance_server.server.id
					  ipam_ip_ids        = [scaleway_ipam_ip.ip.id]
					}     

					data "scaleway_ipam_ip" "by_id" {
					  resource {
						id   = scaleway_instance_private_nic.pnic.id
						type = "instance_private_nic"
					  }
					  type = "ipv4"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.pnic"),
					resource.TestCheckResourceAttrPair(
						"scaleway_instance_private_nic.pnic", "private_network_id",
						"scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_instance_private_nic.pnic", "ipam_ip_ids.0",
						"scaleway_ipam_ip.ip", "id"),
					resource.TestCheckResourceAttrPair(
						"scaleway_ipam_ip.ip", "address_cidr",
						"data.scaleway_ipam_ip.by_id", "address_cidr"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.pnic", "project_id", "scaleway_instance_server.server", "project_id"),
				),
			},
			{
				ResourceName:            "scaleway_instance_private_nic.pnic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipam_ip_ids"},
			},
		},
	})
}

func TestAccPrivateNIC_SecurityGroup(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	unchangedConfig := `
	    resource "scaleway_vpc" "main" {
			name = "TestAccPrivateNIC_SecurityGroup"
		}

		resource "scaleway_vpc_private_network" "pn" {
			name = "TestAccPrivateNIC_SecurityGroup"
			vpc_id = scaleway_vpc.main.id
		}

		resource "scaleway_instance_server" "server" {
		    image = "ubuntu_focal"
		    type  = "DEV1-S"
		}
`

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			isPrivateNICDestroyed(tt),
			vpcchecks.CheckVPCDestroy(tt),
			instancechecks.IsServerDestroyed(tt),
			isSecurityGroupDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: unchangedConfig + `
					resource "scaleway_instance_security_group" "sg01" {
						name = "TestAccPrivateNIC_SecurityGroup"
						tags = [ "sg01" ]
					}

					resource "scaleway_instance_private_nic" "nic" {
						server_id          = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.pn.id
						security_group_id  = scaleway_instance_security_group.sg01.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("scaleway_instance_private_nic.nic", "srn",
						regexp.MustCompile(`^srn://instance\..+/zones/.+/private-network-interfaces/.+$`)),
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "security_group_id",
						"scaleway_instance_security_group.sg01", "id"),
				),
			},
			{
				ResourceName:            "scaleway_instance_private_nic.nic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ipam_ip_ids", "status"},
			},
			{
				Config: unchangedConfig + `
					resource "scaleway_instance_security_group" "sg01" {
						name = "TestAccPrivateNIC_SecurityGroup"
						tags = [ "sg01" ]
					}
					resource "scaleway_instance_security_group" "sg02" {
						name = "TestAccPrivateNIC_SecurityGroup"
						tags = [ "sg02" ]
					}

					resource "scaleway_instance_private_nic" "nic" {
						server_id          = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.pn.id
						security_group_id  = scaleway_instance_security_group.sg02.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("scaleway_instance_private_nic.nic", "srn",
						regexp.MustCompile(`^srn://instance\..+/zones/.+/private-network-interfaces/.+$`)),
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "security_group_id",
						"scaleway_instance_security_group.sg02", "id"),
				),
			},
			{
				Config: unchangedConfig + `
					resource "scaleway_instance_security_group" "sg02" {
						name = "TestAccPrivateNIC_SecurityGroup"
						tags = [ "sg02" ]
					}

					resource "scaleway_instance_private_nic" "nic" {
						server_id          = scaleway_instance_server.server.id
						private_network_id = scaleway_vpc_private_network.pn.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("scaleway_instance_private_nic.nic", "srn",
						regexp.MustCompile(`^srn://instance\..+/zones/.+/private-network-interfaces/.+$`)),
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "security_group_id", ""),
				),
			},
		},
	})
}

func TestAccPrivateNIC_WithoutServer(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	unchangedConfig := `
		resource "scaleway_vpc" "main" {
			name = "TestAccPrivateNIC_WithoutServer"
		}

		resource "scaleway_vpc_private_network" "pn" {
			name = "TestAccPrivateNIC_WithoutServer"
			vpc_id = scaleway_vpc.main.id
		}
`

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			isPrivateNICDestroyed(tt),
			vpcchecks.CheckVPCDestroy(tt),
			instancechecks.IsServerDestroyed(tt),
			isSecurityGroupDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: unchangedConfig + `
					resource "scaleway_instance_private_nic" "nic" {
						private_network_id = scaleway_vpc_private_network.pn.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("scaleway_instance_private_nic.nic", "srn",
						regexp.MustCompile(`^srn://instance\..+/zones/.+/private-network-interfaces/.+$`)),
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "private_network_id",
						"scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "server_id", ""),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "status", "available"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "project_id",
						"scaleway_vpc_private_network.pn", "project_id"),
				),
			},
			{
				Config: unchangedConfig + `
					resource "scaleway_instance_server" "server" {
					    image = "ubuntu_focal"
					    type  = "DEV1-S"
					}

					resource "scaleway_instance_private_nic" "nic" {
						private_network_id = scaleway_vpc_private_network.pn.id
						server_id		   = scaleway_instance_server.server.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("scaleway_instance_private_nic.nic", "srn",
						regexp.MustCompile(`^srn://instance\..+/zones/.+/private-network-interfaces/.+$`)),
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "private_network_id",
						"scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "server_id",
						"scaleway_instance_server.server", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "project_id",
						"scaleway_vpc_private_network.pn", "project_id"),
				),
			},
			{
				Config: unchangedConfig + `
					resource "scaleway_instance_private_nic" "nic" {
						private_network_id = scaleway_vpc_private_network.pn.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("scaleway_instance_private_nic.nic", "srn",
						regexp.MustCompile(`^srn://instance\..+/zones/.+/private-network-interfaces/.+$`)),
					isPrivateNICPresent(tt, "scaleway_instance_private_nic.nic"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "private_network_id",
						"scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "server_id", ""),
					resource.TestCheckResourceAttr("scaleway_instance_private_nic.nic", "status", "available"),
					resource.TestCheckResourceAttrPair("scaleway_instance_private_nic.nic", "project_id",
						"scaleway_vpc_private_network.pn", "project_id"),
				),
			},
		},
	})
}

func isPrivateNICPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, pnicID, _, err := instance.NewAPIV2WithZoneAndNestedID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetPrivateNetworkInterface(&instanceSDK.GetPrivateNetworkInterfaceRequest{
			PrivateNetworkInterfaceID: pnicID,
			Zone:                      zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isPrivateNICDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, instancechecks.DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_instance_private_nic" {
					continue
				}

				api, zone, pnicID, _, err := instance.NewAPIV2WithZoneAndNestedID(tt.Meta, rs.Primary.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				_, err = api.GetPrivateNetworkInterface(&instanceSDK.GetPrivateNetworkInterfaceRequest{
					PrivateNetworkInterfaceID: pnicID,
					Zone:                      zone,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("instance private NIC (%s) still exists", rs.Primary.ID))
				case httperrors.Is404(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}
