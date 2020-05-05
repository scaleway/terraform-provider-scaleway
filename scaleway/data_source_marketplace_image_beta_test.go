package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccScalewayDataSourceMarketplaceImageBeta_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "scaleway_marketplace_image_beta" "test1" {
	label = "ubuntu-focal"
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists("data.scaleway_marketplace_image_beta.test1"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image_beta.test1", "id", "fr-par-1/f974feac-abae-4365-b988-8ec7d1cec10d"),
				),
			},
		},
	})
}
