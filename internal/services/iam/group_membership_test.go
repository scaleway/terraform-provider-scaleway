package iam_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
)

func TestAccGroupMembership_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamGroupDestroy(tt),
			testAccCheckIamApplicationDestroy(tt),
		), Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_iam_group main {
						name = "tf-tests-iam-group-membership-basic"
						external_membership = true
					}

					resource scaleway_iam_application main {
						name = "tf-tests-iam-group-membership-basic"
					}

					resource scaleway_iam_group_membership main {
						group_id = scaleway_iam_group.main.id
						application_ids = [scaleway_iam_application.main.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					checkEntityInGroup(tt, "scaleway_iam_group_membership.main", "scaleway_iam_application.main"),
					acctest.CheckResourceAttrUUID("scaleway_iam_group_membership.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_iam_group main {
						name = "tf-tests-iam-group-membership-basic"
						external_membership = true
					}

					resource scaleway_iam_application main {
						name = "tf-tests-iam-group-membership-basic"
					}

					resource scaleway_iam_group_membership main {
						group_id = scaleway_iam_group.main.id
						application_ids = [scaleway_iam_application.main.id]
					}

					resource scaleway_iam_group_membership import {
						group_id = scaleway_iam_group.main.id
						application_ids = [scaleway_iam_application.main.id]
					}
				`,
				ImportState:  true,
				ResourceName: "scaleway_iam_group_membership.import",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					groupID := state.RootModule().Resources["scaleway_iam_group.main"].Primary.ID
					applicationID := state.RootModule().Resources["scaleway_iam_application.main"].Primary.ID

					return iam.SetGroupMembershipResourceID(groupID, nil, []string{applicationID}), nil
				},
				ImportStatePersist: true,
			},
			{
				Config: `
					resource scaleway_iam_group main {
						name = "tf-tests-iam-group-membership-basic"
						external_membership = true
					}

					resource scaleway_iam_application main {
						name = "tf-tests-iam-group-membership-basic"
					}

					resource scaleway_iam_group_membership main {
						group_id = scaleway_iam_group.main.id
						application_ids = [scaleway_iam_application.main.id]
					}

					resource scaleway_iam_group_membership import {
						group_id = scaleway_iam_group.main.id
						application_ids = [scaleway_iam_application.main.id]
					}
				`,
				PlanOnly: true,
			},
			{
				Config: `
					resource scaleway_iam_group main {
						name = "tf-tests-iam-group-membership-user"
						external_membership = true
					}

					resource scaleway_iam_user foo {
						email = "foo@scaleway.com"
						username = "foo"
					}
					resource scaleway_iam_user bar {
						email = "bar@scaleway.com"
						username = "bar"
					}

					resource scaleway_iam_group_membership main {
						group_id = scaleway_iam_group.main.id
						user_ids = [scaleway_iam_user.foo.id, scaleway_iam_user.bar.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					checkEntityInGroup(tt, "scaleway_iam_group_membership.main", "scaleway_iam_user.foo"),
					checkEntityInGroup(tt, "scaleway_iam_group_membership.main", "scaleway_iam_user.bar"),
					acctest.CheckResourceAttrUUID("scaleway_iam_group_membership.main", "id"),
				),
			},
		},
	})
}

func TestAccGroupMembership_MultipleEntities(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamGroupDestroy(tt),
		), Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_iam_group main {
						name = "tf-tests-iam-group-membership-multiple-entities"
						external_membership = true
					}

					resource scaleway_iam_user foo {
						email = "foo@scaleway.com"
						username = "foo"
					}
					resource scaleway_iam_user bar {
						email = "bar@scaleway.com"
						username = "bar"
					}

					resource scaleway_iam_application app1 {
						name = "tf-tests-iam-group-membership-multiple-entities-app1"
					}

					resource scaleway_iam_application app2 {
						name = "tf-tests-iam-group-membership-multiple-entities-app2"
					}

					resource scaleway_iam_group_membership main {
						group_id = scaleway_iam_group.main.id
						user_ids = [scaleway_iam_user.foo.id, scaleway_iam_user.bar.id]
						application_ids = [scaleway_iam_application.app1.id, scaleway_iam_application.app2.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					checkEntityInGroup(tt, "scaleway_iam_group_membership.main", "scaleway_iam_user.foo"),
					checkEntityInGroup(tt, "scaleway_iam_group_membership.main", "scaleway_iam_user.bar"),
					checkEntityInGroup(tt, "scaleway_iam_group_membership.main", "scaleway_iam_application.app1"),
					checkEntityInGroup(tt, "scaleway_iam_group_membership.main", "scaleway_iam_application.app2"),
					acctest.CheckResourceAttrUUID("scaleway_iam_group_membership.main", "id"),
				),
			},
		},
	})
}

func checkEntityInGroup(tt *acctest.TestTools, groupName string, entityName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// sanity check if the resource exists
		group, ok := state.RootModule().Resources[groupName]
		if !ok {
			return fmt.Errorf("resource not found: %s", groupName)
		}

		// sanity check if the entity exists
		entity, ok := state.RootModule().Resources[entityName]
		if !ok {
			return fmt.Errorf("resource not found: %s", entityName)
		}

		// get entity id from the State
		entityID := entity.Primary.ID
		entityKind := entity.Type

		groupID, _, err := iam.ExpandGroupMembershipResourceID(group.Primary.ID)
		if err != nil {
			return err
		}

		// get the group details from the API
		api := iam.NewAPI(tt.Meta)
		groupDetails, err := api.GetGroup(&iamSDK.GetGroupRequest{
			GroupID: groupID,
		})
		if err != nil {
			return err
		}

		// check if the entity is in the group
		switch entityKind {
		case "scaleway_iam_user":
			if !slices.Contains(groupDetails.UserIDs, entityID) {
				return errors.New("entity not found in group")
			}
		case "scaleway_iam_application":
			if !slices.Contains(groupDetails.ApplicationIDs, entityID) {
				return errors.New("entity not found in group")
			}
		default:
			return fmt.Errorf("unknown entity kind: %s", entityKind)
		}

		return nil
	}
}
