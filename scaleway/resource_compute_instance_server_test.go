package scaleway

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

// Check that type force recreate the server
var testAccScalewayComputeInstanceServerConfig = []string{
	`
		resource "scaleway_compute_instance_server" "webserver" {
  			image = "f974feac-abae-4365-b988-8ec7d1cec10d"
  			type  = "DEV1-M"
		}
	`,
	`
		resource "scaleway_compute_instance_server" "webserver" {
  			image = "f974feac-abae-4365-b988-8ec7d1cec10d"
  			type  = "DEV1-S"
		}
	`,
}

func TestAccScalewayComputeInstanceServer(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayComputeInstanceServerConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.webserver"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "image", "f974feac-abae-4365-b988-8ec7d1cec10d"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "type", "DEV1-M"),
				),
			},
			{
				Config: testAccScalewayComputeInstanceServerConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceServerExists("scaleway_compute_instance_server.webserver"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "image", "f974feac-abae-4365-b988-8ec7d1cec10d"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_server.webserver", "type", "DEV1-S"),
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

func Test_stateToAction(t *testing.T) {
	type args struct {
		state string
	}
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
