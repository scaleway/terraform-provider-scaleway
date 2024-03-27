package scaleway_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	accountV3 "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func init() {
	resource.AddTestSweepers("scaleway_cockpit_grafana_user", &resource.Sweeper{
		Name: "scaleway_cockpit_grafana_user",
		F:    testSweepCockpitGrafanaUser,
	})
}

func testSweepCockpitGrafanaUser(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		accountAPI := accountV3.NewProjectAPI(scwClient)
		cockpitAPI := cockpit.NewAPI(scwClient)

		listProjects, err := accountAPI.ListProjects(&accountV3.ProjectAPIListProjectsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		for _, project := range listProjects.Projects {
			if !strings.HasPrefix(project.Name, "tf_tests") {
				continue
			}

			listGrafanaUsers, err := cockpitAPI.ListGrafanaUsers(&cockpit.ListGrafanaUsersRequest{
				ProjectID: project.ID,
			}, scw.WithAllPages())
			if err != nil {
				if httperrors.Is404(err) {
					return nil
				}

				return fmt.Errorf("failed to list grafana users: %w", err)
			}

			for _, grafanaUser := range listGrafanaUsers.GrafanaUsers {
				err = cockpitAPI.DeleteGrafanaUser(&cockpit.DeleteGrafanaUserRequest{
					ProjectID:     project.ID,
					GrafanaUserID: grafanaUser.ID,
				})
				if err != nil {
					if !httperrors.Is404(err) {
						return fmt.Errorf("failed to delete grafana user: %w", err)
					}
				}
			}
		}

		return nil
	})
}

func TestAccScalewayCockpitGrafanaUser_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_user_basic"
	grafanaTestUsername := "testuserbasic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayCockpitGrafanaUserDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%[1]s"
					}

					resource scaleway_cockpit main {
						project_id = scaleway_account_project.project.id
					}

					resource scaleway_cockpit_grafana_user main {
						project_id = scaleway_cockpit.main.project_id
						login = "%[2]s"
						role = "editor"
					}
				`, projectName, grafanaTestUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitGrafanaUserExists(tt, "scaleway_cockpit_grafana_user.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_grafana_user.main", "project_id", "scaleway_cockpit.main", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_grafana_user.main", "login", grafanaTestUsername),
					resource.TestCheckResourceAttr("scaleway_cockpit_grafana_user.main", "role", "editor"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_grafana_user.main", "password"),
				),
			},
		},
	})
}

func TestAccScalewayCockpitGrafanaUser_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_user_update"
	grafanaTestUsername := "testuserupdate"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayCockpitGrafanaUserDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%[1]s"
				  	}

					resource scaleway_cockpit main {
						project_id = scaleway_account_project.project.id
					}

					resource scaleway_cockpit_grafana_user main {
						project_id = scaleway_cockpit.main.project_id
						login = "%[2]s"
						role = "editor"
					}
				`, projectName, grafanaTestUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitGrafanaUserExists(tt, "scaleway_cockpit_grafana_user.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_grafana_user.main", "project_id", "scaleway_cockpit.main", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_grafana_user.main", "login", grafanaTestUsername),
					resource.TestCheckResourceAttr("scaleway_cockpit_grafana_user.main", "role", "editor"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_grafana_user.main", "password"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%[1]s"
					}

					resource scaleway_cockpit main {
						project_id = scaleway_account_project.project.id
					}

					resource scaleway_cockpit_grafana_user main {
						project_id = scaleway_cockpit.main.project_id
						login = "%[2]s"
						role = "viewer"
					}
				`, projectName, grafanaTestUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitGrafanaUserExists(tt, "scaleway_cockpit_grafana_user.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_grafana_user.main", "project_id", "scaleway_cockpit.main", "project_id"),
					resource.TestCheckResourceAttr("scaleway_cockpit_grafana_user.main", "login", grafanaTestUsername),
					resource.TestCheckResourceAttr("scaleway_cockpit_grafana_user.main", "role", "viewer"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_grafana_user.main", "password"),
				),
			},
		},
	})
}

func TestAccScalewayCockpitGrafanaUser_NonExistentCockpit(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_user_non_existent_cockpit"
	grafanaTestUsername := "testnonexistentuser"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayCockpitGrafanaUserDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%[1]s"
					}

					resource scaleway_cockpit_grafana_user main {
						project_id = scaleway_account_project.project.id
						login = "%[2]s"
						role = "editor"
					}
				`, projectName, grafanaTestUsername),
				ExpectError: regexp.MustCompile("not found"),
			},
		},
	})
}

func testAccCheckScalewayCockpitGrafanaUserExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource cockpit grafana user not found: %s", n)
		}

		api, projectID, grafanaUserID, err := scaleway.CockpitAPIGrafanaUserID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		res, err := api.ListGrafanaUsers(&cockpit.ListGrafanaUsersRequest{
			ProjectID: projectID,
		}, scw.WithAllPages())
		if err != nil {
			return err
		}

		var grafanaUser *cockpit.GrafanaUser
		for _, user := range res.GrafanaUsers {
			if user.ID == grafanaUserID {
				grafanaUser = user
				break
			}
		}

		if grafanaUser == nil {
			return fmt.Errorf("cockpit grafana user (%d) (project %s) not found", grafanaUserID, projectID)
		}

		return nil
	}
}

func testAccCheckScalewayCockpitGrafanaUserDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit_grafana_user" {
				continue
			}

			api, projectID, grafanaUserID, err := scaleway.CockpitAPIGrafanaUserID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			err = api.DeleteGrafanaUser(&cockpit.DeleteGrafanaUserRequest{
				ProjectID:     projectID,
				GrafanaUserID: grafanaUserID,
			})
			if err == nil {
				return fmt.Errorf("cockpit grafana user (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) && !httperrors.Is403(err) {
				return err
			}
		}

		return nil
	}
}
