package secret_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
	secrettestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret/testfuncs"
)

func TestAccEphemeralResourceSecretVersion_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccEphemeralResourceSecretVersion_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	const (
		secretName          = "ephemeralSecretVersionBasic"
		secretDescription   = "secret description"
		secretVersionData   = "my_super_secret_v1"
		secretVersionDataV2 = "my_super_secret_v2"
	)

	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%[1]s"
					  description = "%[2]s"
					  tags        = ["devtools", "provider", "terraform"]
					}

					resource "scaleway_secret_version" "latest" {
					  description = "version1"
					  secret_id   = scaleway_secret.main.id
					  data_wo        = "%[3]s"
					  data_wo_version = 1
					}

					ephemeral "scaleway_secret_version" "data" {
					  secret_id = scaleway_secret.main.id
					  revision  = scaleway_secret_version.latest.revision
					}

					provider "echo" {
					  data  = ephemeral.scaleway_secret_version.data
					}

					resource "echo" "v1" {}
				`, secretName, secretDescription, secretVersionData),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.v1", dataPath.AtMapKey("data"), knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionData)))),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.latest"),
					// Ensure data_wo (and data) and ephemeral secret_version are not in state
					acctest.CheckAttributeNotInState("scaleway_secret_version.latest", "data_wo"),
					acctest.CheckAttributeNotInState("scaleway_secret_version.latest", "data"),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data"),
				),
			},
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%[1]s"
					  description = "%[2]s"
					  tags        = ["devtools", "provider", "terraform"]
					}

					resource "scaleway_secret_version" "latest" {
					  description = "version2"
					  secret_id   = scaleway_secret.main.id
					  data_wo        = "%[3]s"
					  data_wo_version = 2
					}

					ephemeral "scaleway_secret_version" "data" {
					  secret_id = scaleway_secret.main.id
					  revision  = scaleway_secret_version.latest.revision
					}

					provider "echo" {
					  data  = ephemeral.scaleway_secret_version.data
					}

					resource "echo" "v2" {}
				`, secretName, secretDescription, secretVersionDataV2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.v2", dataPath.AtMapKey("data"), knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionDataV2)))),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.latest"),
					// Ensure data_wo (and data) and ephemeral secret_version are not in state
					acctest.CheckAttributeNotInState("scaleway_secret_version.latest", "data_wo"),
					acctest.CheckAttributeNotInState("scaleway_secret_version.latest", "data"),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data"),
				),
			},
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
					  name        = "%[1]s"
					  description = "%[2]s"
					  tags        = ["devtools", "provider", "terraform"]
					}

					resource "scaleway_secret_version" "latest" {
					  description = "version2"
					  secret_id   = scaleway_secret.main.id
					  data_wo        = "%[3]s"
					  data_wo_version = 2
					}

					ephemeral "scaleway_secret_version" "data" {
					  secret_id = scaleway_secret.main.id
					  revision  = scaleway_secret_version.latest.revision
					}

					provider "echo" {
					  data  = ephemeral.scaleway_secret_version.data
					}

					resource "echo" "latest" {}
				`, secretName, secretDescription, secretVersionDataV2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.latest", dataPath.AtMapKey("data"), knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionDataV2)))),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.latest"),
					// Ensure data_wo (and data) and ephemeral secret_version are not in state
					acctest.CheckAttributeNotInState("scaleway_secret_version.latest", "data_wo"),
					acctest.CheckAttributeNotInState("scaleway_secret_version.latest", "data"),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data"),
				),
			},
		},
	})
}

func TestAccEphemeralResourceSecretVersion_ByNameSecret(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccEphemeralResourceSecretVersion_ByNameSecret because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	const (
		secretVersionData = "my_super_secret_v1"
		secretName        = "erSecretVersionByNameSecret"
	)

	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name = "%[1]s"
				  tags = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data_wo        = "%[2]s"
				  data_wo_version = 1
				}

				ephemeral "scaleway_secret_version" "data_by_name" {
				  secret_name = scaleway_secret.main.name
				  revision    = "1"
				  project_id  = scaleway_secret.main.project_id
				  depends_on = [scaleway_secret_version.v1]
				}

				ephemeral "scaleway_secret_version" "data_by_name_latest" {
				  secret_name = scaleway_secret.main.name
				  revision    = "latest"
				  project_id  = scaleway_secret.main.project_id
				  depends_on = [scaleway_secret_version.v1]
				}

				provider "echo" {
				  alias = "by_name"
				  data  = ephemeral.scaleway_secret_version.data_by_name
				}

				provider "echo" {
				  alias = "by_name_latest"
				  data  = ephemeral.scaleway_secret_version.data_by_name_latest
				}

				resource "echo" "test_by_name" {
				  provider = echo.by_name
				}

				resource "echo" "test_by_name_latest" {
				  provider = echo.by_name_latest
				}
				`, secretName, secretVersionData),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_by_name", dataPath.AtMapKey("data"), knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionData)))),
					statecheck.ExpectKnownValue("echo.test_by_name_latest", dataPath.AtMapKey("data"), knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionData)))),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1"),
					// Ensure data_wo (and data) and ephemeral secret_version are not in state
					acctest.CheckAttributeNotInState("scaleway_secret_version.v1", "data_wo"),
					acctest.CheckAttributeNotInState("scaleway_secret_version.v1", "data"),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_by_name"),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_by_name_latest"),
				),
			},
		},
	})
}

func TestAccEphemeralResourceSecretVersion_ComputedAttributes(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccEphemeralResourceSecretVersion_ComputedAttributes because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	const (
		secretName        = "ephemeralSecretVersionComputedAttributes"
		secretDescription = "secret description"
		secretVersionData = "my_super_secret_v1"
	)

	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
					resource "scaleway_secret" "main" {
						name        = "%[1]s"
						description = "%[2]s"
						tags        = ["devtools", "provider", "terraform"]
					}

					resource "scaleway_secret_version" "v1" {
						description     = "version1"
						secret_id       = scaleway_secret.main.id
						data_wo         = "%[3]s"
						data_wo_version = 1
						depends_on      = [scaleway_secret.main]
					}

					ephemeral "scaleway_secret_version" "data_computed_attributes" {
						secret_id  = scaleway_secret.main.id
						depends_on = [scaleway_secret_version.v1]
					}

					provider "echo" {
						data = ephemeral.scaleway_secret_version.data_computed_attributes
					}

					resource "echo" "test_computed_attributes" {}
				`, secretName, secretDescription, secretVersionData),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("revision"),
						knownvalue.StringExact("latest")),
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("region"),
						knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("srn"),
						knownvalue.StringRegexp(regexp.MustCompile(`^srn://secret-manager\..+/regions/.+/secrets/.+/versions/1$`))),
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("secret_id"),
						knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("description"),
						knownvalue.StringExact("version1")),
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("status"),
						knownvalue.StringExact("enabled")),
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("created_at"),
						knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("updated_at"),
						knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.test_computed_attributes",
						dataPath.AtMapKey("data"),
						knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionData)))),
				},
			},
		},
	})
}
