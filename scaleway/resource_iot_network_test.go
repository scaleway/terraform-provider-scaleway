package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iot "github.com/scaleway/scaleway-sdk-go/api/iot/v1"
)

func TestAccScalewayIotNetwork_Minimal(t *testing.T) {
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
						resource "scaleway_iot_network" "default" {
							name   = "default"
							hub_id = scaleway_iot_hub.minimal.id
							type   = "sigfox"
						}
						resource "scaleway_iot_hub" "minimal" {
							name         = "minimal"
							product_plan = "plan_shared"
						}
						`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotNetworkExists(tt, "scaleway_iot_network.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "hub_id"),
					resource.TestCheckResourceAttr("scaleway_iot_network.default", "name", "default"),
					resource.TestCheckResourceAttr("scaleway_iot_network.default", "type", "sigfox"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "endpoint"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "secret"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "created_at"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_network" "default" {
							name   = "default"
							hub_id = scaleway_iot_hub.minimal.id
							type   = "rest"
						}
						resource "scaleway_iot_hub" "minimal" {
							name         = "minimal"
							product_plan = "plan_shared"
						}
						`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotNetworkExists(tt, "scaleway_iot_network.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "hub_id"),
					resource.TestCheckResourceAttr("scaleway_iot_network.default", "name", "default"),
					resource.TestCheckResourceAttr("scaleway_iot_network.default", "type", "rest"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "endpoint"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "secret"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "created_at"),
				),
			},
			{
				Config: `
						resource "scaleway_iot_network" "default" {
							name         = "default"
							hub_id       = scaleway_iot_hub.minimal.id
							type         = "rest"
							topic_prefix = "foo/bar"
						}
						resource "scaleway_iot_hub" "minimal" {
							name         = "minimal"
							product_plan = "plan_shared"
						}
						`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.minimal"),
					testAccCheckScalewayIotNetworkExists(tt, "scaleway_iot_network.default"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "id"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "hub_id"),
					resource.TestCheckResourceAttr("scaleway_iot_network.default", "name", "default"),
					resource.TestCheckResourceAttr("scaleway_iot_network.default", "type", "rest"),
					resource.TestCheckResourceAttr("scaleway_iot_network.default", "topic_prefix", "foo/bar"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "endpoint"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "secret"),
					resource.TestCheckResourceAttrSet("scaleway_iot_network.default", "created_at"),
				),
			},
		},
	})
}

func testAccCheckScalewayIotNetworkExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iotAPI, region, networkID, err := iotAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = iotAPI.GetNetwork(&iot.GetNetworkRequest{
			Region:    region,
			NetworkID: networkID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
