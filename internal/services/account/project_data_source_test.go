package account_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

var (
	ErrExpectedAtLeastOneProject = errors.New("expected at least one project")
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

func TestAccDataSourceProject_List(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		orgID = dummyOrgID
	}

	projectID, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		t.Skip("no default project ID")
	}

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttrSet("data.scaleway_account_projects.projects", "projects.#"),
					resource.TestCheckResourceAttrWith("data.scaleway_account_projects.projects", "projects.#", func(value string) error {
						count, err := strconv.Atoi(value)
						if err != nil {
							return err
						}

						if count < 1 {
							return fmt.Errorf("%w, got %d", ErrExpectedAtLeastOneProject, count)
						}

						return nil
					}),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.0.id", projectID),
					resource.TestCheckResourceAttr("data.scaleway_account_projects.projects", "projects.0.name", "default"),
				),
			},
		},
	})
}
