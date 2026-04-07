package interlink_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceInterlinkPop_ByName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_pop" "by_name" {
					  name   = "DC2"
					  region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_interlink_pop.by_name", "id"),
					resource.TestCheckResourceAttr("data.scaleway_interlink_pop.by_name", "name", "DC2"),
					resource.TestCheckResourceAttr("data.scaleway_interlink_pop.by_name", "hosting_provider_name", "OpCore"),
					resource.TestCheckResourceAttr("data.scaleway_interlink_pop.by_name", "city", "Vitry-sur-Seine"),
				),
			},
		},
	})
}

func TestAccDataSourceInterlinkPop_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_pop" "by_name" {
					  name   = "DC2"
					  region = "fr-par"
					}

					data "scaleway_interlink_pop" "by_id" {
					  pop_id = data.scaleway_interlink_pop.by_name.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_interlink_pop.by_id", "id",
						"data.scaleway_interlink_pop.by_name", "id"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_interlink_pop.by_id", "name",
						"data.scaleway_interlink_pop.by_name", "name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_interlink_pop.by_id", "hosting_provider_name",
						"data.scaleway_interlink_pop.by_name", "hosting_provider_name"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_interlink_pop.by_id", "city",
						"data.scaleway_interlink_pop.by_name", "city"),
				),
			},
		},
	})
}
