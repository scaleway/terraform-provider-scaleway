package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_instance_placement_group", &resource.Sweeper{
		Name: "scaleway_instance_placement_group",
		F:    testSweepInstancePlacementGroup,
	})
}

func testSweepInstancePlacementGroup(_ string) error {
	return sweepZones(scw.AllZones, func(scwClient *scw.Client, zone scw.Zone) error {
		instanceAPI := instance.NewAPI(scwClient)
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

func TestAccScalewayInstancePlacementGroup_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstancePlacementGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_placement_group" "base" {}

					resource "scaleway_instance_placement_group" "scaleway" {
						policy_mode = "enforced"
						policy_type = "low_latency"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePlacementGroupExists(tt, "scaleway_instance_placement_group.base"),
					testAccCheckScalewayInstancePlacementGroupExists(tt, "scaleway_instance_placement_group.scaleway"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.base", "policy_mode", "optional"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.base", "policy_type", "max_availability"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.scaleway", "policy_mode", "enforced"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.scaleway", "policy_type", "low_latency"),
					resource.TestCheckResourceAttr("scaleway_instance_placement_group.scaleway", "policy_respected", "true"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_placement_group" "base" {
						policy_mode = "enforced"
						policy_type = "low_latency"
					}
			
					resource "scaleway_instance_placement_group" "scaleway" {}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstancePlacementGroupExists(tt, "scaleway_instance_placement_group.base"),
					testAccCheckScalewayInstancePlacementGroupExists(tt, "scaleway_instance_placement_group.scaleway"),
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

func testAccCheckScalewayInstancePlacementGroupExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = instanceAPI.GetPlacementGroup(&instance.GetPlacementGroupRequest{
			Zone:             zone,
			PlacementGroupID: ID,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayInstancePlacementGroupDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_instance_placement_group" {
				continue
			}

			instanceAPI, zone, ID, err := instanceAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = instanceAPI.GetPlacementGroup(&instance.GetPlacementGroupRequest{
				Zone:             zone,
				PlacementGroupID: ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("placement group (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}
		return nil
	}
}
