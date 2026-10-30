package tem_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionTEMUpdateOfferSubscription_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionTEMUpdateOfferSubscription_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "terraform_data" "trigger" {
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_tem_update_offer_subscription.main]
							}
						}
					}

					action "scaleway_tem_update_offer_subscription" "main" {
						config {
							offer_name = "essential"
						}
					}

					data "scaleway_tem_offer_subscription" "main" {
						depends_on = [terraform_data.trigger]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_tem_offer_subscription.main", "offer_name", "essential"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_offer_subscription.main", "project_id"),
				),
			},
		},
	})
}
