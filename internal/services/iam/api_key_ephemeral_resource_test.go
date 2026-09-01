package iam_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	iamSDK "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
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
		expiresAt = "2026-06-10T16:22:39Z"
	}

	description := "tf_test_api_key_er_with_app"
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "%[2]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[2]s"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), description, expiresAt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("application_id"), knownvalue.NotNull()),
				},
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
		expiresAt = "2026-06-10T16:22:39Z"
	}

	description := "tf_test_api_key_er_project"
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			iamchecks.CheckUserDestroyed(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "%[2]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[2]s"
						expires_at = "%[3]s"
						default_project_id = "%[4]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), description, expiresAt, projectID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("default_project_id"), knownvalue.StringExact(projectID)),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Create tests that when no previous
// resource exists and ephemeral_lifecycle is "persist", a new resource is created, identifying the
// new API key via its description.
func TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Create(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Create because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	descriptionIdentifier := "tf_test_desc_identifier_create_unique"
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "%[2]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description_identifier = "%[2]s"
						ephemeral_lifecycle = "persist"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), descriptionIdentifier, expiresAt),
				Check: testAccCheckIamAPIKeyExistsViaDescription(tt, descriptionIdentifier),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("description_identifier"), knownvalue.StringExact(descriptionIdentifier)),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Replace tests that when no previous
// resource exists and ephemeral_lifecycle is "replace", a new resource is created, identifying the
// new API key via its description.
func TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Replace(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Replace because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	descriptionIdentifier := "tf_test_desc_identifier_replace_unique"
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "%[2]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description_identifier = "%[2]s"
						ephemeral_lifecycle = "replace"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), descriptionIdentifier, expiresAt),
				Check: testAccCheckIamAPIKeyExistsViaDescription(tt, descriptionIdentifier),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("description_identifier"), knownvalue.StringExact(descriptionIdentifier)),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("ephemeral_lifecycle"), knownvalue.StringExact("replace")),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_NoRecreate tests that when a resource already
