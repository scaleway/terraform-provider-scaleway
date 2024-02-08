package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
)

func TestAccScalewayDataSourceBaremetalOption_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBaremetalServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_baremetal_option" "by_name" {
						name = "Remote Access"
					}
					
					data "scaleway_baremetal_option" "by_id" {
						option_id = "931df052-d713-4674-8b58-96a63244c8e2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalOptionExists(tt, "data.scaleway_baremetal_option.by_id"),
					testAccCheckScalewayBaremetalOptionExists(tt, "data.scaleway_baremetal_option.by_name"),

					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_option.by_name", "name"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_option.by_name", "option_id"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_option.by_name", "manageable", "true"),

					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_option.by_id", "name"),
					resource.TestCheckResourceAttrSet("data.scaleway_baremetal_option.by_id", "option_id"),
					resource.TestCheckResourceAttr("data.scaleway_baremetal_option.by_id", "manageable", "true"),
				),
			},
		},
	})
}

func testAccCheckScalewayBaremetalOptionExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, ID, err := parseZonedID(rs.Primary.ID)
		if err != nil {
			return err
		}

		baremetalAPI := baremetal.NewAPI(tt.Meta.scwClient)
		_, err = baremetalAPI.GetOption(&baremetal.GetOptionRequest{
			OptionID: ID,
			Zone:     zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
