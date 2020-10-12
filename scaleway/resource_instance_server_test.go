package scaleway

import (
	"fmt"
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
		listServers, err := instanceAPI.ListServers(&instance.ListServersRequest{}, scw.WithAllPages())
		if err != nil {
			l.Warningf("error listing servers in (%s) in sweeper: %s", zone, err)
			return nil
		}

		for _, srv := range listServers.Servers {
			if srv.State == instance.ServerStateStopped || srv.State == instance.ServerStateStoppedInPlace {
				err := instanceAPI.DeleteServer(&instance.DeleteServerRequest{
					ServerID: srv.ID,
				})
				if err != nil {
					return fmt.Errorf("error deleting server in sweeper: %s", err)
				}
			} else if srv.State == instance.ServerStateRunning {
				_, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
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

func TestAccScalewayInstanceServerMinimal1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigMinimal("f974feac-abae-4365-b988-8ec7d1cec10d"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "fr-par-1/f974feac-abae-4365-b988-8ec7d1cec10d"),
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
				Config: testAccCheckScalewayInstanceServerConfigMinimal("ubuntu_focal"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
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

func TestAccScalewayInstanceServerC2S(t *testing.T) {
	t.Skip("C2S instance are EOL")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "scaleway_instance_server" "base" {
  image = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type  = "C2S"

  tags = [ "terraform-test", "scaleway_instance_server", "C2S" ]
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "fr-par-1/f974feac-abae-4365-b988-8ec7d1cec10d"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "C2S"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "boot_type", "bootscript"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "enable_dynamic_ip", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "C2S"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerRootVolume1(t *testing.T) {
	t.Skip("C2S often don't start. This is an issue on API. This server type is deprecated anyway")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigRootVolume("51", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "51"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "root_volume"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigRootVolume("52", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "52"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "root_volume"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerBasic1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigServerType("DEV1-M"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "type", "DEV1-M"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "name", "test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.1", "scaleway_instance_server"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "tags.2", "basic"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigServerType("DEV1-S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
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

func TestAccScalewayInstanceServerState1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigState("started"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "started"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigState("standby"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "standby"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigState("stopped"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "stopped"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerState2(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigState("stopped"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "stopped"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigState("standby"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "state", "standby"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerUserData1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigUserData(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.459781404.key", "plop"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.459781404.value", "world"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.599848950.key", "blanquette"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "user_data.599848950.value", "hareng pomme à l'huile"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "cloud_init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigUserData(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "user_data"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "cloud_init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigUserData(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "user_data"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "cloud_init", ""),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerUserData2(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigUserData(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "user_data"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "cloud_init"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigUserData(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_instance_server.base", "user_data"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "cloud_init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerAdditionalVolumes1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigVolumes(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "20"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigVolumes(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_block"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_block", "size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "20"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerAdditionalVolumes2(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigVolumes(true, 5, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume0"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume1"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_block"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume0", "size_in_gb", "5"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume1", "size_in_gb", "5"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_block", "size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigVolumes(true, 4, 3, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume0"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume1"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume2"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume3"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_block"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume0", "size_in_gb", "4"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume1", "size_in_gb", "3"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume2", "size_in_gb", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume3", "size_in_gb", "1"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_block", "size_in_gb", "10"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_block", "type", "b_ssd"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "10"),
				),
			},
			{
				Config: testAccCheckScalewayInstanceServerConfigVolumes(false, 4, 3, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume0"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume1"),
					testAccCheckScalewayInstanceVolumeExists("scaleway_instance_volume.base_volume2"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume0", "size_in_gb", "4"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume1", "size_in_gb", "3"),
					resource.TestCheckResourceAttr("scaleway_instance_volume.base_volume2", "size_in_gb", "2"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "root_volume.0.size_in_gb", "11"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerWithPlacementGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayInstanceServerConfigWithPlacementGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base.0"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base.1"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base.2"),
					testAccCheckScalewayInstancePlacementGroupExists("scaleway_instance_placement_group.ha"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base.0", "placement_group_policy_respected", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base.1", "placement_group_policy_respected", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base.2", "placement_group_policy_respected", "true"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerSwapVolume(t *testing.T) {
	tplFunc := newTemplateFunc(`
		resource "scaleway_instance_volume" "volume1" {
		  size_in_gb = 10
		  type       = "l_ssd"
		}
		resource "scaleway_instance_volume" "volume2" {
		  size_in_gb = 10
		  type       = "l_ssd"
		}
		resource "scaleway_instance_server" "server1" {
		  image = "ubuntu_focal"
		  type  = "DEV1-S"
		  root_volume {
			size_in_gb = 10
		  }
		  additional_volume_ids = [ scaleway_instance_volume.volume{{index . 0}}.id ]
		}
		resource "scaleway_instance_server" "server2" {
		  image = "ubuntu_focal"
		  type  = "DEV1-S"
		  root_volume {
			size_in_gb = 10
		  }
		  additional_volume_ids = [ scaleway_instance_volume.volume{{index . 1}}.id ]
		}
	`)

	var volume1Id, volume2Id string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: tplFunc([]int{1, 2}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.server1"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.server2"),
					testAccGetResourceAttr("scaleway_instance_server.server1", "additional_volume_ids.0", &volume1Id),
					testAccGetResourceAttr("scaleway_instance_server.server2", "additional_volume_ids.0", &volume2Id),
				),
			},
			{
				Config: tplFunc([]int{2, 1}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.server1"),
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.server2"),
					resource.TestCheckResourceAttrPtr("scaleway_instance_server.server1", "additional_volume_ids.0", &volume2Id),
					resource.TestCheckResourceAttrPtr("scaleway_instance_server.server2", "additional_volume_ids.0", &volume1Id),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerIpv6(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
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
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.server01"),
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
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.server01"),
					resource.TestCheckResourceAttr("scaleway_instance_server.server01", "ipv6_address", ""),
					resource.TestCheckResourceAttr("scaleway_instance_server.server01", "ipv6_gateway", ""),
					resource.TestCheckResourceAttr("scaleway_instance_server.server01", "ipv6_prefix_length", "0"),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerImport(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "server01" {
						type  = "DEV1-S"
						image = "f974feac-abae-4365-b988-8ec7d1cec10d"
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

func TestAccScalewayInstanceServerWithReservedIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						image = "f974feac-abae-4365-b988-8ec7d1cec10d"
						type  = "DEV1-S"
						ip_id = scaleway_instance_ip.first.id
						tags  = [ "terraform-test", "scaleway_instance_server", "reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.first", "address", "scaleway_instance_server.base", "public_ip"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.first", "id", "scaleway_instance_server.base", "ip_id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						image = "f974feac-abae-4365-b988-8ec7d1cec10d"
						type  = "DEV1-S"
						ip_id = scaleway_instance_ip.second.id
						tags  = [ "terraform-test", "scaleway_instance_server", "reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					testAccCheckScalewayInstanceIPPairWithServer("scaleway_instance_ip.second", "scaleway_instance_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.second", "address", "scaleway_instance_server.base", "public_ip"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.second", "id", "scaleway_instance_server.base", "ip_id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						image = "f974feac-abae-4365-b988-8ec7d1cec10d"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					testAccCheckScalewayInstanceServerNoIPAssigned("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "public_ip", ""),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "ip_id", ""),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						image = "f974feac-abae-4365-b988-8ec7d1cec10d"
						type  = "DEV1-S"
						enable_dynamic_ip = true
						tags  = [ "terraform-test", "scaleway_instance_server", "reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					testAccCheckScalewayInstanceServerNoIPAssigned("scaleway_instance_server.base"),
					testCheckResourceAttrIPv4("scaleway_instance_server.base", "public_ip"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "ip_id", ""),
				),
			},
		},
	})
}

func TestAccScalewayInstanceServerImageDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_instance_image" "ubuntu_focal" {
					  name = "Ubuntu 20.04 Focal Fossa"
					}

					resource "scaleway_instance_server" "base" {
					  type  = "DEV1-S"
					  image = data.scaleway_instance_image.ubuntu_focal.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceServerExists("scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "image", "fr-par-1/4e84fc90-baef-43c2-ba9c-caa135de7afd"),
				),
			},
			// Ensure that image diffSuppressFunc results in no plan.
			{
				Config: `
					data "scaleway_instance_image" "ubuntu_focal" {
					  name = "Ubuntu 20.04 Focal Fossa"
					}

					resource "scaleway_instance_server" "base" {
					  type  = "DEV1-S"
					  image = data.scaleway_instance_image.ubuntu_focal.id
					}
				`,
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckScalewayInstanceServerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
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

func testAccCheckScalewayInstanceServerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_instance_server" {
			continue
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetServer(&instance.GetServerRequest{
			ServerID: ID,
			Zone:     zone,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("Server (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}

func testAccCheckScalewayInstanceServerConfigMinimal(imageValue string) string {
	return fmt.Sprintf(`
resource "scaleway_instance_server" "base" {
  image = "%s"
  type  = "DEV1-S"

  tags = [ "terraform-test", "scaleway_instance_server", "minimal" ]
}`, imageValue)
}

func testAccCheckScalewayInstanceServerConfigServerType(serverType string) string {
	return fmt.Sprintf(`
data "scaleway_marketplace_image_beta" "ubuntu" {
  instance_type   = "%s"
  label         = "ubuntu_focal"
}

resource "scaleway_instance_server" "base" {
  name  = "test"
  image = "${data.scaleway_marketplace_image_beta.ubuntu.id}"
  type  = "%s"

  tags = [ "terraform-test", "scaleway_instance_server", "basic" ]
}`, serverType, serverType)
}

func testAccCheckScalewayInstanceServerConfigRootVolume(size, deleteOnTermination string) string {
	return fmt.Sprintf(`
resource "scaleway_instance_server" "base" {
  image = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type  = "C2S"
  root_volume {
    size_in_gb = %s
    delete_on_termination = %s
  }
  tags = [ "terraform-test", "scaleway_instance_server", "root_volume" ]
}`, size, deleteOnTermination)
}

func testAccCheckScalewayInstanceServerConfigState(state string) string {
	return fmt.Sprintf(`
data "scaleway_marketplace_image_beta" "ubuntu" {
  instance_type = "DEV1-S"
  label         = "ubuntu_focal"
}

resource "scaleway_instance_server" "base" {
  image = "${data.scaleway_marketplace_image_beta.ubuntu.id}"
  type  = "DEV1-S"
  state = "%s"
  tags  = [ "terraform-test", "scaleway_instance_server", "state" ]
}`, state)
}

func testAccCheckScalewayInstanceServerConfigUserData(withRandomUserData, withCloudInit bool) string {
	additionalUserData := ""
	if withRandomUserData {
		additionalUserData += `
  user_data {
    key   = "plop"
    value = "world"
  }

  user_data {
    key   = "blanquette"
    value = "hareng pomme à l'huile"
  }`
	}

	if withCloudInit {
		additionalUserData += `
  cloud_init = <<EOF
#cloud-config
apt_update: true
apt_upgrade: true
EOF`
	}

	return fmt.Sprintf(`
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu 20.04 Focal Fossa"
  most_recent  = true
}

resource "scaleway_instance_server" "base" {
  image = "${data.scaleway_image.ubuntu.id}"
  type  = "DEV1-S"
  tags  = [ "terraform-test", "scaleway_instance_server", "user_data" ]
%s
}`, additionalUserData)
}

func testAccCheckScalewayInstanceServerConfigVolumes(withBlock bool, localVolumesInGB ...int) string {
	additionalVolumeResources := ""
	baseVolume := 20
	var additionalVolumeIDs []string
	for i, size := range localVolumesInGB {
		additionalVolumeResources += fmt.Sprintf(`
resource "scaleway_instance_volume" "base_volume%d" {
  size_in_gb = %d
  type       = "l_ssd"
}`, i, size)
		additionalVolumeIDs = append(additionalVolumeIDs, fmt.Sprintf(`"${scaleway_instance_volume.base_volume%d.id}"`, i))
		baseVolume -= size
	}

	if withBlock {
		additionalVolumeResources += `
resource "scaleway_instance_volume" "base_block" {
  size_in_gb = 10
  type       = "b_ssd"
}`
		additionalVolumeIDs = append(additionalVolumeIDs, `"${scaleway_instance_volume.base_block.id}"`)
	}
	return fmt.Sprintf(`
%s

resource "scaleway_instance_server" "base" {
  image = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type  = "DEV1-S"
  root_volume {
    size_in_gb = %d
  }
  tags = [ "terraform-test", "scaleway_instance_server", "additional_volume_ids" ]

  additional_volume_ids  = [ %s ]
}`, additionalVolumeResources, baseVolume, strings.Join(additionalVolumeIDs, ","))
}

var testAccCheckScalewayInstanceServerConfigWithPlacementGroup = `
resource "scaleway_instance_placement_group" "ha" {
	policy_mode = "enforced"
	policy_type = "max_availability"
}

resource "scaleway_instance_server" "base" {
	count = 3
	image = "f974feac-abae-4365-b988-8ec7d1cec10d"
	type  = "DEV1-S"
	placement_group_id = "${scaleway_instance_placement_group.ha.id}"
    tags  = [ "terraform-test", "scaleway_instance_server", "placement_group" ]
}
`

// todo: add a test with security groups
