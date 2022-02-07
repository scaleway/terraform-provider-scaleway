package scaleway

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_server", &resource.Sweeper{
		Name: "scaleway_instance_server",
		F:    testSweepInstanceServer,
	})
}

func testSweepInstanceServer(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		instanceAPI := instance.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the instance server in (%s)", zone)
		listServers, err := instanceAPI.ListServers(&instance.ListServersRequest{Zone: zone}, scw.WithAllPages())
		if err != nil {
			l.Warningf("error listing servers in (%s) in sweeper: %s", zone, err)
			return nil
		}

		for _, srv := range listServers.Servers {
			if srv.State == instance.ServerStateStopped || srv.State == instance.ServerStateStoppedInPlace {
				err := instanceAPI.DeleteServer(&instance.DeleteServerRequest{
					Zone:     zone,
					ServerID: srv.ID,
				})
				if err != nil {
					return fmt.Errorf("error deleting server in sweeper: %s", err)
				}
			} else if srv.State == instance.ServerStateRunning {
				_, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
					Zone:     zone,
					ServerID: srv.ID,
					Action:   instance.ServerActionTerminate,
				})
				if err != nil {
					return fmt.Errorf("error terminating server in sweeper: %s", err)
				}
			}
		}

		return nil
	})
}

func TestAccScalewayInstanceServer_Minimal1(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "base" {
					  image = "ubuntu_focal"
					  type  = "DEV1-S"
					
					  tags = [ "terraform-test", "scaleway_instance_server", "minimal" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "20"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "enable_dynamic_ip", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "minimal"),
				),
			},
			{
				// Image label such as ubuntu_focal
				Config: `
					resource "scaleway_instance_server" "base" {
					  image = "ubuntu_focal"
					  type  = "DEV1-S"
					
					  tags = [ "terraform-test", "scaleway_instance_server", "minimal" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "ubuntu_focal"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "20"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "minimal"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_RootVolume1(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						root_volume {
							size_in_gb = 10
							delete_on_termination = true
						}
						tags = [ "terraform-test", "scaleway_instance_server", "root_volume" ]
					}`,
				ExpectError: regexp.MustCompile("total local volume size must be equal to 20 GB"),
			},
		},
	})
}

