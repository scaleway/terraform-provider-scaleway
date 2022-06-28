package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceIamApplication_Basic(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIamApplicationDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_application" "app_ds_basic" {
						name        = "test_data_source_basic"
					}
			
					data "scaleway_iam_application" "find_by_id_basic" {
						application_id 	= scaleway_iam_application.app_ds_basic.id
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}
					data "scaleway_iam_application" "find_by_name_basic" {
						name        = scaleway_iam_application.app_ds_basic.name
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamApplicationExists(tt, "scaleway_iam_application.app_ds_basic"),
					resource.TestCheckResourceAttr("data.scaleway_iam_application.find_by_id_basic", "name", "test_data_source_basic"),
					resource.TestCheckResourceAttr("data.scaleway_iam_application.find_by_name_basic", "name", "test_data_source_basic"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_application.find_by_id_basic", "id", "scaleway_iam_application.app_ds_basic", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_application.find_by_name_basic", "id", "scaleway_iam_application.app_ds_basic", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_application" "app_ds_basic" {
						name        = "test_data_source_basic_renamed"
						description = "test_data_source_basic_description"
					}
			
					data "scaleway_iam_application" "find_by_id_basic" {
						application_id 	= scaleway_iam_application.app_ds_basic.id
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}
					data "scaleway_iam_application" "find_by_name_basic" {
						name        = scaleway_iam_application.app_ds_basic.name
						organization_id = "08555df8-bb26-43bc-b749-1b98c5d02343"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamApplicationExists(tt, "scaleway_iam_application.app_ds_basic"),
					resource.TestCheckResourceAttr("data.scaleway_iam_application.find_by_id_basic", "name", "test_data_source_basic_renamed"),
					resource.TestCheckResourceAttr("data.scaleway_iam_application.find_by_name_basic", "name", "test_data_source_basic_renamed"),
					resource.TestCheckResourceAttr("data.scaleway_iam_application.find_by_id_basic", "description", "test_data_source_basic_description"),
					resource.TestCheckResourceAttr("data.scaleway_iam_application.find_by_name_basic", "description", "test_data_source_basic_description"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_application.find_by_id_basic", "id", "scaleway_iam_application.app_ds_basic", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_application.find_by_name_basic", "id", "scaleway_iam_application.app_ds_basic", "id"),
				),
			},
		},
	})
}
