package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceIamGroup_Basic(t *testing.T) {
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
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_id", "name", "test-terraform"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_id", "id", "scaleway_iam_group.main", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_group" "main" {
						name        = "test-terraform"
					}

					data "scaleway_iam_group" "find_by_name" {
						name        = scaleway_iam_group.main.name
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main"),
					resource.TestCheckResourceAttr("data.scaleway_iam_group.find_by_name", "name", "test-terraform"),
					resource.TestCheckResourceAttrPair("data.scaleway_iam_group.find_by_name", "id", "scaleway_iam_group.main", "id"),
				),
			},
		},
	})
}
