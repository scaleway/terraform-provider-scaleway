package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccScalewayDataSourceBootscript_Filtered(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayBootscriptFilterConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBootscriptID("data.scaleway_bootscript.debug"),
					resource.TestCheckResourceAttr("data.scaleway_bootscript.debug", "architecture", "arm"),
					resource.TestCheckResourceAttr("data.scaleway_bootscript.debug", "public", "true"),
				),
			},
		},
	})
}

func testAccCheckBootscriptID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find bootscript data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("bootscript data source ID not set")
		}

		scaleway := testAccProvider.Meta().(*Meta).deprecatedClient
		_, err := scaleway.GetBootscript(rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}

const testAccCheckScalewayBootscriptFilterConfig = `
data "scaleway_bootscript" "debug" {
  architecture = "arm"
  name_filter = "rescue"
}
`
