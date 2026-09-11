package partner_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	identitycheck "github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest/identity"
)

func TestAccListPartnerOrganizations_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListPartnerOrganizations_Basic because list resources are not yet supported on OpenTofu")
	}

	tt, orgID := newPartnerTestTools(t)
	defer tt.Cleanup()

	email := uuid.NewString() + "@test.test"
	orgName := "tf_test_partner_list_" + uuid.NewString()[:8]
	customerID := "customer-list-" + uuid.NewString()[:8]

	resourceName := "scaleway_partner_organization.main"
	identity := identitycheck.Identity()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_partner_organization" "main" {
						email             = "%s"
						organization_name = "%s"
						partner_id        = "%s"
						owner_firstname   = "John"
						owner_lastname    = "Doe"
						customer_id       = "%s"
					}
				`, email, orgName, orgID, customerID),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_partner_organization" "all" {
						provider = scaleway

						config {
							customer_id = "%s"
						}
					}
				`, customerID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					identitycheck.ExpectIdentityFunc("scaleway_partner_organization.all", identity.Checks()),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_partner_organization" "all" {
						provider = scaleway
						include_resource = true

						config {
							customer_id = "%s"
							partner_id  = "%s"
						}
					}
				`, customerID, orgID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_partner_organization.all", 1),
					identitycheck.ExpectIdentityFunc("scaleway_partner_organization.all", identity.Checks()),
					querycheck.ExpectResourceKnownValues(
						"scaleway_partner_organization.all",
						identitycheck.FilterByResourceIdentityFunc(identity.Checks()),
						[]querycheck.KnownValueCheck{
							identitycheck.KnownValueCheck(tfjsonpath.New("partner_id"), knownvalue.StringExact(orgID)),
						},
					),
				},
			},
		},
	})
}
