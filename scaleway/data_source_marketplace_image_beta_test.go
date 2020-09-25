package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceMarketplaceImageBeta_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "scaleway_marketplace_image_beta" "test1" {
	label = "ubuntu_focal"
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists("data.scaleway_marketplace_image_beta.test1"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image_beta.test1", "id", "fr-par-1/cf44b8f5-77e2-42ed-8f1e-09ed5bb028fc"),
				),
			},
		},
	})
}
