package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceAccountProject_Basic(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	orgID, orgIDExists := tt.Meta.scwClient.GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayAccountProjectDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_account_project "project" {
						name = "test-terraform-account-project"
					}

					data scaleway_account_project "by_name" {
						name = scaleway_account_project.project.name
						organization_id = "%s"
					}
					
					data scaleway_account_project "by_id" {
						project_id = scaleway_account_project.project.id
					}`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_account_project.by_name", "id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_account_project.by_name", "name", "scaleway_account_project.project", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_account_project.by_id", "id", "scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_account_project.by_id", "name", "scaleway_account_project.project", "name"),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceAccountProject_Default(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	orgID, orgIDExists := tt.Meta.scwClient.GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data scaleway_account_project "project" {
						name = "default"
						organization_id = "%s"
					}`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_account_project.project", "id"),
					resource.TestCheckResourceAttrSet("data.scaleway_account_project.project", "name"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data scaleway_account_project "project" {
						name = "default"
						organization_id = "%s"
					}

					data scaleway_account_project project2 {
						name = "default"
						organization_id = data.scaleway_account_project.project.id
					}`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_account_project.project", "id", "data.scaleway_account_project.project2", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_account_project.project", "name", "data.scaleway_account_project.project2", "name"),
				),
			},
		},
	})
}
