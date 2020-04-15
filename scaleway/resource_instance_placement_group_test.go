package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_placement_group", &resource.Sweeper{
		Name: "scaleway_instance_placement_group",
		F:    testSweepInstancePlacementGroup,
	})
}

func testSweepInstancePlacementGroup(region string) error {
	return sweepZones(region, func(scwClient *scw.Client) error {
		instanceAPI := instance.NewAPI(scwClient)
		zone, _ := scwClient.GetDefaultZone()
		l.Debugf("sweeper: destroying the instance placement group in (%s)", zone)
		listPlacementGroups, err := instanceAPI.ListPlacementGroups(&instance.ListPlacementGroupsRequest{}, scw.WithAllPages())
		if err != nil {
			l.Warningf("error listing placement groups in (%s) in sweeper: %s", zone, err)
			return nil
		}

		for _, pg := range listPlacementGroups.PlacementGroups {
			err := instanceAPI.DeletePlacementGroup(&instance.DeletePlacementGroupRequest{
				PlacementGroupID: pg.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting placement group in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayInstancePlacementGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayInstancePlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayInstancePlacementGroupConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePlacementGroupExists("scaleway_instance_placement_group.base"),
					testAccCheckScalewayInstancePlacementGroupExists("scaleway_instance_placement_group.scaleway"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.base", "policy_mode", "optional"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.base", "policy_type", "max_availability"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.scaleway", "policy_mode", "enforced"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.scaleway", "policy_type", "low_latency"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.scaleway", "policy_respected", "true"),
				),
			},
			{
				Config: testAccScalewayInstancePlacementGroupConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePlacementGroupExists("scaleway_instance_placement_group.base"),
					testAccCheckScalewayInstancePlacementGroupExists("scaleway_instance_placement_group.scaleway"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.base", "policy_mode", "enforced"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.base", "policy_type", "low_latency"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.base", "policy_respected", "true"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.scaleway", "policy_mode", "optional"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.scaleway", "policy_type", "max_availability"),
				),
			},
		},
	})
}

func testAccCheckScalewayInstancePlacementGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceApi, zone, ID, err := instanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetPlacementGroup(&instance.GetPlacementGroupRequest{
			Zone:             zone,
			PlacementGroupID: ID,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayInstancePlacementGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_instance_placement_group" {
			continue
		}

		instanceApi, zone, ID, err := instanceAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceApi.GetPlacementGroup(&instance.GetPlacementGroupRequest{
			Zone:             zone,
			PlacementGroupID: ID,
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
var testAccScalewayInstancePlacementGroupConfig = []string{
	`
		resource "scaleway_instance_placement_group" "base" {}
		resource "scaleway_instance_placement_group" "scaleway" {
			policy_mode = "enforced"
			policy_type = "low_latency"
		}
	`,
	`
		resource "scaleway_instance_placement_group" "base" {
			policy_mode = "enforced"
			policy_type = "low_latency"
		}
		resource "scaleway_instance_placement_group" "scaleway" {}
	`,
}
