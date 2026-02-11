package secret_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
		secretVersionDataV3 = "my_super_secret_v3"
	)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
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

				# We set the value of the ephemeral resource in a scaleway_secret_version,
				# to be able to call its data source and retrieve the ephemeral values
				# in the checks.
				# This is not something that should be done in a real configuration. 
				resource "scaleway_secret_version" "v1_from_ephemeral" {
				  description = "version1_from_ephemeral"
				  secret_id   = scaleway_secret.main.id
				  data_wo        = base64decode(ephemeral.scaleway_secret_version.data_v1.data)
				  data_wo_version = 3
				  depends_on = [ephemeral.scaleway_secret_version.data_v1]
				}

				resource "scaleway_secret_version" "v2_from_ephemeral" {
				  description 		= "version2_from_ephemeral"
				  secret_id   		= scaleway_secret.main.id
				  data_wo        	= base64decode(ephemeral.scaleway_secret_version.data_v2.data)
				  data_wo_version	= 4
				  depends_on = [scaleway_secret_version.v1_from_ephemeral]
				}

				resource "scaleway_secret_version" "latest_from_ephemeral" {
				  description		= "version2_from_ephemeral"
				  secret_id			= scaleway_secret.main.id
				  data_wo			= base64decode(ephemeral.scaleway_secret_version.data_v2.data)
				  data_wo_version	= 5
				  depends_on = [scaleway_secret_version.v2_from_ephemeral]
				}

				# We use datasources to be able to retrieve the ephemeral values in the checks.
				# This is not something that should be done in a real configuration. 
				data "scaleway_secret_version" "v1_from_ephemeral" {
					secret_id 	= scaleway_secret.main.id
					revision  	= "3"
					depends_on 	= [scaleway_secret_version.v1_from_ephemeral]
				}

				data "scaleway_secret_version" "v2_from_ephemeral" {
					secret_id 	= scaleway_secret.main.id
					revision  	= "4"
					depends_on 	= [scaleway_secret_version.v2_from_ephemeral]
				} 

				data "scaleway_secret_version" "latest_from_ephemeral" {
					secret_id 	= scaleway_secret.main.id
					revision  	= "5"
					depends_on 	= [scaleway_secret_version.latest_from_ephemeral]
				} 
				`, secretName, secretDescription, secretVersionData, secretVersionDataV2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.v1_from_ephemeral", "data", secret.Base64Encoded([]byte(secretVersionData))),
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v2"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.v2_from_ephemeral", "data", secret.Base64Encoded([]byte(secretVersionDataV2))),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.latest_from_ephemeral", "data", secret.Base64Encoded([]byte(secretVersionDataV2))),
					// Ensure data_wo (and data) and ephemeral secret_version are not in state
					testAccCheckAttributeNotInState("scaleway_secret_version.v1", "data_wo"),
					testAccCheckAttributeNotInState("scaleway_secret_version.v1", "data"),
					testAccCheckAttributeNotInState("scaleway_secret_version.v2", "data_wo"),
					testAccCheckAttributeNotInState("scaleway_secret_version.v2", "data"),
					testAccCheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_v1"),
					testAccCheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_v2"),
					testAccCheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_latest"),
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

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
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

				# We set the value of the ephemeral resource in a scaleway_secret_version,
				# to be able to call its data source and retrieve the ephemeral values
				# in the checks.
				# This is not something that should be done in a real configuration. 
				resource "scaleway_secret_version" "v1_from_ephemeral" {
				  description 		= "version1_from_ephemeral"
				  secret_id   		= scaleway_secret.main.id
				  data_wo        	= base64decode(ephemeral.scaleway_secret_version.data_by_name.data)
				  data_wo_version 	= 2
				  depends_on 		= [ephemeral.scaleway_secret_version.data_by_name]
				}

				resource "scaleway_secret_version" "by_name_from_ephemeral" {
				  description 		= "by_name_from_ephemeral"
				  secret_id   		= scaleway_secret.main.id
				  data_wo        	= base64decode(ephemeral.scaleway_secret_version.data_by_name_latest.data)
				  data_wo_version 	= 3
				  depends_on 		= [ephemeral.scaleway_secret_version.data_by_name_latest]
				}

				# We use datasources to be able to retrieve the ephemeral values in the checks.
				# This is not something that should be done in a real configuration. 
				data "scaleway_secret_version" "v1_from_ephemeral" {
					secret_id 	= scaleway_secret.main.id
					revision  	= "2"
					depends_on 	= [scaleway_secret_version.v1_from_ephemeral]
				}

				data "scaleway_secret_version" "by_name_from_ephemeral" {
					secret_id 	= scaleway_secret.main.id
					revision  	= "3"
					depends_on 	= [scaleway_secret_version.by_name_from_ephemeral]
				}
				`, secretName, secretVersionData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.v1_from_ephemeral"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.v1_from_ephemeral", "data", secret.Base64Encoded([]byte(secretVersionData))),
					testAccCheckSecretVersionExists(tt, "scaleway_secret_version.by_name_from_ephemeral"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.by_name_from_ephemeral", "data", secret.Base64Encoded([]byte(secretVersionData))),
					// Ensure data_wo (and data) and ephemeral secret_version are not in state
					testAccCheckAttributeNotInState("scaleway_secret_version.v1_from_ephemeral", "data_wo"),
					testAccCheckAttributeNotInState("scaleway_secret_version.v1_from_ephemeral", "data"),
					testAccCheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_by_name"),
					testAccCheckEphemeralResourceNotInState("ephemeral.scaleway_secret_version.data_by_name_latest"),
				),
			},
		},
	})
}

func testAccCheckEphemeralResourceNotInState(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if _, ok := state.RootModule().Resources[resource]; ok {
			return fmt.Errorf("ephemeral resource %s should not be persisted in state", resource)
		}

		return nil
	}
}
