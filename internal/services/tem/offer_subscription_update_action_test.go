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

	// Scale has a commitment period, so we cannot flip essential ↔ scale on the shared
	// cassette project. Upgrading to scale (from essential in the cassette) exercises
	// UpdateOfferSubscription once.
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
							offer_name = "scale"
						}
					}

					data "scaleway_tem_offer_subscription" "main" {
						depends_on = [terraform_data.trigger]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_tem_offer_subscription.main", "offer_name", "scale"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_offer_subscription.main", "project_id"),
				),
			},
		},
	})
}
