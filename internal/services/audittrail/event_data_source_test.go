package audittrail_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	audit_trail "github.com/scaleway/scaleway-sdk-go/api/audit_trail/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/stretchr/testify/require"
)

func TestAccDataSourceEvent_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	ctx := t.Context()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = dummyOrgID
	}

	auditTrailAPI := audit_trail.NewAPI(tt.Meta.ScwClient())

	project, iamAPIKey, _, terminateFakeSideProject, err := acctest.CreateFakeIAMManager(tt)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.FakeSideProjectProviders(ctx, tt, project, iamAPIKey),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			func(_ *terraform.State) error {
				return terminateFakeSideProject()
			},
			testAccCheckSecretDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}
				`, secretName, project.ID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "no_filter" {
					}
				`, secretName, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.no_filter", "events.#"),
				),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_project" {
						project_id = scaleway_secret.main.project_id
					}
				`, secretName, project.ID),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_project", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_type" {
						project_id = scaleway_secret.main.project_id
						resource_type = "secret_manager_secret"
					}
				`, secretName, project.ID),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_type", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_id" {
						project_id = scaleway_secret.main.project_id
						resource_id = split("/", scaleway_secret.main.id)[1]
					}
				`, secretName, project.ID),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_id", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_product" {
						project_id = scaleway_secret.main.project_id
						product_name = "secret-manager"
					}
				`, secretName, project.ID),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_product", orgID),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_audit_trail_event" "not_found" {
						project_id    = "%s"
						resource_id = "%s"
					}
				`, project.ID, uuid.New().String()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_audit_trail_event.not_found", "events.#", "0"),
				),
			},
		},
	})
}
