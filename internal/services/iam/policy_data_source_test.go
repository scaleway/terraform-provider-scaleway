package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/stretchr/testify/require"
)

func TestAccDataSourcePolicy_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	ctx := t.Context()
	project, iamAPIKey, iamPolicy, terminateFakeSideProject, err := acctest.CreateFakeIAMManager(tt)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
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
					data "scaleway_iam_policy" "by_name" {
						name = "%s"
					}
					
					data "scaleway_iam_policy" "by_id" {
						policy_id = "%s"
					}`, iamPolicy.Name, iamPolicy.ID),
				Check: resource.ComposeTestCheckFunc(
					// Check by name
					testAccCheckIamPolicyExists(tt, "data.scaleway_iam_policy.by_name"),
					resource.TestCheckResourceAttr("data.scaleway_iam_policy.by_name", "name", iamPolicy.Name),
					resource.TestCheckResourceAttr("data.scaleway_iam_policy.by_name", "application_id", *iamPolicy.ApplicationID),
					resource.TestCheckResourceAttr("data.scaleway_iam_policy.by_name", "rule.0.organization_id", project.OrganizationID),
					// Check by id
					testAccCheckIamPolicyExists(tt, "data.scaleway_iam_policy.by_id"),
					resource.TestCheckResourceAttr("data.scaleway_iam_policy.by_id", "name", iamPolicy.Name),
					resource.TestCheckResourceAttr("data.scaleway_iam_policy.by_id", "application_id", *iamPolicy.ApplicationID),
					resource.TestCheckResourceAttr("data.scaleway_iam_policy.by_id", "rule.0.organization_id", project.OrganizationID),

					// Ensure both refer to the same policy
					resource.TestCheckResourceAttrPair(
						"data.scaleway_iam_policy.by_name",
						"id",
						"data.scaleway_iam_policy.by_id",
						"id",
					),
				),
			},
		},
	})
}
