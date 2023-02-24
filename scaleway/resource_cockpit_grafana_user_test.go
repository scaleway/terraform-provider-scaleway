package scaleway

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	grafanaTestUsername = "testuser"
)

func TestAccScalewayCockpitGrafanaUser_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_user_basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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
	tt := NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_user_update"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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
	tt := NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_grafana_user_non_existent_cockpit"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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

func testAccCheckScalewayCockpitGrafanaUserExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource cockpit grafana user not found: %s", n)
		}

		api, projectID, grafanaUserID, err := cockpitAPIGrafanaUserID(tt.Meta, rs.Primary.ID)
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

func testAccCheckScalewayCockpitGrafanaUserDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit_grafana_user" {
				continue
			}

			api, projectID, grafanaUserID, err := cockpitAPIGrafanaUserID(tt.Meta, rs.Primary.ID)
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

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
