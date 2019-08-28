package scaleway

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestAccScalewayComputeInstanceServerMinimal1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "image_id", "f974feac-abae-4365-b988-8ec7d1cec10d"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.size_in_gb", "20"),
					resource.TestCheckResourceAttrSet("scaleway_compute_instance_server.base", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.1", "scaleway_compute_instance_server"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.2", "minimal"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerRootVolume1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigRootVolume("60", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.size_in_gb", "60"),
					resource.TestCheckResourceAttrSet("scaleway_compute_instance_server.base", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.2", "root_volume"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigRootVolume("200", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.size_in_gb", "200"),
					resource.TestCheckResourceAttrSet("scaleway_compute_instance_server.base", "root_volume.0.volume_id"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.2", "root_volume"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerBasic1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigServerType("x86_64", "DEV1-M"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "type", "DEV1-M"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "name", "test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.1", "scaleway_compute_instance_server"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.2", "basic"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigServerType("x86_64", "DEV1-S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "name", "test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.1", "scaleway_compute_instance_server"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.2", "basic"),
				),
			},
		},
	})
}
func TestAccScalewayComputeInstanceServerRemoteExec(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigWithIPAndVolume,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.webserver"),
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.data"),
					testAccCheckScalewayInstanceIPExists("scaleway_instance_ip.myip"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerState1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigState("started"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "state", "started"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigState("standby"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "state", "standby"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigState("stopped"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "state", "stopped"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerState2(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigState("stopped"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "state", "stopped"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigState("standby"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "state", "standby"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerUserData1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigUserData(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "user_data.459781404.key", "plop"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "user_data.459781404.value", "world"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "user_data.599848950.key", "blanquette"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "user_data.599848950.value", "hareng pomme à l'huile"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "cloud_init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigUserData(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_compute_instance_server.base", "user_data"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "cloud_init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigUserData(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_compute_instance_server.base", "user_data"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "cloud_init", ""),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerUserData2(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigUserData(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_compute_instance_server.base", "user_data"),
					resource.TestCheckNoResourceAttr("scaleway_compute_instance_server.base", "cloud_init"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigUserData(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckNoResourceAttr("scaleway_compute_instance_server.base", "user_data"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "cloud_init", "#cloud-config\napt_update: true\napt_upgrade: true\n"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerAdditionalVolumes1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigVolumes(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.size_in_gb", "20"),
				),
			},
			// Comming Soon
			// {
			// 	Config: testAccCheckScalewayComputeInstanceServerConfigVolumes(true),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_block"),
			// 		testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
			// 		resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_block", "size_in_gb", "100"),
			// 		resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.size_in_gb", "20"),
			// 	),
			// },
		},
	})
}

func TestAccScalewayComputeInstanceServerAdditionalVolumes2(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigVolumes(false, 5, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume0"),
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume1"),
					// testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_block"),
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume0", "size_in_gb", "5"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume1", "size_in_gb", "5"),
					// resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_block", "size_in_gb", "100"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.size_in_gb", "10"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigVolumes(false, 4, 3, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume0"),
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume1"),
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume2"),
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume3"),
					// testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_block"),
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume0", "size_in_gb", "4"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume1", "size_in_gb", "3"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume2", "size_in_gb", "2"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume3", "size_in_gb", "1"),
					// resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_block", "size_in_gb", "100"),
					// resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_block", "type", "l_ssd"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.size_in_gb", "10"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigVolumes(false, 4, 3, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume0"),
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume1"),
					testAccCheckScalewayComputeInstanceVolumeExists("scaleway_compute_instance_volume.base_volume2"),
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume0", "size_in_gb", "4"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume1", "size_in_gb", "3"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_volume.base_volume2", "size_in_gb", "2"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "root_volume.0.size_in_gb", "11"),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceServerWithPlacementGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigWithPlacementGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base.0"),
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base.1"),
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base.2"),
					testAccCheckScalewayComputeInstancePlacementGroupExists("scaleway_compute_instance_placement_group.ha"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base.0", "placement_group_policy_respected", "true"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base.1", "placement_group_policy_respected", "true"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base.2", "placement_group_policy_respected", "true"),
				),
			},
		},
	})
}

func testAccCheckScalewayComputeInstanceServerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
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

func testAccCheckScalewayComputeInstanceServerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_instance_ip" {
			continue
		}

		instanceAPI, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
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

func testAccCheckScalewayComputeInstanceServerConfigMinimal() string {
	return `
resource "scaleway_compute_instance_server" "base" {
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type     = "DEV1-S"

  tags = [ "terraform-test", "scaleway_compute_instance_server", "minimal" ]
}`
}

func testAccCheckScalewayComputeInstanceServerConfigServerType(architecture, serverType string) string {
	return fmt.Sprintf(`
data "scaleway_image" "ubuntu" {
  architecture = "%s"
  name         = "Ubuntu Bionic"
  most_recent  = true
}

resource "scaleway_compute_instance_server" "base" {
  name     = "test"
  image_id = "${data.scaleway_image.ubuntu.id}"
  type     = "%s"

  tags = [ "terraform-test", "scaleway_compute_instance_server", "basic" ]
}`, architecture, serverType)
}

func testAccCheckScalewayComputeInstanceServerConfigRootVolume(size, deleteOnTermination string) string {
	return fmt.Sprintf(`
resource "scaleway_compute_instance_server" "base" {
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type     = "C2S"
  root_volume {
    size_in_gb = %s
    delete_on_termination = %s
  }
  tags = [ "terraform-test", "scaleway_compute_instance_server", "root_volume" ]
}`, size, deleteOnTermination)
}

func testAccCheckScalewayComputeInstanceServerConfigState(state string) string {
	return fmt.Sprintf(`
data "scaleway_image" "ubuntu" {
  architecture = "x86_64"
  name         = "Ubuntu Bionic"
  most_recent  = true
}

resource "scaleway_compute_instance_server" "base" {
  image_id = "${data.scaleway_image.ubuntu.id}"
  type     = "DEV1-S"
  state    = "%s"
  tags     = [ "terraform-test", "scaleway_compute_instance_server", "state" ]
}`, state)
}

func testAccCheckScalewayComputeInstanceServerConfigUserData(withRandomUserData, withCloudInit bool) string {
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
  name         = "Ubuntu Bionic"
  most_recent  = true
}

resource "scaleway_compute_instance_server" "base" {
  image_id = "${data.scaleway_image.ubuntu.id}"
  type     = "DEV1-S"
  tags     = [ "terraform-test", "scaleway_compute_instance_server", "user_data" ]
%s
}`, additionalUserData)
}

func testAccCheckScalewayComputeInstanceServerConfigVolumes(withBlock bool, localVolumesInGB ...int) string {
	additionalVolumeResources := ""
	baseVolume := 20
	var additionalVolumeIDs []string
	for i, size := range localVolumesInGB {
		additionalVolumeResources += fmt.Sprintf(`
resource "scaleway_compute_instance_volume" "base_volume%d" {
  size_in_gb = %d
  type       = "l_ssd"
}`, i, size)
		additionalVolumeIDs = append(additionalVolumeIDs, fmt.Sprintf(`"${scaleway_compute_instance_volume.base_volume%d.id}"`, i))
		baseVolume -= size
	}

	if withBlock {
		additionalVolumeResources += fmt.Sprintf(`
resource "scaleway_compute_instance_volume" "base_block" {
  size_in_gb = 100
}`)
		additionalVolumeIDs = append(additionalVolumeIDs, `"${scaleway_compute_instance_volume.base_block.id}"`)

	}
	return fmt.Sprintf(`
%s

resource "scaleway_compute_instance_server" "base" {
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type     = "DEV1-S"
  root_volume {
    size_in_gb = %d
  }
  tags = [ "terraform-test", "scaleway_compute_instance_server", "additional_volume_ids" ]

  additional_volume_ids  = [ %s ]
}`, additionalVolumeResources, baseVolume, strings.Join(additionalVolumeIDs, ","))
}

var testAccCheckScalewayComputeInstanceServerConfigWithIPAndVolume = `
resource "scaleway_instance_ip" "myip" {
  server_id = "${scaleway_compute_instance_server.webserver.id}"
}

resource "scaleway_compute_instance_volume" "data" {
  size_in_gb = 100
  type       = "l_ssd"
}

resource "scaleway_compute_instance_server" "webserver" {
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type     = "DEV1-S"

  additional_volume_ids = [ "${scaleway_compute_instance_volume.data.id}" ]
}
`

var testAccCheckScalewayComputeInstanceServerConfigWithPlacementGroup = `
resource "scaleway_compute_instance_placement_group" "ha" {
	policy_mode = "enforced"
	policy_type = "max_availability"
}

resource "scaleway_compute_instance_server" "base" {
	count = 3
	image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
	type     = "DEV1-S"
	placement_group_id = "${scaleway_compute_instance_placement_group.ha.id}"
}
`

// todo: add a test with security groups
