package interlink_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceInterlinkPartners_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_partners" "all" {
					  region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_interlink_partners.all", "partners.#"),
					resource.TestMatchResourceAttr("data.scaleway_interlink_partners.all", "partners.0.srn", regexp.MustCompile(`^srn://interlink\..+/regions/.+/partners/.+$`)),
				),
			},
		},
	})
}
