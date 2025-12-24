package webhosting_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceOffer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_webhosting_offer" "by_name" {
						name = "essential"
					}

					data "scaleway_webhosting_offer" "by_id" {
						offer_id = "b88b9cf9-7f35-4d36-aa4c-0de8cf301f87"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_webhosting_offer.by_id", "id"),
					resource.TestCheckResourceAttr("data.scaleway_webhosting_offer.by_id", "name", "essential"),
					resource.TestCheckResourceAttrSet("data.scaleway_webhosting_offer.by_name", "id"),
					resource.TestCheckResourceAttr("data.scaleway_webhosting_offer.by_name", "name", "essential"),
					resource.TestCheckResourceAttr("data.scaleway_webhosting_offer.by_name", "offer.0.name", "essential"),
					resource.TestCheckResourceAttr("data.scaleway_webhosting_offer.by_name", "offer.0.price", "€ 9.99"),
				),
			},
		},
	})
}
