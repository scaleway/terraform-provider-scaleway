package interlink_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceInterlinkPartner_ByName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_interlink_partner" "by_name" {
					  name   = "FreePro"
					  region = "fr-par"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_interlink_partner.by_name", "id"),
					resource.TestCheckResourceAttr("data.scaleway_interlink_partner.by_name", "name", "FreePro"),
					resource.TestCheckResourceAttrSet("data.scaleway_interlink_partner.by_name", "contact_email"),
				),
			},
		},
	})
}
