package scaleway

import (
	"fmt"
	"reflect"
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
				Config: testAccCheckScalewayComputeInstanceServerConfigMinimal1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.webserver"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "image", "f974feac-abae-4365-b988-8ec7d1cec10d"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "tags.1", "scaleway_compute_instance_server"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "tags.2", "minimal1"),
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
				Config: testAccCheckScalewayComputeInstanceServerConfigBasic1("x86_64", "DEV1-M"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "type", "DEV1-M"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "name", "test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.1", "scaleway_compute_instance_server"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.2", "basic1"),
				),
			},
			{
				Config: testAccCheckScalewayComputeInstanceServerConfigBasic1("x86_64", "DEV1-S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "type", "DEV1-S"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "name", "test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.1", "scaleway_compute_instance_server"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.base", "tags.2", "basic1"),
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

func testAccCheckScalewayComputeInstanceServerConfigMinimal1() string {
	return `
resource "scaleway_compute_instance_server" "webserver" {
  image = "f974feac-abae-4365-b988-8ec7d1cec10d"
  type  = "DEV1-S"
  tags = [ "terraform-test", "scaleway_compute_instance_server", "minimal1" ]
}`
}

func testAccCheckScalewayComputeInstanceServerConfigBasic1(architecture, serverType string) string {
	return fmt.Sprintf(`
data "scaleway_image" "ubuntu" {
  architecture = "%s"
  name         = "Ubuntu Xenial"
  most_recent  = true
}

resource "scaleway_compute_instance_server" "base" {
  name = "test"
  image = "${data.scaleway_image.ubuntu.id}"
  type = "%s"
  tags = [ "terraform-test", "scaleway_compute_instance_server", "basic1" ]
}`, architecture, serverType)
}

// todo: add tests with IP attachment

// todo: add tests with additional volume attachement

// todo: add a test with security groups

// todo: add a test with user data

// todo: add a test with cloud init (in user data)

func Test_stateToAction(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  instance.ServerAction
	}{
		{
			name:  "Started",
			state: "started",
			want:  instance.ServerActionPoweron,
		},
		{
			name:  "Stopped",
			state: "stopped",
			want:  instance.ServerActionPoweroff,
		},
		{
			name:  "Standby",
			state: "standby",
			want:  instance.ServerActionStopInPlace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stateToAction(tt.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateToAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
