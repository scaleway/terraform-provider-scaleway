package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceIamGroup_Basic(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIamGroupDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_group" "main" {
						name        = "test-terraform"
					}
			
					data "scaleway_iam_group" "find_by_id" {
						group_id 	= scaleway_iam_group.main.id
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}

					data "scaleway_iam_group" "find_by_name" {
						name        = scaleway_iam_group.main.name
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id", "name", "test-terraform"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name", "name", "test-terraform"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_id", "id", "scaleway_iam_group.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_name", "id", "scaleway_iam_group.main", "id"),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceIamGroup_UsersAndApplications(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIamGroupDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_application" "app01" {
						name = "first app"
					}
					resource "scaleway_iam_group" "main" {
						name = "test-terraform"
						application_ids = [
							scaleway_iam_application.app01.id,
						]
						user_ids = [
							"ce18cffd-e7c8-47f8-8de8-00e97e50a0d3",
							"255b63c2-b4de-4af6-9ed4-967f69d9dd85",
						]
					}
			
					data "scaleway_iam_group" "find_by_id" {
						group_id 	= scaleway_iam_group.main.id
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}

					data "scaleway_iam_group" "find_by_name" {
						name        = scaleway_iam_group.main.name
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id", "name", "test-terraform"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name", "name", "test-terraform"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_id", "id", "scaleway_iam_group.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_name", "id", "scaleway_iam_group.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id", "user_ids.0", "ce18cffd-e7c8-47f8-8de8-00e97e50a0d3"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id", "user_ids.1", "255b63c2-b4de-4af6-9ed4-967f69d9dd85"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name", "user_ids.0", "ce18cffd-e7c8-47f8-8de8-00e97e50a0d3"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name", "user_ids.1", "255b63c2-b4de-4af6-9ed4-967f69d9dd85"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_id", "application_ids.0", "scaleway_iam_application.app01", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_name", "application_ids.0", "scaleway_iam_application.app01", "id"),
				),
			},
		},
	})
}
