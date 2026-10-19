package iam_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
)

// accessKeyPattern matches the format of a Scaleway IAM access key.
var accessKeyPattern = regexp.MustCompile(`^SCW[0-9A-Z]{17}$`)

func TestAccApiKeyEphemeralResource_WithApplication(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithApplication because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	tt.ProviderFactories["echo"] = echoprovider.NewProviderServer()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)

	description := "tf_test_api_key_er_with_app"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckEphemeralIamAPIKeyDeleted(tt, "echo.test_api_key"),
		),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: fmt.Sprintf(`
					resource "scaleway_iam_application" "main" {
						name = "%[1]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[1]s"
						expires_at = "%[2]s"
						delete_on_close = false
					}

					provider "echo" {
						data = ephemeral.scaleway_iam_api_key.main
					}

					resource "echo" "test_api_key" {}
				`, description, expiresAt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("access_key"), knownvalue.StringRegexp(accessKeyPattern)),
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("description"), knownvalue.StringExact(description)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEphemeralIamAPIKeyExists(tt, "echo.test_api_key"),
					testAccCheckEchoAttributeMatches("echo.test_api_key", "application_id", "scaleway_iam_application.main", "id"),
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

	tt.ProviderFactories["echo"] = echoprovider.NewProviderServer()

	projectID, projectIDExists := tt.Meta.ScwClient().GetDefaultProjectID()
	if !projectIDExists {
		t.Skip("no default project ID")
	}

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)

	description := "tf_test_api_key_er_project"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckEphemeralIamAPIKeyDeleted(tt, "echo.test_api_key"),
		),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: fmt.Sprintf(`
					resource "scaleway_iam_application" "main" {
						name = "%[1]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[1]s"
						expires_at = "%[2]s"
						default_project_id = "%[3]s"
						delete_on_close = false
					}

					provider "echo" {
						data = ephemeral.scaleway_iam_api_key.main
					}

					resource "echo" "test_api_key" {}
				`, description, expiresAt, projectID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("access_key"), knownvalue.StringRegexp(accessKeyPattern)),
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("default_project_id"), knownvalue.StringExact(projectID)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEphemeralIamAPIKeyExists(tt, "echo.test_api_key"),
				),
			},
		},
	})
}

func TestAccApiKeyEphemeralResource_DeleteOnCloseDefault(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_DeleteOnCloseDefault because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	tt.ProviderFactories["echo"] = echoprovider.NewProviderServer()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)

	description := "tf_test_api_key_er_delete_default"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckEphemeralIamAPIKeyDeleted(tt, "echo.test_api_key"),
		),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: fmt.Sprintf(`
					resource "scaleway_iam_application" "main" {
						name = "%[1]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[1]s"
						expires_at = "%[2]s"
					}

					provider "echo" {
						data = ephemeral.scaleway_iam_api_key.main
					}

					resource "echo" "test_api_key" {}
				`, description, expiresAt),
				ConfigStateChecks: []statecheck.StateCheck{
					// Default delete_on_close = true
					// we rely on CheckDestroy to verify it was deleted.
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("access_key"), knownvalue.StringRegexp(accessKeyPattern)),
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("description"), knownvalue.StringExact(description)),
				},
			},
		},
	})
}

func TestAccApiKeyEphemeralResource_DeleteOnCloseFalse(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_DeleteOnCloseFalse because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	tt.ProviderFactories["echo"] = echoprovider.NewProviderServer()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)

	description := "tf_test_api_key_er_delete_false"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: fmt.Sprintf(`
					resource "scaleway_iam_application" "main" {
						name = "%[1]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[1]s"
						expires_at = "%[2]s"
						delete_on_close = false
					}

					provider "echo" {
						data = ephemeral.scaleway_iam_api_key.main
					}

					resource "echo" "test_api_key" {}
				`, description, expiresAt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("access_key"), knownvalue.StringRegexp(accessKeyPattern)),
					statecheck.ExpectKnownValue("echo.test_api_key", dataPath.AtMapKey("description"), knownvalue.StringExact(description)),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEphemeralIamAPIKeyExists(tt, "echo.test_api_key"),
				),
			},
		},
	})
}

// testAccCheckEphemeralIamAPIKeyDeleted asserts that the ephemeral API key
// echoed by the given echo resource instance has been deleted.
func testAccCheckEphemeralIamAPIKeyDeleted(tt *acctest.TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		accessKey, err := acctest.GetResourceAttributeValue(s, name, "data.access_key")
		if err != nil {
			return err
		}

		iamAPI := iam.NewAPI(tt.Meta)
		ctx := tt.T.Context()

		return retry.RetryContext(ctx, iamchecks.DestroyWaitTimeout, func() *retry.RetryError {
			_, err := iamAPI.GetAPIKey(&iamSDK.GetAPIKeyRequest{
				AccessKey: accessKey,
			})

			switch {
			case err == nil:
				return retry.RetryableError(fmt.Errorf("IAM API key (%s) still exists", accessKey))
			case httperrors.Is404(err):
				return nil
			default:
				return retry.NonRetryableError(err)
			}
		})
	}
}

// testAccCheckEchoAttributeMatches asserts that the attribute echoed by the
// given echo resource instance matches the value of an attribute of another
// resource in state.
func testAccCheckEchoAttributeMatches(echoResource, echoAttribute, expectedResource, expectedAttribute string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		echoValue, err := acctest.GetResourceAttributeValue(state, echoResource, "data."+echoAttribute)
		if err != nil {
			return err
		}

		expectedValue, err := acctest.GetResourceAttributeValue(state, expectedResource, expectedAttribute)
		if err != nil {
			return err
		}

		if echoValue != expectedValue {
			return fmt.Errorf("echo data.%s (%s) does not match %s.%s (%s)", echoAttribute, echoValue, expectedResource, expectedAttribute, expectedValue)
		}

		return nil
	}
}

func testAccCheckEphemeralIamAPIKeyExists(tt *acctest.TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		key, err := acctest.GetResourceAttributeValue(s, name, "data.access_key")
		if err != nil {
			return err
		}

		iamAPI := iam.NewAPI(tt.Meta)

		_, err = iamAPI.GetAPIKey(&iamSDK.GetAPIKeyRequest{
			AccessKey: key,
		})
		if err != nil {
			return fmt.Errorf("could not find api key: %w", err)
		}

		return nil
	}
}
