package iam_test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
	secrettestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret/testfuncs"
)

func TestAccApiKeyEphemeralResource_WithApplication(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithApplication because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		// This hardcoded value has to be replaced with the expiration in cassettes.
		// Should be in the first "POST /api-keys" request.
		expiresAt = "2026-01-22T16:35:05Z"
	}

	description := "tf_test_api_key_er_with_app"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
			secrettestfuncs.CheckSecretDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_application" "main" {
						name = "%[1]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[1]s"
						expires_at = "%[2]s"
					}

					resource "scaleway_secret" "main" {
						name        = "%[1]s"
					}

					resource "scaleway_secret_version" "access_key" {
						description = "%[1]s"
						secret_id   = scaleway_secret.main.id
						data_wo     = ephemeral.scaleway_iam_api_key.main.access_key
					}

					data "scaleway_secret_version" "access_key" {
						secret_id = scaleway_secret.main.id
						revision  = "1"
						depends_on = [scaleway_secret_version.access_key]
					} 

					resource "scaleway_secret_version" "desc" {
						description = "%[1]s"
						secret_id   = scaleway_secret.main.id
						data_wo     = ephemeral.scaleway_iam_api_key.main.description
						depends_on 	= [scaleway_secret_version.access_key]
					}

					data "scaleway_secret_version" "desc" {
						secret_id = scaleway_secret.main.id
						revision  = "2"
						depends_on = [scaleway_secret_version.desc]
					} 

					resource "scaleway_secret_version" "app_id" {
						description = "%[1]s"
						secret_id   = scaleway_secret.main.id
						data_wo     = ephemeral.scaleway_iam_api_key.main.application_id
						depends_on 	= [scaleway_secret_version.desc]
					}

					data "scaleway_secret_version" "app_id" {
						secret_id = scaleway_secret.main.id
						revision  = "3"
						depends_on = [scaleway_secret_version.app_id]
					} 
					`, description, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEphemeralIamAPIKeyExists(tt, "data.scaleway_secret_version.access_key"),
					testAccCheckSecretVersionDataEquals("data.scaleway_secret_version.app_id", "data", "scaleway_iam_application.main", "id"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.desc", "data", secret.Base64Encoded([]byte(description))),
				),
			},
		},
	})
}

func TestAccApiKeyEphemeralResource_DefaultProject(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_DefaultProject because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	projectID, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		projectID = "105bdce1-64c0-48ab-899d-868455867ecf"
	}

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		// This hardcoded value has to be replaced with the expiration in cassettes.
		// Should be in the first "POST /api-keys" request.
		expiresAt = "2026-01-22T16:35:28Z"
	}

	description := "tf_test_api_key_er_project"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			iamchecks.CheckUserDestroyed(tt),
			testAccCheckIamAPIKeyDestroy(tt),
			secrettestfuncs.CheckSecretDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_application" "main" {
						name = "%[1]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[1]s"
						expires_at = "%[2]s"
						default_project_id = "%[3]s"
					}

					resource "scaleway_secret" "main" {
						name        = "%[1]s"
					}

					resource "scaleway_secret_version" "access_key" {
						description = "%[1]s"
						secret_id   = scaleway_secret.main.id
						data_wo     = ephemeral.scaleway_iam_api_key.main.access_key
					}

					data "scaleway_secret_version" "access_key" {
						secret_id = scaleway_secret.main.id
						revision  = "1"
						depends_on = [scaleway_secret_version.access_key]
					} 

					resource "scaleway_secret_version" "desc" {
						description = "%[1]s"
						secret_id   = scaleway_secret.main.id
						data_wo     = ephemeral.scaleway_iam_api_key.main.description
						depends_on 	= [scaleway_secret_version.access_key]
					}

					data "scaleway_secret_version" "desc" {
						secret_id = scaleway_secret.main.id
						revision  = "2"
						depends_on = [scaleway_secret_version.desc]
					} 

					resource "scaleway_secret_version" "project_id" {
						description = "%[1]s"
						secret_id   = scaleway_secret.main.id
						data_wo     = ephemeral.scaleway_iam_api_key.main.default_project_id
						depends_on 	= [scaleway_secret_version.desc]
					}

					data "scaleway_secret_version" "project_id" {
						secret_id = scaleway_secret.main.id
						revision  = "3"
						depends_on = [scaleway_secret_version.project_id]
					} 
					`, description, expiresAt, projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEphemeralIamAPIKeyExists(tt, "data.scaleway_secret_version.access_key"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.project_id", "data", secret.Base64Encoded([]byte(projectID))),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.desc", "data", secret.Base64Encoded([]byte(description))),
				),
			},
		},
	})
}

func testAccCheckSecretVersionDataEquals(secretVersionDataSource, secretVersionAttribute, expectedResource, expectedAttribute string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		secretRs, ok := state.RootModule().Resources[secretVersionDataSource]
		if !ok {
			return fmt.Errorf("secret version data source not found: %s", secretVersionDataSource)
		}

		expectedRs, ok := state.RootModule().Resources[expectedResource]
		if !ok {
			return fmt.Errorf("expected resource not found: %s", expectedResource)
		}

		encodedData := secretRs.Primary.Attributes[secretVersionAttribute]
		if encodedData == "" {
			return fmt.Errorf("secret version attribute %s is empty", secretVersionAttribute)
		}

		decodedBytes, err := base64.StdEncoding.DecodeString(encodedData)
		if err != nil {
			return fmt.Errorf("failed to decode base64 data from secret version: %w", err)
		}

		decodedData := string(decodedBytes)

		expectedValue := expectedRs.Primary.Attributes[expectedAttribute]
		if expectedValue == "" {
			return fmt.Errorf("expected attribute %s is empty", expectedAttribute)
		}

		if decodedData != expectedValue {
			return fmt.Errorf("secret version data (decoded: %s) does not match expected value (%s)", decodedData, expectedValue)
		}

		return nil
	}
}

func testAccCheckEphemeralIamAPIKeyExists(tt *acctest.TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		key, err := base64.StdEncoding.DecodeString(rs.Primary.Attributes["data"])
		if err != nil {
			return fmt.Errorf("could not find api key: %w", err)
		}

		iamAPI := iam.NewAPI(tt.Meta)

		_, err = iamAPI.GetAPIKey(&iamSDK.GetAPIKeyRequest{
			AccessKey: string(key),
		})
		if err != nil {
			return fmt.Errorf("could not find api key: %w", err)
		}

		return nil
	}
}
