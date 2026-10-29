package interlink_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceInterlinkPops_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_pops" "all" {
					  region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_interlink_pops.all", "pops.#"),
				),
			},
		},
	})
}

func TestAccDataSourceInterlinkPops_ByProviderName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_pops" "by_provider_name" {
					  hosting_provider_name = "OpCore"
					  region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_interlink_pops.by_provider_name", "pops.#"),
					resource.TestCheckResourceAttr("data.scaleway_interlink_pops.by_provider_name", "pops.0.hosting_provider_name", "OpCore"),
				),
			},
		},
	})
}
