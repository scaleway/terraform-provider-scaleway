package applesilicon_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	applesiliconSDK "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/applesilicon"
)

func TestAccRunner_BasicGithub(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	t.Skip("can not register this cassette for security issue")
	var githubUrl = os.Getenv("GITHUB_URL_AS")
	var githubToken = os.Getenv("GITHUB_TOKEN_AS")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isRunnerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_apple_silicon_runner" "main" {
						name       = "TestAccRunnerGithub"
						ci_provider   = "github"
						url        = "%s"
						token      = "%s"
					}
				`, githubUrl, githubToken),
				Check: resource.ComposeTestCheckFunc(
					isRunnerPresent(tt, "scaleway_apple_silicon_runner.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_runner.main", "name", "TestAccRunnerGithub"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_runner.main", "ci_provider", "github"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_runner.main", "url", githubUrl),

					// Computed
					resource.TestCheckResourceAttrSet("scaleway_apple_silicon_runner.main", "status"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_apple_silicon_runner" "main" {
						name       = "TestAccRunnerGithubUpdated"
						ci_provider   = "github"
						url        = "%s"
						token      = "%s"
					}
				`, githubUrl, githubToken),
				Check: resource.ComposeTestCheckFunc(
					isRunnerPresent(tt, "scaleway_apple_silicon_runner.main"),
					resource.TestCheckResourceAttr("scaleway_apple_silicon_runner.main", "name", "TestAccRunnerGithubUpdated"),
				),
			},
		},
	})
}

func isRunnerPresent(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		api, zone, id, err := applesilicon.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetRunner(&applesiliconSDK.GetRunnerRequest{
			Zone:     zone,
			RunnerID: id,
		})

		return err
	}
}

func isRunnerDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_apple_silicon_runner" {
				continue
			}

			api, zone, id, err := applesilicon.NewAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetRunner(&applesiliconSDK.GetRunnerRequest{
				Zone:     zone,
				RunnerID: id,
			})

			if err == nil {
				return fmt.Errorf("runner still exists: %s", rs.Primary.ID)
			}

			if !httperrors.Is403(err) {
				return fmt.Errorf("unexpected error: %s", err)
			}
		}

		return nil
	}
}
