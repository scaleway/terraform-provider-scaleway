package partner_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func partnerPassword(t *testing.T) string {
	t.Helper()

	if *acctest.UpdateCassettes {
		return "SecurePassword123!"
	}

	return "xxxxxxxx"
}

func TestAccPartnerOrganizationRequestAdminRoleAction_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccPartnerOrganizationRequestAdminRoleAction_Basic because actions are not yet supported on OpenTofu")
	}

	if *acctest.UpdateCassettes {
		t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })
	}

	tt, orgID := newPartnerTestTools(t)
	defer tt.Cleanup()

	email := partnerEmail(t)
	newAdminEmail := partnerEmail(t)
	password := partnerPassword(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsPartnerOrganizationLocked(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_partner_organization" "main" {
						email             = "%s"
						organization_name = "tf_test_admin_role_org"
						partner_id        = "%s"
						owner_firstname   = "John"
						owner_lastname    = "Doe"
						customer_id       = "customer-admin-role"

						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_partner_organization_request_admin_role.main]
							}
						}
					}

					action "scaleway_partner_organization_request_admin_role" "main" {
						config {
							organization_id = scaleway_partner_organization.main.id
							username        = "testuser"
							email           = "%s"
							password        = "%s"
						}
					}
				`, email, orgID, newAdminEmail, password),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_partner_organization.main", "id"),
				),
			},
		},
	})
}
