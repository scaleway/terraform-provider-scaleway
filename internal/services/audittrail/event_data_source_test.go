package audittrail_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

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
						resource_type = "%s"
					}
				`, secretName, project.ID, resourceType),
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
						product_name = "%s"
					}
				`, secretName, project.ID, productName),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_product", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_service" {
						project_id = scaleway_secret.main.project_id
						service_name = "%s"
					}
				`, secretName, project.ID, serviceName),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_service", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_method" {
						project_id = scaleway_secret.main.project_id
						method_name = "%s"
					}
				`, secretName, project.ID, methodCreate),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_method", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_method" {
						project_id = scaleway_secret.main.project_id
						method_name = "%s"
					}

					data "scaleway_audit_trail_event" "by_principal" {
						project_id = scaleway_secret.main.project_id
						principal_id = data.scaleway_audit_trail_event.by_method.events.0.principal_id
					}
				`, secretName, project.ID, methodCreate),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_principal", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_method" {
						project_id = scaleway_secret.main.project_id
						method_name = "%s"
					}

					data "scaleway_audit_trail_event" "by_ip" {
						project_id = scaleway_secret.main.project_id
						source_ip = data.scaleway_audit_trail_event.by_method.events.0.source_ip
					}
				`, secretName, project.ID, methodCreate),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_ip", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "by_status" {
						project_id = scaleway_secret.main.project_id
						status = 200
					}
				`, secretName, project.ID),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.by_status", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "recorded_after" {
						project_id = scaleway_secret.main.project_id
						recorded_after = "%s"
					}
				`, secretName, project.ID, time.Now().Add(-time.Minute*10).Format(time.RFC3339)),
				Check: createEventDataSourceChecks("data.scaleway_audit_trail_event.recorded_after", orgID),
			},
			{
				PreConfig: func() { waitForAuditTrailEvents(t, ctx, auditTrailAPI, project) },
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%s"
					  description = "DataSourceAuditTrail test description"
					  project_id  = "%s"
					}

					data "scaleway_audit_trail_event" "recorded_before" {
						project_id = scaleway_secret.main.project_id
						recorded_before = "%s"
					}
				`, secretName, project.ID, time.Now().Add(-time.Minute*30).Format(time.RFC3339)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_audit_trail_event.recorded_before", "events.#", "0"),
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

					data "scaleway_audit_trail_event" "order_by" {
						project_id = scaleway_secret.main.project_id
						order_by = "recorded_at_asc"
					}
				`, secretName, project.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_audit_trail_event.order_by", "events.#"),
				),
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

func TestAccDataSourceEvent_Warning(t *testing.T) {
	// Test that a resource_type that is not yet supported on the
	// provider only raises a warning before calling the API
	// anyway (it could exist on API side)

	// NOTE: Currently, we cannot programmatically assert that a warning was emitted
	// during the test step. This needs support from the testing framework:
	// https://github.com/hashicorp/terraform-plugin-testing/issues/69
	// Once implemented, we should add a check like:
	//   ExpectWarning: regexp.MustCompile(`expected resourceType to be one of [\"unknown_type\" \"secm_secret\" \"secm_secret_version\" \"kube_cluster\" \"kube_pool\" \"kube_node\" \"kube_acl\" \"keym_key\" \"iam_user\" \"iam_application\" \"iam_group\" \"iam_policy\" \"iam_api_key\" \"iam_ssh_key\" \"iam_rule\" \"iam_saml\" \"iam_saml_certificate\" \"secret_manager_secret\" \"secret_manager_version\" \"key_manager_key\" \"account_user\" \"account_organization\" \"account_project\" \"instance_server\" \"instance_placement_group\" \"instance_security_group\" \"instance_volume\" \"instance_snapshot\" \"instance_image\" \"apple_silicon_server\" \"baremetal_server\" \"baremetal_setting\" \"ipam_ip\" \"sbs_volume\" \"sbs_snapshot\" \"load_balancer_lb\" \"load_balancer_ip\" \"load_balancer_frontend\" \"load_balancer_backend\" \"load_balancer_route\" \"load_balancer_acl\" \"load_balancer_certificate\" \"sfs_filesystem\" \"vpc_private_network\"], got a_new_resource_type`)
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_audit_trail_event" "unknown_resource_type" {
						recorded_after = "%s"
						resource_type = "a_new_resource_type"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_audit_trail_event.unknown_resource_type", "events.#", "0"),
				),
				// In this test, we still expect a 400 from the API since `a_new_resource_type`
				// does not actually exist on API side.
				ExpectError: regexp.MustCompile(`400 Bad Request`),
			},
		},
	})
}
