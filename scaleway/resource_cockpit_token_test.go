package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
)

func TestAccScalewayCockpitToken_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_token_basic"
	tokenName := "tf_tests_cockpit_token_basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayCockpitTokenDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%[1]s"
				  	}

					resource scaleway_cockpit main {
						project_id = scaleway_account_project.project.id
					}

					resource scaleway_cockpit_token main {
						project_id = scaleway_cockpit.main.project_id
						name = "%[2]s"
						scopes {
							query_metrics = true
							write_logs = false
						}
					}
				`, projectName, tokenName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitTokenExists(tt, "scaleway_cockpit_token.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_token.main", "project_id", "scaleway_cockpit.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_token.main", "secret_key"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "name", tokenName),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.query_metrics", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.write_metrics", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.setup_metrics_rules", "false"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.query_logs", "false"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.write_logs", "false"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.setup_logs_rules", "false"),
				),
			},
		},
	})
}

func TestAccScalewayCockpitToken_Update(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	projectName := "tf_tests_cockpit_token_update"
	tokenName := "tf_tests_cockpit_token_update"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayCockpitTokenDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "project" {
						name = "%[1]s"
				  	}

					resource scaleway_cockpit main {
						project_id = scaleway_account_project.project.id
					}

					resource scaleway_cockpit_token main {
						project_id = scaleway_cockpit.main.project_id
						name = "%[2]s"
						scopes {
							query_metrics = true
							write_logs = false
						}
					}
				`, projectName, tokenName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitTokenExists(tt, "scaleway_cockpit_token.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_token.main", "project_id", "scaleway_cockpit.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_token.main", "secret_key"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "name", tokenName),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.query_metrics", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.write_metrics", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.setup_metrics_rules", "false"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.query_logs", "false"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.write_logs", "false"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.setup_logs_rules", "false"),
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

					resource scaleway_cockpit_token main {
						project_id = scaleway_cockpit.main.project_id
						name = "%[2]s"
						scopes {
							query_metrics = true
							write_logs = true
						}
					}
				`, projectName, tokenName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayCockpitTokenExists(tt, "scaleway_cockpit_token.main"),
					resource.TestCheckResourceAttrPair("scaleway_cockpit_token.main", "project_id", "scaleway_cockpit.main", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_cockpit_token.main", "secret_key"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "name", tokenName),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.query_metrics", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.write_metrics", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.setup_metrics_rules", "false"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.query_logs", "false"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.write_logs", "true"),
					resource.TestCheckResourceAttr("scaleway_cockpit_token.main", "scopes.0.setup_logs_rules", "false"),
				),
			},
		},
	})
}

func testAccCheckScalewayCockpitTokenExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource cockpit token not found: %s", n)
		}

		api, err := cockpitAPI(tt.Meta)
		if err != nil {
			return err
		}

		_, err = api.GetToken(&cockpit.GetTokenRequest{
			TokenID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayCockpitTokenDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_cockpit_token" {
				continue
			}

			api, err := cockpitAPI(tt.Meta)
			if err != nil {
				return err
			}

			err = api.DeleteToken(&cockpit.DeleteTokenRequest{
				TokenID: rs.Primary.ID,
			})
			if err == nil {
				return fmt.Errorf("cockpit token (%s) still exists", rs.Primary.ID)
			}

			// Currently the API returns a 403 error when we try to delete a token that does not exist
			if !is404Error(err) && !is403Error(err) {
				return err
			}
		}

		return nil
	}
}