// exists and ephemeral_lifecycle is "persist", the existing resource is reused (not recreated), identifying the
// existing API key via its description.
func TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_NoRecreate(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_NoRecreate because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	descriptionIdentifier := "tf_test_desc_identifier_reuse_unique"
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	var firstAccessKey string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "%[2]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description_identifier = "%[2]s"
						ephemeral_lifecycle = "persist"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), descriptionIdentifier, expiresAt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("description_identifier"), knownvalue.StringExact(descriptionIdentifier)),
				},
			},
			{
				PreConfig: func() {
					var err error

					firstAccessKey, err = getAPIKeyAccessKeyByDescription(tt, descriptionIdentifier)
					if err != nil {
						tt.T.Fatalf("expected exactly one API key before second step: %v", err)
					}
				},
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "%[2]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description_identifier = "%[2]s"
						ephemeral_lifecycle = "persist"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), descriptionIdentifier, expiresAt),
				Check: func(s *terraform.State) error {
					currentAccessKey, err := getAPIKeyAccessKeyByDescription(tt, descriptionIdentifier)
					if err != nil {
						return err
					}

					if currentAccessKey != firstAccessKey {
						return fmt.Errorf("expected access key to remain the same, but it changed: was %q, now %q", firstAccessKey, currentAccessKey)
					}

					return nil
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("description_identifier"), knownvalue.StringExact(descriptionIdentifier)),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Recreate tests that when a resource already
// exists and ephemeral_lifecycle is "replace", a new resource is created and the old one is deleted, identifying
// the API key via its description.
func TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Recreate(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_Recreate because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-08-25T14:29:31Z"
	}

	descriptionIdentifier := "tf_test_desc_identifier_recreate_unique"
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	var firstAccessKey string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "%[2]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description_identifier = "%[2]s"
						ephemeral_lifecycle = "replace"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), descriptionIdentifier, expiresAt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("description_identifier"), knownvalue.StringExact(descriptionIdentifier)),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("ephemeral_lifecycle"), knownvalue.StringExact("replace")),
				},
			},
			{
				PreConfig: func() {
					var err error

					firstAccessKey, err = getAPIKeyAccessKeyByDescription(tt, descriptionIdentifier)
					if err != nil {
						tt.T.Fatalf("expected exactly one API key before second step: %v", err)
					}
				},
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "%[2]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description_identifier = "%[2]s"
						ephemeral_lifecycle = "replace"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), descriptionIdentifier, expiresAt),
				Check: func(s *terraform.State) error {
					return resource.ComposeTestCheckFunc(
						testAccCheckIamAPIKeyAccessKeyDeleted(tt, firstAccessKey),
						func(s *terraform.State) error {
							currentAccessKey, err := getAPIKeyAccessKeyByDescription(tt, descriptionIdentifier)
							if err != nil {
								return err
							}

							if currentAccessKey == firstAccessKey {
								return fmt.Errorf("expected a new API key to be created, but access key is the same: %s", currentAccessKey)
							}

							return nil
						},
					)(s)
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("description_identifier"), knownvalue.StringExact(descriptionIdentifier)),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Create tests that when no previous
// resource exists and ephemeral_lifecycle is "persist", a new resource is created, identifying the
// new API key via its annotations.
func TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Create(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Create because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
			testAccCheckIamAPIKeyAnnotationBindingDestroy(tt, "my_identifier_value"),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "tf_test_api_key_annotations_create"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						annotation_identifier = "my_identifier_value"
						ephemeral_lifecycle = "persist"
						expires_at = "%[2]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), expiresAt),
				Check: testAccCheckIamAPIKeyExistsViaAnnotations(tt, "my_identifier_value"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("annotation_identifier"), knownvalue.StringExact("my_identifier_value")),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Replace tests that when no previous
// resource exists and ephemeral_lifecycle is "replace", a new resource is created, identifying the
// new API key via its annotations.
func TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Replace(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Replace because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "tf_test_api_key_annotations_replace"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						annotation_identifier = "my_identifier_value"
						ephemeral_lifecycle = "replace"
						expires_at = "%[2]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), expiresAt),
				Check: testAccCheckIamAPIKeyExistsViaAnnotations(tt, "my_identifier_value"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("annotation_identifier"), knownvalue.StringExact("my_identifier_value")),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("ephemeral_lifecycle"), knownvalue.StringExact("replace")),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_NoRecreate tests that when a resource already
// exists and ephemeral_lifecycle is "persist", the existing resource is reused (not recreated), identifying the
// existing API key via its annotations.
func TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_NoRecreate(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_NoRecreate because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")
	annotationValue := "my_identifier_value"

	var firstAccessKey string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "tf_test_api_key_annotations_reuse"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						annotation_identifier = "%[2]s"
						ephemeral_lifecycle = "persist"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), annotationValue, expiresAt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("annotation_identifier"), knownvalue.StringExact(annotationValue)),
				},
			},
			{
				PreConfig: func() {
					var err error

					firstAccessKey, err = getAPIKeyAccessKeyByAnnotation(tt, annotationValue)
					if err != nil {
						tt.T.Fatalf("expected exactly one API key before second step: %v", err)
					}
				},
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "tf_test_api_key_annotations_reuse"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						annotation_identifier = "%[2]s"
						ephemeral_lifecycle = "persist"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), annotationValue, expiresAt),
				Check: func(s *terraform.State) error {
					currentAccessKey, err := getAPIKeyAccessKeyByAnnotation(tt, annotationValue)
					if err != nil {
						return err
					}

					if currentAccessKey != firstAccessKey {
						return fmt.Errorf("expected access key to remain the same, but it changed: was %q, now %q", firstAccessKey, currentAccessKey)
					}

					return nil
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("annotation_identifier"), knownvalue.StringExact(annotationValue)),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Recreate tests that when a resource already
// exists and ephemeral_lifecycle is "replace", a new resource is created and the old one is deleted, identifying
// the API key via its annotations.
func TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Recreate(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_Recreate because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")
	annotationValue := "my_identifier_value"

	var firstAccessKey string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.MergeProviderFactories(tt.ProviderFactories, acctest.ProtoV6ProviderFactoriesEcho()),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "tf_test_api_key_annotations_replace"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						annotation_identifier = "%[2]s"
						ephemeral_lifecycle = "replace"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), annotationValue, expiresAt),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("annotation_identifier"), knownvalue.StringExact(annotationValue)),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("ephemeral_lifecycle"), knownvalue.StringExact("replace")),
				},
			},
			{
				PreConfig: func() {
					var err error

					firstAccessKey, err = getAPIKeyAccessKeyByAnnotation(tt, annotationValue)
					if err != nil {
						tt.T.Fatalf("expected exactly one API key before second step: %v", err)
					}
				},
				Config: fmt.Sprintf(`
					%s

					resource "scaleway_iam_application" "main" {
						name = "tf_test_api_key_annotations_replace"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						annotation_identifier = "%[2]s"
						ephemeral_lifecycle = "replace"
						expires_at = "%[3]s"
					}
					`, acctest.ConfigWithEchoProvider("ephemeral.scaleway_iam_api_key.main"), annotationValue, expiresAt),
				Check: func(s *terraform.State) error {
					return resource.ComposeTestCheckFunc(
						testAccCheckIamAPIKeyAccessKeyDeleted(tt, firstAccessKey),
						func(s *terraform.State) error {
							currentAccessKey, err := getAPIKeyAccessKeyByAnnotation(tt, annotationValue)
							if err != nil {
								return err
							}

							if currentAccessKey == firstAccessKey {
								return fmt.Errorf("expected a new API key to be created, but access key is the same: %s", currentAccessKey)
							}

							return nil
						},
					)(s)
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("annotation_identifier"), knownvalue.StringExact(annotationValue)),
				},
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_ErrorMismatch tests that an error occurs
// when the description attribute is specified together with description_identifier.
func TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_ErrorMismatch(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithDescriptionIdentifier_ErrorMismatch because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	descriptionIdentifier := "tf_test_desc_identifier_mismatch_unique"
	description := "tf_test_api_key_desc_mismatch"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_application" "main" {
						name = "%[1]s"
					}

					ephemeral "scaleway_iam_api_key" "main" {
						application_id = scaleway_iam_application.main.id
						description = "%[2]s"
						description_identifier = "%[3]s"
						ephemeral_lifecycle = "persist"
						expires_at = "%[4]s"
					}
					`, description, description, descriptionIdentifier, expiresAt),
				ExpectError: regexp.MustCompile(`Attribute "description" cannot be specified when "description_identifier"`),
			},
		},
	})
}

// TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_ErrorMismatch tests that an error occurs
// when the description attribute is specified together with annotation_identifier.
func TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_ErrorMismatch(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccApiKeyEphemeralResource_WithAnnotationsIdentifier_ErrorMismatch because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	expiresAt := time.Now().Add(time.Minute * 10).UTC().Format(time.RFC3339)
	if !*acctest.UpdateCassettes {
		expiresAt = "2026-06-10T16:22:39Z"
	}

	description := "tf_test_api_key_annotations_mismatch"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckIamApplicationDestroy(tt),
			testAccCheckIamAPIKeyDestroy(tt),
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
						annotation_identifier = "my_identifier_value"
						ephemeral_lifecycle = "persist"
						expires_at = "%[2]s"
					}
					`, description, expiresAt),
				ExpectError: regexp.MustCompile(`Attribute "description" cannot be specified when "annotation_identifier"`),
			},
		},
	})
}

