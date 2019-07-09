package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestAccScalewayComputeInstancePlacementGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayComputeInstancePlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayComputeInstancePlacementGroupConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstancePlacementGroupExists("scaleway_compute_instance_placement_group.base"),
					testAccCheckScalewayComputeInstancePlacementGroupExists("scaleway_compute_instance_placement_group.scaleway"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.base", "policy_mode", "optional"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.base", "policy_type", "low_latency"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.scaleway", "policy_mode", "enforced"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.scaleway", "policy_type", "max_availability"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.scaleway", "policy_respected", "true"),
				),
			},
			{
				Config: testAccScalewayComputeInstancePlacementGroupConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayComputeInstancePlacementGroupExists("scaleway_compute_instance_placement_group.base"),
					testAccCheckScalewayComputeInstancePlacementGroupExists("scaleway_compute_instance_placement_group.scaleway"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.base", "policy_mode", "enforced"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.base", "policy_type", "max_availability"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.base", "policy_respected", "true"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.scaleway", "policy_mode", "optional"),
					resource.TestCheckResourceAttr("scaleway_compute_instance_placement_group.scaleway", "policy_type", "low_latency"),
				),
			},
		},
	})
}

func testAccCheckScalewayComputeInstancePlacementGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetComputeCluster(&instance.GetComputeClusterRequest{
			Zone:             zone,
			ComputeClusterID: ID,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayComputeInstancePlacementGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_compute_instance_placement_group" {
			continue
		}

		instanceApi, zone, ID, err := getInstanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetComputeCluster(&instance.GetComputeClusterRequest{
			Zone:             zone,
			ComputeClusterID: ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("Placement group (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}

// Check that reverse is handled at creation and update time
var testAccScalewayComputeInstancePlacementGroupConfig = []string{
	`
		resource "scaleway_compute_instance_placement_group" "base" {}
		resource "scaleway_compute_instance_placement_group" "scaleway" {
			policy_mode = "enforced"
			policy_type = "max_availability"
		}
	`,
	`
		resource "scaleway_compute_instance_placement_group" "base" {
			policy_mode = "enforced"
			policy_type = "max_availability"
		}
		resource "scaleway_compute_instance_placement_group" "scaleway" {}
	`,
}
