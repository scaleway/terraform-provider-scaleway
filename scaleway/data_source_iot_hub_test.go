package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceIotHub_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIotHubDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iot_hub" "test" {
							name = "test_iot_hub_datasource"
							product_plan = "plan_shared"
					}

					data "scaleway_iot_hub" "test" {
						name = scaleway_iot_hub.test.name
					}

					data "scaleway_iot_hub" "test2" {
						hub_id = scaleway_iot_hub.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIotHubExists(tt, "scaleway_iot_hub.test"),

					resource.TestCheckResourceAttr("data.scaleway_iot_hub.test", "name", "test_iot_hub_datasource"),
					resource.TestCheckResourceAttrSet("data.scaleway_iot_hub.test", "id"),

					resource.TestCheckResourceAttr("data.scaleway_iot_hub.test2", "name", "test_iot_hub_datasource"),
					resource.TestCheckResourceAttrSet("data.scaleway_iot_hub.test2", "id"),
				),
			},
		},
	})
}
