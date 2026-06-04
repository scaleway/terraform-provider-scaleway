package interlink_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceInterlinkLink_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckInterlinkLinkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_pop" "pop" {
						name   = "Telehouse TH2"
						region = "fr-par"
					}

					data "scaleway_interlink_partner" "partner" {
						name   = "FranceIX"
						region = "fr-par"
					}

					resource "scaleway_interlink_link" "main" {
						name           = "tf-test-interlink-link-ds"
						pop_id         = data.scaleway_interlink_pop.pop.id
						partner_id     = data.scaleway_interlink_partner.partner.id
						bandwidth_mbps = 50
						region         = "fr-par"
					}
				`,
			},
			{
				Config: `
					data "scaleway_interlink_pop" "pop" {
						name   = "Telehouse TH2"
						region = "fr-par"
					}

					data "scaleway_interlink_partner" "partner" {
						name   = "FranceIX"
						region = "fr-par"
					}

					resource "scaleway_interlink_link" "main" {
						name           = "tf-test-interlink-link-ds"
						pop_id         = data.scaleway_interlink_pop.pop.id
						partner_id     = data.scaleway_interlink_partner.partner.id
						bandwidth_mbps = 50
						region         = "fr-par"
					}

					data "scaleway_interlink_link" "by_name" {
						name = scaleway_interlink_link.main.name
					}

					data "scaleway_interlink_link" "by_id" {
						link_id = scaleway_interlink_link.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInterlinkLinkExists(tt, "scaleway_interlink_link.main"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_interlink_link.by_name", "name",
						"scaleway_interlink_link.main", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_interlink_link.by_name", "bandwidth_mbps",
						"scaleway_interlink_link.main", "bandwidth_mbps"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_interlink_link.by_id", "link_id",
						"scaleway_interlink_link.main", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_interlink_link.by_id", "name",
						"scaleway_interlink_link.main", "name"),
				),
			},
		},
	})
}
