package account_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	accounttestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account/testfuncs"
)

func TestAccProject_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_account_project" "main" {
							name = "tf_tests_project_basic"
							description = "a description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					isProjectPresent(tt, "scaleway_account_project.main"),
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
					isProjectPresent(tt, "scaleway_account_project.main"),
					resource.TestCheckResourceAttr("scaleway_account_project.main", "name", "tf_tests_project_basic_rename"),
					resource.TestCheckResourceAttr("scaleway_account_project.main", "description", "another description"),
				),
			},
		},
	})
}

func TestAccProject_NoUpdate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             accounttestfuncs.IsProjectDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_account_project" "main" {
							name = "tf_tests_project_noupdate"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					isProjectPresent(tt, "scaleway_account_project.main"),
				),
			},
			{
				Config: `
						resource "scaleway_account_project" "main" {
							name = "tf_tests_project_noupdate"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					isProjectPresent(tt, "scaleway_account_project.main"),
				),
			},
		},
	})
}

func isProjectPresent(tt *acctest.TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		accountAPI := account.NewProjectAPI(tt.Meta)

		_, err := accountAPI.GetProject(&accountSDK.ProjectAPIGetProjectRequest{
			ProjectID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("could not find project: %w", err)
		}

		return nil
	}
}
