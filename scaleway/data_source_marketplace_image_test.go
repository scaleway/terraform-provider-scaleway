package scaleway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccScalewayDataSourceMarketplaceImage_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_marketplace_image" "test1" {
						label = "ubuntu_focal"
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayInstanceImageExists(tt, "data.scaleway_marketplace_image.test1"),
					resource.TestCheckResourceAttr("data.scaleway_marketplace_image.test1", "label", "ubuntu_focal"),
				),
			},
		},
	})
}
