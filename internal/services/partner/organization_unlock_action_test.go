package partner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	partnerSDK "github.com/scaleway/scaleway-sdk-go/api/partner/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func TestAccPartnerOrganizationUnlockAction_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccPartnerOrganizationUnlockAction_Basic because actions are not yet supported on OpenTofu")
	}

	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt, orgID := newPartnerTestTools(t)
	defer tt.Cleanup()

	email := partnerEmail(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsPartnerOrganizationLocked(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_partner_organization" "main" {
						email             = "%s"
						organization_name = "tf_test_unlock_org"
						partner_id        = "%s"
						owner_firstname   = "John"
						owner_lastname    = "Doe"
						customer_id       = "customer-unlock"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_partner_organization_lock.main]
							}
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_partner_organization_unlock.main]
							}
						}
					}

					action "scaleway_partner_organization_lock" "main" {
						config {
							organization_id = scaleway_partner_organization.main.id
						}
					}

					action "scaleway_partner_organization_unlock" "main" {
						config {
							organization_id = scaleway_partner_organization.main.id
						}
					}
				`, email, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_partner_organization.main", "id"),
					isPartnerOrganizationUnlockedByPartner(tt, "scaleway_partner_organization.main"),
				),
			},
		},
	})
}

func isPartnerOrganizationUnlockedByPartner(tt *acctest.TestTools, orgResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[orgResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", orgResourceName)
		}

		partnerAPI := partnerSDK.NewAPI(meta.ExtractScwClient(tt.Meta))

		organization, err := partnerAPI.GetOrganization(&partnerSDK.GetOrganizationRequest{
			OrganizationID: rs.Primary.ID,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return fmt.Errorf("failed to get organization: %w", err)
		}

		if organization.LockedBy == partnerSDK.OrganizationLockedByPartner {
			return fmt.Errorf("organization %s is still locked by partner after unlock action", rs.Primary.ID)
		}

		return nil
	}
}
