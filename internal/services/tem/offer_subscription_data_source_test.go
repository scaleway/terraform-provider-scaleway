package tem_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceOfferSubscription_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()

	if !orgIDExists {
		orgID = "00000000-0000-0000-0000-000000000000"
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`

					data scaleway_account_project "project" {
						name = "default"
						organization_id = "%s"
					}
				
					data "scaleway_tem_offer_subscription" "test" {
						project_id = data.scaleway_account_project.project.id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_tem_offer_subscription.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_offer_subscription.test", "organization_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_offer_subscription.test", "offer_name"),
				),
			},
			{
				Config: fmt.Sprintf(`

					data "scaleway_tem_offer_subscription" "test" {
						organization_id = "%s"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_tem_offer_subscription.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_offer_subscription.test", "organization_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_offer_subscription.test", "offer_name"),
				),
			},
		},
	})
}
