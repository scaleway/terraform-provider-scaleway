package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/stretchr/testify/require"
)

func TestAccListIAMPolicies_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccListIAMPolicies_Basic because list resources are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	ctx := t.Context()
	project, iamAPIKey, _, terminateFakeSideProject, err := acctest.CreateFakeIAMManager(tt)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.FakeSideProjectProviders(ctx, tt, project, iamAPIKey),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			func(_ *terraform.State) error {
				return terminateFakeSideProject()
			},
			testAccCheckIamPolicyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_policy" "policy1" {
						name         = "test-policy-list-1"
						description  = "Test policy 1"
						no_principal = true
						rule {
							organization_id      = "%s"
							permission_set_names = ["ContainerRegistryReadOnly"]
						}
						provider = side
					}
					`, project.OrganizationID),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_policy" "policy1" {
						name         = "test-policy-list-1"
						description  = "Test policy 1"
						no_principal = true
						rule {
							organization_id      = "%s"
							permission_set_names = ["ContainerRegistryReadOnly"]
						}
						provider = side
					}

					resource "scaleway_iam_policy" "policy2" {
						name         = "test-policy-list-2"
						description  = "Test policy 2"
						no_principal = true
						tags         = ["test-tag"]
						rule {
							organization_id      = "%[1]s"
							permission_set_names = ["AllProductsReadOnly"]
						}
						provider = side
					}
					`, project.OrganizationID),
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_iam_policy" "all" {
						provider = scaleway

						config {
							organization_id = "%s"
						}
					}
					`, project.OrganizationID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_iam_policy.all", 2),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_iam_policy" "by_tag" {
						provider = scaleway

						config {
							organization_id = "%s"
							tag              = "test-tag"
						}
					}
					`, project.OrganizationID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_policy.by_tag", 1),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_iam_policy" "by_editable" {
						provider = scaleway

						config {
							organization_id = "%s"
							editable         = true
						}
					}
					`, project.OrganizationID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("list.scaleway_iam_policy.by_editable", 1),
				},
			},
			{
				Query: true,
				Config: fmt.Sprintf(`
					list "scaleway_iam_policy" "by_ids" {
						provider = scaleway

						config {
							organization_id = "%s"
							policy_ids       = [scaleway_iam_policy.policy1.id]
						}
					}
					`, project.OrganizationID),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("list.scaleway_iam_policy.by_ids", 1),
				},
			},
		},
	})
}
