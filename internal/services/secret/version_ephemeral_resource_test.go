package secret_test

import (
	"fmt"
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
				// lintignore:AT004
				Config: fmt.Sprintf(`
				resource "scaleway_secret" "main" {
				  name        = "%[1]s"
				  description = "%[2]s"
				  tags        = ["devtools", "provider", "terraform"]
				}

				resource "scaleway_secret_version" "v1" {
				  description = "version1"
				  secret_id   = scaleway_secret.main.id
				  data_wo        = "%[3]s"
				  data_wo_version = 1
				  depends_on = [scaleway_secret.main]
				}

				resource "scaleway_secret_version" "v2" {
				  description = "version2"
				  secret_id   = scaleway_secret.main.id
				  data_wo        = "%[4]s"
				  data_wo_version = 2
				  depends_on = [scaleway_secret_version.v1]
				}

				ephemeral "scaleway_secret_version" "data_v1" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.v1]
				}

				ephemeral "scaleway_secret_version" "data_v2" {
				  secret_id = scaleway_secret.main.id
				  revision  = "2"
				  depends_on = [scaleway_secret_version.v2]
				}

				ephemeral "scaleway_secret_version" "data_latest" {
				  secret_id = scaleway_secret.main.id
				  revision  = "latest"
				  depends_on = [scaleway_secret_version.v2]
				}

				provider "echo" {
				  alias = "v1"
				  data  = ephemeral.scaleway_secret_version.data_v1
				}

				provider "echo" {
				  alias = "v2"
				  data  = ephemeral.scaleway_secret_version.data_v2
				}

				provider "echo" {
				  alias  = "latest"
				  data   = ephemeral.scaleway_secret_version.data_latest
				}

				resource "echo" "test_v1" {
				  provider = echo.v1
				}

				resource "echo" "test_v2" {
				  provider = echo.v2
				}

				resource "echo" "test_latest" {
				  provider = echo.latest
				}
				`, secretName, secretDescription, secretVersionData, secretVersionDataV2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_v1", dataPath.AtMapKey("data"), knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionData)))),
					statecheck.ExpectKnownValue("echo.test_v2", dataPath.AtMapKey("data"), knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionDataV2)))),
					statecheck.ExpectKnownValue("echo.test_latest", dataPath.AtMapKey("data"), knownvalue.StringExact(secret.Base64Encoded([]byte(secretVersionDataV2)))),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1"),
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v2"),
					// Ensure data_wo (and data) and ephemeral secret_version are not in state
					acctest.CheckAttributeNotInState("scaleway_secret_version.v1", "data_wo"),
					acctest.CheckAttributeNotInState("scaleway_secret_version.v1", "data"),
					acctest.CheckAttributeNotInState("scaleway_secret_version.v2", "data_wo"),
					acctest.CheckAttributeNotInState("scaleway_secret_version.v2", "data"),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_v1"),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_v2"),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_latest"),
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
				// lintignore:AT004
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
