package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iot "github.com/scaleway/scaleway-sdk-go/api/iot/v1"
)

func TestAccScalewayIotDevice_Minimal(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		// Destruction is done via the hub destruction.
		CheckDestroy: testAccCheckScalewayIotHubDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
							allow_insecure = true
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "true"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
							allow_insecure = false
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_device" "default" {
							name = "default"
							hub_id = scaleway_iot_hub.minimal.id
							message_filters {
								publish {
									policy = "reject"
									topics = ["1", "2", "3"]
								}
								subscribe {
									policy = "accept"
									topics = ["4", "5", "6"]
								}
							}
						}
						resource "scaleway_iot_device" "empty" {
							name = "empty"
							hub_id = scaleway_iot_hub.minimal.id
							message_filters {
								publish { }
								subscribe { }
							}
						}
						resource "scaleway_iot_hub" "minimal" {
							name = "minimal"
							product_plan = "plan_shared"
						}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotDeviceExists(tt, "scaleway_iot_device.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "hub_id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_device.default", "name"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_insecure", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "allow_multiple_connections", "false"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "message_filters.0.publish.0.policy", "reject"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "message_filters.0.publish.0.topics.0", "1"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "message_filters.0.subscribe.0.policy", "accept"),
					resource.TestCheckResourceAttr("scaleway_iot_device.default", "message_filters.0.subscribe.0.topics.0", "4"),
				),
			},
		},
	})
}

func testAccCheckScalewayIotDeviceExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iotAPI, region, deviceID, err := iotAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = iotAPI.GetDevice(&iot.GetDeviceRequest{
			Region:   region,
			DeviceID: deviceID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