// getAPIKeyAccessKeyByDescription returns the access key of the single API key with the given description.
// It fails if no API key or more than one API key is found.
func getAPIKeyAccessKeyByDescription(tt *acctest.TestTools, description string) (string, error) {
	iamAPI := iam.NewAPI(tt.Meta)

	orgID, exists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !exists {
		return "", errors.New("organization ID not found")
	}

	apiKeys, err := iamAPI.ListAPIKeys(&iamSDK.ListAPIKeysRequest{
		OrganizationID: &orgID,
		Description:    &description,
	}, scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("failed to list API keys: %w", err)
	}

	if len(apiKeys.APIKeys) == 0 {
		return "", fmt.Errorf("no API key found with description %q", description)
	}

	if len(apiKeys.APIKeys) > 1 {
		return "", fmt.Errorf("found %d API keys with description %q, expected exactly 1", len(apiKeys.APIKeys), description)
	}

	return apiKeys.APIKeys[0].AccessKey, nil
}

// testAccCheckIamAPIKeyExistsViaDescription returns a check function that verifies exactly one API key
// exists with the given description by calling the API.
func testAccCheckIamAPIKeyExistsViaDescription(tt *acctest.TestTools, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getAPIKeyAccessKeyByDescription(tt, description)

		return err
	}
}

// testAccCheckIamAPIKeyExistsViaAnnotations returns a check function that verifies exactly one API key
// exists with the given annotation value by calling the API.
func testAccCheckIamAPIKeyExistsViaAnnotations(tt *acctest.TestTools, annotationValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getAPIKeyAccessKeyByAnnotation(tt, annotationValue)

		return err
	}
}

