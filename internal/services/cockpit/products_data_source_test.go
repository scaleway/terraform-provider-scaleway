package cockpit_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccCockpitProducts_DataSource_Basic(t *testing.T) {
	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_cockpit_products" "main" {
						region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_products.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_cockpit_products.main", "region", "fr-par"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_products.main", "products.#"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_products.main", "names.#"),
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_products.main", "products.#", "data.scaleway_cockpit_products.main", "names.#"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_products.main", "products.0.name"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_products.main", "products.0.display_name"),
					resource.TestCheckResourceAttrSet("data.scaleway_cockpit_products.main", "products.0.family_name"),
					resource.TestCheckResourceAttrPair("data.scaleway_cockpit_products.main", "products.0.name", "data.scaleway_cockpit_products.main", "names.0"),
				),
			},
		},
	})
}
