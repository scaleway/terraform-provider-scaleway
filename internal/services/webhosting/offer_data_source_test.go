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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_webhosting_offer" "by_name" {
						name = "professional"
					}

					data "scaleway_webhosting_offer" "by_id" {
						offer_id = "de2426b4-a9e9-11ec-b909-0242ac120002"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_webhosting_offer.by_id", "id"),
					resource.TestCheckResourceAttr("data.scaleway_webhosting_offer.by_id", "name", "professional"),
					resource.TestCheckResourceAttrSet("data.scaleway_webhosting_offer.by_name", "id"),
					resource.TestCheckResourceAttr("data.scaleway_webhosting_offer.by_name", "name", "professional"),
					resource.TestCheckResourceAttr("data.scaleway_webhosting_offer.by_name", "offer.0.name", "professional"),
					resource.TestCheckResourceAttr("data.scaleway_webhosting_offer.by_name", "offer.0.price", "€ 18.99"),
				),
			},
		},
	})
}
