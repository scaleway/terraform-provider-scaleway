package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iot "github.com/scaleway/scaleway-sdk-go/api/iot/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_iot_hub", &resource.Sweeper{
		Name: "scaleway_iot_hub",
		F:    testSweepIotHub,
	})
}

func testSweepIotHub(_ string) error {
	return sweepRegions(scw.AllRegions, func(scwClient *scw.Client, region scw.Region) error {
		iotAPI := iot.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the iot hub in (%s)", region)
		listHubs, err := iotAPI.ListHubs(&iot.ListHubsRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			l.Debugf("sweeper: destroying the iot hub in (%s)", region)
			return fmt.Errorf("error listing hubs in (%s) in sweeper: %s", region, err)
		}

		deleteDevices := true
		for _, hub := range listHubs.Hubs {
			err := iotAPI.DeleteHub(&iot.DeleteHubRequest{
				HubID:         hub.ID,
				Region:        hub.Region,
				DeleteDevices: &deleteDevices,
			})
			if err != nil {
				return fmt.Errorf("error deleting hub in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayIotHub_Minimal(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIotHubDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "product_plan", "plan_shared"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "status", iot.HubStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_iot_hub.minimal", "endpoint"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "device_count", "0"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "connected_device_count", "0"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
							enabled = false
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "status", iot.HubStatusDisabled.String()),
					resource.TestCheckResourceAttrSet("scaleway_iot_hub.minimal", "endpoint"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "device_count", "0"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "connected_device_count", "0"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "enabled", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "product_plan", "plan_shared"),
				),
			},
		},
	})
}

func TestAccScalewayIotHub_Dedicated(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIotHubDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_dedicated"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "product_plan", "plan_dedicated"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "status", iot.HubStatusReady.String()),
					resource.TestCheckResourceAttrSet("scaleway_iot_hub.minimal", "endpoint"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "device_count", "0"),
					resource.TestCheckResourceAttr("scaleway_iot_hub.minimal", "connected_device_count", "0"),
				),
			},
		},
	})
}

func testAccCheckScalewayIotHubDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_iot_hub" {
				continue
			}

			iotAPI, region, hubID, err := iotAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = iotAPI.GetHub(&iot.GetHubRequest{
				Region: region,
				HubID:  hubID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("hub (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return fmt.Errorf("unexpected error when getting hub (%s) : %s", rs.Primary.ID, err)
			}
		}
		return nil
	}
}

func testAccCheckScalewayIotHubExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iotAPI, region, hubID, err := iotAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = iotAPI.GetHub(&iot.GetHubRequest{
			Region: region,
			HubID:  hubID,
		})
		if err != nil {
			return fmt.Errorf("error getting hub (%s) : %s", rs.Primary.ID, err)
		}

		return nil
	}
}
