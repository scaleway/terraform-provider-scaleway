package scaleway

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_iam_group", &resource.Sweeper{
		Name: "scaleway_iam_group",
		F:    testSweepIamGroup,
	})
}

func testSweepIamGroup(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		api := iam.NewAPI(scwClient)

		// Requiring organization_id in list request is temporary
		organizationID := os.Getenv("DEFAULT_ORGANIZATION_ID")
		listApps, err := api.ListGroups(&iam.ListGroupsRequest{
			OrganizationID: &organizationID,
		})
		if err != nil {
			return fmt.Errorf("failed to list groups: %w", err)
		}
		for _, app := range listApps.Groups {
			err = api.DeleteGroup(&iam.DeleteGroupRequest{
				GroupID: app.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete group: %w", err)
			}
		}
		return nil
	})
}

func TestAccScalewayIamGroup_Basic(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_group" "main" {}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main"),
					resource.TestCheckResourceAttrSet("scaleway_iam_group.main", "name"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main", "description", ""),
				),
			},
			{
				Config: `
						resource "scaleway_iam_group" "main" {
							name = "iam_group_basic"
							description = "basic description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main", "name", "iam_group_basic"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main", "description", "basic description"),
				),
			},
		},
	})
}

func testAccCheckScalewayIamGroupExists(tt *TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		iamAPI := iamAPI(tt.Meta)

		_, err := iamAPI.GetGroup(&iam.GetGroupRequest{
			GroupID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("could not find group: %w", err)
		}

		return nil
	}
}

func testAccCheckScalewayIamGroupDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_iam_group" {
				continue
			}

			iamAPI := iamAPI(tt.Meta)

			_, err := iamAPI.GetGroup(&iam.GetGroupRequest{
				GroupID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("resource %s(%s) still exist", rs.Type, rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
