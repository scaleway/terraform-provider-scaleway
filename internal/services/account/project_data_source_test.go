package account_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

const dummyOrgID = "AB7BD9BF-E1BD-41E8-9F1D-F16A2E3F3925"

func TestAccDataSourceProject_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = dummyOrgID
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isProjectDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_account_project "project" {
						name = "tf-tests-terraform-account-project"
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

func TestAccDataSourceProject_Default(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = dummyOrgID
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
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

func TestAccDataSourceProject_Extract(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectID, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		t.Skip("no default project ID")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data scaleway_account_project "project" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_account_project.project", "id", projectID),
					resource.TestCheckResourceAttrSet("data.scaleway_account_project.project", "name"),
				),
			},
		},
	})
}

// This test was recorded using the hashicorp test account and expects projects from the 'terraform-provider-scaleway' organization
func TestAccDataSourceProject_List(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = dummyOrgID
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isProjectDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data scaleway_account_projects "projects" {
						organization_id = "%s"
					}`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.#", "4"),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.0.id", "105bdce1-64c0-48ab-899d-868455867ecf"),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.1.id", "c567f266-af4f-4da0-a35b-98c34086f991"),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.2.id", "fe479fbe-6cae-44c5-bb7a-7fc9f04acad5"),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.3.id", "f5375b18-7efc-4416-ab13-c42af955602c"),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.0.name", "default"),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.1.name", "Packer Plugin Scaleway"),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.2.name", "SDK Python"),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.3.name", "ansible"),
				),
			},
		},
	})
}
