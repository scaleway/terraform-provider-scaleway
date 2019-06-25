package scaleway

import (
	"fmt"
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
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "cloud_init", ""),
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

func testAccCheckScalewayComputeInstanceServerExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetServer(&instance.GetServerRequest{ServerID: ID, Zone: zone})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayComputeInstanceServerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_compute_instance_ip" {
			continue
		}

		instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetServer(&instance.GetServerRequest{
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
  type  = "DEV1-S"
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
  name = "test"
  image_id = "${data.scaleway_image.ubuntu.id}"
  type = "%s"
  tags = [ "terraform-test", "scaleway_compute_instance_server", "basic" ]
}`, architecture, serverType)
}

func testAccCheckScalewayComputeInstanceServerConfigRootVolume(size, deleteOnTermination string) string {
	return fmt.Sprintf(`
resource "scaleway_compute_instance_server" "base" {
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type  = "C2S"
  tags = [ "terraform-test", "scaleway_compute_instance_server", "root_volume" ]
  root_volume {
    size_in_gb = %s
    delete_on_termination = %s
  }
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
  type  = "DEV1-S"
  state = "%s"
  tags  = [ "terraform-test", "scaleway_compute_instance_server", "state" ]
}`, state)
}

func testAccCheckScalewayComputeInstanceServerConfigUserData(withRandomUserData, withCloudInit bool) string {
	additionalUserData := ""
	if withRandomUserData {
		additionalUserData += `
  user_data {
    key = "plop"
    value = "world"
  }

  user_data {
    key = "blanquette"
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
  type  = "DEV1-S"
  tags  = [ "terraform-test", "scaleway_compute_instance_server", "user_data" ]
%s
}`, additionalUserData)
}

// todo: add tests with IP attachment

// todo: add tests with additional volume attachement

// todo: add a test with security groups
