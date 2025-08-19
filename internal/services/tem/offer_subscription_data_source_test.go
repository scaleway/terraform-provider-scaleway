package tem_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceOfferSubscription_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()

	if !orgIDExists {
		orgID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
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
					resource.TestCheckResourceAttr("data.scaleway_tem_offer_subscription.test", "offer_name", "essential"),
				),
			},
		},
	})
}