// testAccCheckIamAPIKeyAccessKeyDeleted returns a check function that verifies an API key
// with the given access key has been deleted.
func testAccCheckIamAPIKeyAccessKeyDeleted(tt *acctest.TestTools, accessKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamAPI := iam.NewAPI(tt.Meta)

		_, err := iamAPI.GetAPIKey(&iamSDK.GetAPIKeyRequest{
			AccessKey: accessKey,
		})
		if err == nil {
			return fmt.Errorf("API key with access key %q still exists, expected it to be deleted", accessKey)
		}

		if !httperrors.Is403(err) && !httperrors.Is404(err) {
			return fmt.Errorf("failed to check API key deletion: %w", err)
		}

		return nil
	}
}

func getAPIKeyAccessKeyByAnnotation(tt *acctest.TestTools, annotationValue string) (string, error) {
	annotationsAPI := annotations.NewAPI(tt.Meta.ScwClient())

	orgID, exists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !exists {
		return "", errors.New("organization ID not found")
	}

	keysResp, err := annotationsAPI.ListKeys(&annotations.ListKeysRequest{
		OrganizationID: orgID,
	}, scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("failed to list annotation keys: %w", err)
	}

	var keyID string

	for _, key := range keysResp.Keys {
		if key.Name == "iam_terraform_identifier" {
			keyID = key.ID

			break
		}
	}

	if keyID == "" {
		return "", errors.New("annotation key 'iam_terraform_identifier' not found")
	}

	valuesResp, err := annotationsAPI.ListValues(&annotations.ListValuesRequest{
		KeyID: &keyID,
	}, scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("failed to list annotation values: %w", err)
	}

	var valueID string

	for _, value := range valuesResp.Values {
		if value.Name == annotationValue {
			valueID = value.ID

			break
		}
	}

	if valueID == "" {
		return "", fmt.Errorf("annotation value %q not found", annotationValue)
	}

	bindingsResp, err := annotationsAPI.ListBindings(&annotations.ListBindingsRequest{
		OrganizationID: orgID,
		ValueID:        &valueID,
	}, scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("failed to list bindings: %w", err)
	}

	if len(bindingsResp.Bindings) == 0 {
		return "", fmt.Errorf("no binding found for annotation value %q", annotationValue)
	}

	if len(bindingsResp.Bindings) > 1 {
		return "", fmt.Errorf("found %d bindings for annotation value %q, expected exactly 1", len(bindingsResp.Bindings), annotationValue)
	}

	binding := bindingsResp.Bindings[0]

	accessKey := binding.Srn
	if idx := strings.LastIndex(binding.Srn, "/"); idx != -1 {
		accessKey = binding.Srn[idx+1:]
	}

	return accessKey, nil
}

func testAccCheckIamAPIKeyAnnotationBindingDestroy(tt *acctest.TestTools, annotationValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		annotationsAPI := annotations.NewAPI(tt.Meta.ScwClient())

		orgID, exists := tt.Meta.ScwClient().GetDefaultOrganizationID()
		if !exists {
			return errors.New("organization ID not found")
		}

		keysResp, err := annotationsAPI.ListKeys(&annotations.ListKeysRequest{
			OrganizationID: orgID,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list annotation keys: %w", err)
		}

		var keyID string

		for _, key := range keysResp.Keys {
			if key.Name == "iam_terraform_identifier" {
				keyID = key.ID

				break
			}
		}

		if keyID == "" {
			return nil
		}

		valuesResp, err := annotationsAPI.ListValues(&annotations.ListValuesRequest{
			KeyID: &keyID,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list annotation values: %w", err)
		}

		var valueID string

		for _, value := range valuesResp.Values {
			if value.Name == annotationValue {
				valueID = value.ID

				break
			}
		}

		if valueID == "" {
			return nil
		}

		bindingsResp, err := annotationsAPI.ListBindings(&annotations.ListBindingsRequest{
			OrganizationID: orgID,
			ValueID:        &valueID,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("failed to list bindings: %w", err)
		}

		for _, binding := range bindingsResp.Bindings {
			if err := annotationsAPI.DeleteBinding(&annotations.DeleteBindingRequest{
				BindingID: binding.ID,
			}); err != nil && !httperrors.Is404(err) {
				return fmt.Errorf("failed to delete binding %s: %w", binding.ID, err)
			}
		}

		if err := annotationsAPI.DeleteValue(&annotations.DeleteValueRequest{
			ValueID: valueID,
		}); err != nil && !httperrors.Is404(err) {
			return fmt.Errorf("failed to delete annotation value %s: %w", valueID, err)
		}

		return nil
	}
}
