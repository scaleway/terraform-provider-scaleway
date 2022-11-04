package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	if !terraformBetaEnabled {
		return
	}
	resource.AddTestSweepers("scaleway_iam_api_key", &resource.Sweeper{
		Name: "scaleway_iam_api_key",
		F:    testSweepIamAPIKey,
	})
}

func testSweepIamAPIKey(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		api := iam.NewAPI(scwClient)

		l.Debugf("sweeper: destroying the api keys")

		listAPIKeys, err := api.ListAPIKeys(&iam.ListAPIKeysRequest{})
		if err != nil {
			return fmt.Errorf("failed to list api keys: %w", err)
		}
		for _, app := range listAPIKeys.APIKeys {
			err = api.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
				AccessKey: app.AccessKey,
			})
			if err != nil {
				return fmt.Errorf("failed to delete api key: %w", err)
			}
		}
		return nil
	})
}

func TestAccScalewayIamApiKey_WithApplication(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIamAPIKeyDestroy(tt),
			testAccCheckScalewayIamApplicationDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_basic"
						}

						resource "scaleway_iam_api_key" "main" {
							application_id = scaleway_iam_application.main.id
							description = "a description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamAPIKeyExists(tt, "scaleway_iam_api_key.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_api_key.main", "application_id", "scaleway_iam_application.main", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_api_key.main", "description", "a description"),
					resource.TestCheckResourceAttrSet("scaleway_iam_api_key.main", "secret_key"),
				),
			},
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_basic"
						}

						resource "scaleway_iam_api_key" "main" {
							application_id = scaleway_iam_application.main.id
							description = "another description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamAPIKeyExists(tt, "scaleway_iam_api_key.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_api_key.main", "application_id", "scaleway_iam_application.main", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_api_key.main", "description", "another description"),
					resource.TestCheckResourceAttrSet("scaleway_iam_api_key.main", "secret_key"),
				),
			},
			{
				ResourceName:            "scaleway_iam_api_key.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_key"},
			},
		},
	})
}

func TestAccScalewayIamApiKey_WithUser(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamAPIKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						data "scaleway_iam_user" "main" {
							email = "developer-tools-team@scaleway.com"
							organization_id = "dd5b8103-52ef-40b6-b157-35a426650401"
						}

						resource "scaleway_iam_api_key" "main" {
							user_id = data.scaleway_iam_user.main.id
							description = "a description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamAPIKeyExists(tt, "scaleway_iam_api_key.main"),
					testAccCheckScalewayIamUserExists(tt, "data.scaleway_iam_user.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_api_key.main", "user_id", "data.scaleway_iam_user.main", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_api_key.main", "description", "a description"),
				),
			},
			{
				Config: `
						data "scaleway_iam_user" "main" {
							email = "developer-tools-team@scaleway.com"
							organization_id = "dd5b8103-52ef-40b6-b157-35a426650401"
						}

						resource "scaleway_iam_api_key" "main" {
							user_id = data.scaleway_iam_user.main.id
							description = "another description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamAPIKeyExists(tt, "scaleway_iam_api_key.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_api_key.main", "user_id", "data.scaleway_iam_user.main", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_api_key.main", "description", "another description"),
				),
			},
		},
	})
}

func TestAccScalewayIamApiKey_Expires(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIamAPIKeyDestroy(tt),
			testAccCheckScalewayIamApplicationDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_expires_at"
						}

						resource "scaleway_iam_api_key" "main" {
							application_id = scaleway_iam_application.main.id
							description = "an api key with an expiration date"
							expires_at = "2025-07-06T09:00:00Z"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamAPIKeyExists(tt, "scaleway_iam_api_key.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_api_key.main", "application_id", "scaleway_iam_application.main", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_api_key.main", "description", "an api key with an expiration date"),
					resource.TestCheckResourceAttr("scaleway_iam_api_key.main", "expires_at", "2025-07-06T09:00:00Z"),
				),
			},
		},
	})
}

func TestAccScalewayIamApiKey_NoUpdate(t *testing.T) {
	SkipBetaTest(t)
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamAPIKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_noupdate"
						}

						resource "scaleway_iam_api_key" "main" {
							application_id = scaleway_iam_application.main.id
							description = "no update"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamAPIKeyExists(tt, "scaleway_iam_api_key.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_api_key.main", "application_id", "scaleway_iam_application.main", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_api_key.main", "description", "no update"),
				),
			},
			{
				Config: `
						resource "scaleway_iam_application" "main" {
							name = "tf_tests_app_noupdate"
						}

						resource "scaleway_iam_api_key" "main" {
							application_id = scaleway_iam_application.main.id
							description = "no update"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamAPIKeyExists(tt, "scaleway_iam_api_key.main"),
					resource.TestCheckResourceAttrPair("scaleway_iam_api_key.main", "application_id", "scaleway_iam_application.main", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_api_key.main", "description", "no update"),
				),
			},
		},
	})
}

func testAccCheckScalewayIamAPIKeyExists(tt *TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		iamAPI := iamAPI(tt.Meta)

		_, err := iamAPI.GetAPIKey(&iam.GetAPIKeyRequest{
			AccessKey: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("could not find api key: %w", err)
		}

		return nil
	}
}

func testAccCheckScalewayIamAPIKeyDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_iam_api_key" {
				continue
			}

			iamAPI := iamAPI(tt.Meta)

			_, err := iamAPI.GetAPIKey(&iam.GetAPIKeyRequest{
				AccessKey: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("resource %s(%s) still exist", rs.Type, rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return fmt.Errorf("error which is not an expected 404: %w", err)
			}
		}

		return nil
	}
}
