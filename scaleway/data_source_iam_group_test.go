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
					resource "scaleway_iam_group" "main_ds_basic" {
						name        = "test_data_source_basic"
						user_ids = []
						application_ids = []
					}
			
					data "scaleway_iam_group" "find_by_id_basic" {
						group_id 	= scaleway_iam_group.main_ds_basic.id
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}

					data "scaleway_iam_group" "find_by_name_basic" {
						name        = scaleway_iam_group.main_ds_basic.name
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_ds_basic"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id_basic", "name", "test_data_source_basic"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name_basic", "name", "test_data_source_basic"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_id_basic", "id", "scaleway_iam_group.main_ds_basic", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_name_basic", "id", "scaleway_iam_group.main_ds_basic", "id"),
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
					resource "scaleway_iam_application" "app00" {
						name = "app"
					}
					resource "scaleway_iam_group" "main_ds_mix" {
						name = "test_data_source_mix"
						application_ids = [
							scaleway_iam_application.app00.id,
						]
						user_ids = [
							"ce18cffd-e7c8-47f8-8de8-00e97e50a0d3",
							"255b63c2-b4de-4af6-9ed4-967f69d9dd85",
						]
					}
			
					data "scaleway_iam_group" "find_by_id_mix" {
						group_id 	= scaleway_iam_group.main_ds_mix.id
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}

					data "scaleway_iam_group" "find_by_name_mix" {
						name        = scaleway_iam_group.main_ds_mix.name
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_ds_mix"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id_mix", "name", "test_data_source_mix"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name_mix", "name", "test_data_source_mix"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_id_mix", "id", "scaleway_iam_group.main_ds_mix", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_name_mix", "id", "scaleway_iam_group.main_ds_mix", "id"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id_mix", "user_ids.0", "ce18cffd-e7c8-47f8-8de8-00e97e50a0d3"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id_mix", "user_ids.1", "255b63c2-b4de-4af6-9ed4-967f69d9dd85"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name_mix", "user_ids.0", "ce18cffd-e7c8-47f8-8de8-00e97e50a0d3"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name_mix", "user_ids.1", "255b63c2-b4de-4af6-9ed4-967f69d9dd85"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_id_mix", "application_ids.0", "scaleway_iam_application.app00", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_name_mix", "application_ids.0", "scaleway_iam_application.app00", "id"),
				),
			},
		},
	})
}