func TestAccScalewayInstanceServer_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				// DEV1-M
				Config: `
					data "scaleway_marketplace_image" "ubuntu" {
					  instance_type   = "DEV1-M"
					  label         = "ubuntu_focal"
					}
					
					resource "scaleway_instance_server" "base" {
					  name  = "test"
					  image = "${data.scaleway_marketplace_image.ubuntu.id}"
					  type  = "DEV1-M"
					
					  tags = [ "terraform-test", "scaleway_instance_server", "basic" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-M"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "basic"),
				),
			},
			{
				// DEV1-S
				Config: `
					data "scaleway_marketplace_image" "ubuntu" {
					  instance_type   = "DEV1-S"
					  label         = "ubuntu_focal"
					}
					
					resource "scaleway_instance_server" "base" {
					  name  = "test"
					  image = "${data.scaleway_marketplace_image.ubuntu.id}"
					  type  = "DEV1-S"
					
					  tags = [ "terraform-test", "scaleway_instance_server", "basic" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "basic"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_State1(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				// started
				Config: `
					data "scaleway_marketplace_image" "ubuntu" {
					  instance_type = "DEV1-S"
					  label         = "ubuntu_focal"
					}
					
					resource "scaleway_instance_server" "base" {
					  image = "${data.scaleway_marketplace_image.ubuntu.id}"
					  type  = "DEV1-S"
					  state = "started"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "started"),
				),
			},
			{
				// standby
				Config: `
					data "scaleway_marketplace_image" "ubuntu" {
					  instance_type = "DEV1-S"
					  label         = "ubuntu_focal"
					}
					
					resource "scaleway_instance_server" "base" {
					  image = "${data.scaleway_marketplace_image.ubuntu.id}"
					  type  = "DEV1-S"
					  state = "standby"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "standby"),
				),
			},
			{
				// stopped
				Config: `
					data "scaleway_marketplace_image" "ubuntu" {
					  instance_type = "DEV1-S"
					  label         = "ubuntu_focal"
					}
					
					resource "scaleway_instance_server" "base" {
					  image = "${data.scaleway_marketplace_image.ubuntu.id}"
					  type  = "DEV1-S"
					  state = "stopped"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "stopped"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_State2(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				// stopped
				Config: `
					data "scaleway_marketplace_image" "ubuntu" {
					  instance_type = "DEV1-S"
					  label         = "ubuntu_focal"
					}
					
					resource "scaleway_instance_server" "base" {
					  image = "${data.scaleway_marketplace_image.ubuntu.id}"
					  type  = "DEV1-S"
					  state = "stopped"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "stopped"),
				),
			},
			{
				// standby
				Config: `
					data "scaleway_marketplace_image" "ubuntu" {
					  instance_type = "DEV1-S"
					  label         = "ubuntu_focal"
					}
					
					resource "scaleway_instance_server" "base" {
					  image = "${data.scaleway_marketplace_image.ubuntu.id}"
					  type  = "DEV1-S"
					  state = "standby"
					  tags  = [ "terraform-test", "scaleway_instance_server", "state" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "standby"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_UserData_WithCloudInitAtStart(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_instance_server" "base" {
					image = "ubuntu_focal"
					type  = "DEV1-S"

					user_data = {
				   		foo   = "bar"
						cloud-init =  <<EOF
#cloud-config
apt_update: true
apt_upgrade: true
EOF 
				 	}
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.foo", "bar"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.cloud-init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_UserData_WithoutCloudInitAtStart(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				// Without cloud-init
				Config: `
					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						root_volume {
							size_in_gb = 20
						}
						tags  = [ "terraform-test", "scaleway_instance_server", "user_data" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "user_data"),
				),
			},
			{
				// With cloud-init
				Config: `
					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "user_data" ]
						root_volume {
							size_in_gb = 20
						}
						user_data = {
							cloud-init = <<EOF
#cloud-config
apt_update: true
apt_upgrade: true
EOF
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.cloud-init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_AdditionalVolumes(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				// With additional local
				Config: `
					resource "scaleway_instance_volume" "local" {
						size_in_gb = 10
						type = "l_ssd"
					}

					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type = "DEV1-S"
						
						root_volume {
							size_in_gb = 10
						}

						tags = [ "terraform-test", "scaleway_instance_server", "additional_volume_ids" ]

						additional_volume_ids = [
							scaleway_instance_volume.local.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
				),
			},
			{
				// With additional local and block
				Config: `
					resource "scaleway_instance_volume" "local" {
						size_in_gb = 10
						type = "l_ssd"
					}

					resource "scaleway_instance_volume" "block" {
						size_in_gb = 10
						type = "b_ssd"
					}

					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type = "DEV1-S"
						
						root_volume {
							size_in_gb = 10
						}

						tags = [ "terraform-test", "scaleway_instance_server", "additional_volume_ids" ]

						additional_volume_ids = [
							scaleway_instance_volume.local.id,
							scaleway_instance_volume.block.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists(tt, "scaleway_instance_volume.block"),
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.block", "size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_WithPlacementGroup(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_placement_group" "ha" {
						policy_mode = "enforced"
						policy_type = "max_availability"
					}
					
					resource "scaleway_instance_server" "base" {
						count = 3
						image = "ubuntu_focal"
						type  = "DEV1-S"
						placement_group_id = "${scaleway_instance_placement_group.ha.id}"
						tags  = [ "terraform-test", "scaleway_instance_server", "placement_group" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base.0"),
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base.1"),
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base.2"),
					testAccCheckScalewayInstancePlacementGroupExists(tt, "scaleway_instance_placement_group.ha"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base.0", "placement_group_policy_respected", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base.1", "placement_group_policy_respected", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base.2", "placement_group_policy_respected", "true"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_Ipv6(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "server01" {
						image = "ubuntu_focal"
		  				type  = "DEV1-S"
		  				enable_ipv6 = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server01"),
					testCheckResourceAttrIPv6("scaleway_instance_server.server01", "ipv6_address"),
					testCheckResourceAttrIPv6("scaleway_instance_server.server01", "ipv6_gateway"),
					resource.TestCheckResourceAttr("scaleway_instance_server.server01", "ipv6_prefix_length", "64"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "server01" {
						image = "ubuntu_focal"
		  				type  = "DEV1-S"
		  				enable_ipv6 = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.server01"),
					resource.TestCheckResourceAttr("scaleway_instance_server.server01", "ipv6_address", ""),
					resource.TestCheckResourceAttr("scaleway_instance_server.server01", "ipv6_gateway", ""),
					resource.TestCheckResourceAttr("scaleway_instance_server.server01", "ipv6_prefix_length", "0"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_Basic2(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_marketplace_image" "ubuntu" {
					  instance_type   = "DEV1-M"
					  label         = "ubuntu_focal"
					}

					resource "scaleway_instance_server" "server01" {
						type  = "DEV1-S"
						image = data.scaleway_marketplace_image.ubuntu.id
						state = "stopped"
					}
				`,
			},
			{
				ResourceName:      "scaleway_instance_server.server01",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccScalewayInstanceServer_WithReservedIP(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						ip_id = scaleway_instance_ip.first.id
						tags  = [ "terraform-test", "scaleway_instance_server", "reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.first", "address", "scaleway_instance_server.base", "public_ip"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.first", "id", "scaleway_instance_server.base", "ip_id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						ip_id = scaleway_instance_ip.second.id
						tags  = [ "terraform-test", "scaleway_instance_server", "reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					testAccCheckScalewayInstanceIPPairWithServer(tt, "scaleway_instance_ip.second", "scaleway_instance_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.second", "address", "scaleway_instance_server.base", "public_ip"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.second", "id", "scaleway_instance_server.base", "ip_id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					testAccCheckScalewayInstanceServerNoIPAssigned(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "public_ip", ""),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "ip_id", ""),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						enable_dynamic_ip = true
						tags  = [ "terraform-test", "scaleway_instance_server", "reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					testAccCheckScalewayInstanceServerNoIPAssigned(tt, "scaleway_instance_server.base"),
					testCheckResourceAttrIPv4("scaleway_instance_server.base", "public_ip"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "ip_id", ""),
				),
			},
		},
	})
}

func testAccCheckScalewayInstanceServerExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetServer(&instance.GetServerRequest{ServerID: ID, Zone: zone})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayInstancePrivateNICsExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		res, err := instanceAPI.ListPrivateNICs(&instance.ListPrivateNICsRequest{ServerID: ID, Zone: zone})
		if err != nil {
			return err
		}

		privateNetworksOnServer := make(map[string]struct{})
		// build current private networks on server
		for _, key := range res.PrivateNics {
			privateNetworksOnServer[key.PrivateNetworkID] = struct{}{}
		}

		privateNetworksToCheckOnSchema := make(map[string]struct{})
		// build terraform private networks
		for key, value := range rs.Primary.Attributes {
			if strings.Contains(key, "pn_id") {
				privateNetworksToCheckOnSchema[expandID(value)] = struct{}{}
			}
		}

		// check if private networks are present on server
		for pnKey := range privateNetworksToCheckOnSchema {
			if _, exist := privateNetworksOnServer[pnKey]; !exist {
				return fmt.Errorf("private network does not exist")
			}
		}

		return nil
	}
}

func testAccCheckScalewayInstanceServerDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_server" {
				continue
			}

			instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = instanceAPI.GetServer(&instance.GetServerRequest{
				ServerID: ID,
				Zone:     zone,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("server (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func TestAccScalewayInstanceServer_Bootscript(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	// Quick tip to get all the different bootscript:
	// curl -sH "X-Auth-Token: $(scw config get secret-key)" https://api.scaleway.com/instance/v1/zones/fr-par-1/bootscripts | jq -r '.bootscripts[] | [.id, .architecture, .title] | @tsv'
	bootscript := "7decf961-d3e9-4711-93c7-b16c254e99b9"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_instance_server" "base" {
						type  = "DEV1-S"
						image = "ubuntu_focal"
						boot_type = "bootscript"
						bootscript_id = "%s"
					}
				`, bootscript),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "bootscript_id", bootscript),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_AlterTags(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "base" {
						type  = "DEV1-L"
						image = "ubuntu_focal"
						state = "stopped"
						tags = [ "front", "web" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "front"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "web"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "base" {
						type  = "DEV1-L"
						state = "stopped"
						image = "ubuntu_focal"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "tags"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_WithDefaultRootVolumeAndAdditionalVolume(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "data" {
						size_in_gb = 100
						type = "b_ssd"
					}

					resource "scaleway_instance_server" "main" {
						type = "DEV1-S"
						image = "ubuntu-bionic"
						root_volume {
							delete_on_termination = false
					  	}
						additional_volume_ids = [ scaleway_instance_volume.data.id ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.main"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_Enterprise(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "main" {
						type  = "ENT1-S"
						image = "ubuntu_focal"
						zone  = "fr-par-2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.main"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_ServerWithBlockNonDefaultZone(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_volume" "main" {
						type       = "b_ssd"
						name       = "main"
						size_in_gb = 1
						zone       = "nl-ams-1"
					}

					resource "scaleway_instance_server" "main" {
						zone              = "nl-ams-1"
						image             = "ubuntu_focal"
						type              = "DEV1-S"
						root_volume {
							delete_on_termination = true
							size_in_gb            = 20
						}
						additional_volume_ids = [scaleway_instance_volume.main.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.main"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServer_PrivateNetwork(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network internal {
						name = "private_network_instance"
						zone = "fr-par-2"
					}

					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						zone = "fr-par-2"

						private_network {
							pn_id = scaleway_vpc_private_network.internal.id
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePrivateNICsExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.zone"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "private_network.0.pn_id",
						"scaleway_vpc_private_network.internal", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "private_network_instance"
					}

					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"
						zone  = "fr-par-1"

						private_network {
							pn_id = scaleway_vpc_private_network.pn01.id
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePrivateNICsExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.zone"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "private_network.0.pn_id",
						"scaleway_vpc_private_network.pn01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "private_network_instance"
					}

					resource scaleway_vpc_private_network pn02 {
						name = "private_network_instance_02"
					}

					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"

						private_network {
							pn_id = scaleway_vpc_private_network.pn02.id
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePrivateNICsExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.zone"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "private_network.0.pn_id",
						"scaleway_vpc_private_network.pn02", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "private_network_instance"
					}

					resource scaleway_vpc_private_network pn02 {
						name = "private_network_instance_02"
					}

					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"

						private_network {
							pn_id = scaleway_vpc_private_network.pn02.id
						}

						private_network {
							pn_id = scaleway_vpc_private_network.pn01.id
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePrivateNICsExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base",
						"private_network.0.pn_id",
						"scaleway_vpc_private_network.pn02", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "2"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.zone"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "private_network.1.pn_id",
						"scaleway_vpc_private_network.pn01", "id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.1.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.1.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.1.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.1.zone"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "private_network_instance"
					}

					resource scaleway_vpc_private_network pn02 {
						name = "private_network_instance_02"
					}

					resource "scaleway_instance_server" "base" {
						image = "ubuntu_focal"
						type  = "DEV1-S"	
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "0"),
				),
			},
		},
	})
}
