package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

// Check that reverse is handled at creation and update time
var testAccScalewayComputeInstanceIPConfig = []string{
	`
		resource "scaleway_compute_instance_ip" "base" {}
		resource "scaleway_compute_instance_ip" "scaleway" {
			reverse = "www.scaleway.com"
		}
	`,
	`
		resource "scaleway_compute_instance_ip" "base" {
			reverse = "www.scaleway.com"	
		}
		resource "scaleway_compute_instance_ip" "scaleway" {}
	`,
}

// Check that we can change the zone of an ip (delete + create)
var testAccScalewayComputeInstanceIPZoneConfig = []string{
	`
		resource "scaleway_compute_instance_ip" "base" {}
	`,
	`
		resource "scaleway_compute_instance_ip" "base" {
			zone = "nl-ams-1"	
		}
	`,
}

func TestAccScalewayComputeInstanceIP(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayComputeInstanceIPConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base"),
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.scaleway"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base", "reverse", ""),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.scaleway", "reverse", "www.scaleway.com"),
				),
			},
			{
				Config: testAccScalewayComputeInstanceIPConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base"),
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.scaleway"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base", "reverse", "www.scaleway.com"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.scaleway", "reverse", ""),
				),
			},
		},
	})
}

func TestAccScalewayComputeInstanceIP_Zone(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstanceIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayComputeInstanceIPZoneConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base", "zone", "fr-par-1"),
				),
			},
			{
				Config: testAccScalewayComputeInstanceIPZoneConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstanceIPExists("scaleway_compute_instance_ip.base"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_ip.base", "zone", "nl-ams-1"),
				),
			},
		},
	})
}

func testAccCheckScalewayComputeInstanceIPExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetIP(&instance.GetIPRequest{
			IPID: ID,
			Zone: zone,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayComputeInstanceIPDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_compute_instance_ip" {
			continue
		}

		instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetIP(&instance.GetIPRequest{
			Zone: zone,
			IPID: ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("IP (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		// We check for 403 because instance API return 403 for deleted IP
		if !is404Error(err) && !is403Error(err) {
			return err
		}
	}

	return nil
}
