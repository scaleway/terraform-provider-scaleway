package scaleway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccScalewayDataSourceWebhosting_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayWebhostingDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				data "scaleway_webhosting_offer" "by_name" {
				  name = "lite"
				}

				resource "scaleway_webhosting" "main" {
				  offer_id     = data.scaleway_webhosting_offer.by_name.offer_id
				  email        = "hashicorp@scaleway.com"
				  domain       = "foobar.com"
				}`,
			},
			{
				Config: `
				data "scaleway_webhosting_offer" "by_name" {
				  name = "lite"
				}

				resource "scaleway_webhosting" "main" {
				  offer_id     = data.scaleway_webhosting_offer.by_name.offer_id
				  email        = "hashicorp@scaleway.com"
				  domain       = "foobar.com"
				}
				
				data "scaleway_webhosting" "by_domain" {
				  domain = "foobar.com"
				}

				data "scaleway_webhosting" "by_id" {
                  webhosting_id = "${scaleway_webhosting.main.id}"
				}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayWebhostingExists(tt, "scaleway_webhosting.main"),
					resource.TestCheckResourceAttrPair("data.scaleway_webhosting.by_domain", "webhosting_id", "scaleway_webhosting.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_webhosting.by_id", "domain", "scaleway_webhosting.main", "domain"),
				),
			},
		},
	})
}
