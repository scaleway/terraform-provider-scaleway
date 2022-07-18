package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	accountV2 "github.com/scaleway/scaleway-sdk-go/api/account/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	if !terraformBetaEnabled {
		return
	}
	resource.AddTestSweepers("scaleway_account_project", &resource.Sweeper{
		Name: "scaleway_account_project",
		F:    testSweepAccountproject,
	})
}

func testSweepAccountproject(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		accountAPI := accountV2.NewAPI(scwClient)

		l.Debugf("sweeper: destroying the project")

		listProjects, err := accountAPI.ListProjects(&accountV2.ListProjectsRequest{})
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}
		for _, project := range listProjects.Projects {
			err = accountAPI.DeleteProject(&accountV2.DeleteProjectRequest{
				ProjectID: project.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete project: %w", err)
			}
		}
		return nil
	})
}

func TestAccScalewayAccountProject_Basic(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayAccountProjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_account_project" "main" {
							name = "tf_tests_project_basic"
							description = "a description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountProjectExists(tt, "scaleway_account_project.main"),
					resource.TestCheckResourceAttr("scaleway_account_project.main", "name", "tf_tests_project_basic"),
					resource.TestCheckResourceAttr("scaleway_account_project.main", "description", "a description"),
				),
			},
			{
				Config: `
						resource "scaleway_account_project" "main" {
							name = "tf_tests_project_basic_rename"
							description = "another description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountProjectExists(tt, "scaleway_account_project.main"),
					resource.TestCheckResourceAttr("scaleway_account_project.main", "name", "tf_tests_project_basic_rename"),
					resource.TestCheckResourceAttr("scaleway_account_project.main", "description", "another description"),
				),
			},
		},
	})
}

func TestAccScalewayAccountProject_NoUpdate(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayAccountProjectDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_account_project" "main" {
							name = "tf_tests_project_noupdate"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountProjectExists(tt, "scaleway_account_project.main"),
				),
			},
			{
				Config: `
						resource "scaleway_account_project" "main" {
							name = "tf_tests_project_noupdate"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountProjectExists(tt, "scaleway_account_project.main"),
				),
			},
		},
	})
}

func testAccCheckScalewayAccountProjectExists(tt *TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		accountAPI := accountV2API(tt.Meta)

		_, err := accountAPI.GetProject(&accountV2.GetProjectRequest{
			ProjectID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("could not find project: %w", err)
		}

		return nil
	}
}

func testAccCheckScalewayAccountProjectDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_account_project" {
				continue
			}

			accountAPI := accountV2API(tt.Meta)

			_, err := accountAPI.GetProject(&accountV2.GetProjectRequest{
				ProjectID: rs.Primary.ID,
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
