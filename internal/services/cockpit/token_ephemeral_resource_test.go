package cockpit_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
	secrettestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret/testfuncs"
)

func TestAccTokenEphemeralResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccTokenEphemeralResource_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	tokenName := "tf-tests-cpt-tok-eph-basic"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             secrettestfuncs.CheckSecretDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					ephemeral "scaleway_cockpit_token" "main" {
						name = "%[1]s"
						scopes {
							query_metrics = true
							write_logs    = false
						}
					}

					resource "scaleway_secret" "main" {
						name = "%[1]s"
					}

					resource "scaleway_secret_version" "secret_key" {
						description = "%[1]s"
						secret_id   = scaleway_secret.main.id
						data_wo     = ephemeral.scaleway_cockpit_token.main.secret_key
					}

					data "scaleway_secret_version" "secret_key" {
						secret_id  = scaleway_secret.main.id
						revision   = "1"
						depends_on = [scaleway_secret_version.secret_key]
					}

					resource "scaleway_secret_version" "name" {
						description = "%[1]s"
						secret_id   = scaleway_secret.main.id
						data_wo     = ephemeral.scaleway_cockpit_token.main.name
						depends_on  = [scaleway_secret_version.secret_key]
					}

					data "scaleway_secret_version" "name" {
						secret_id  = scaleway_secret.main.id
						revision   = "2"
						depends_on = [scaleway_secret_version.name]
					}
				`, tokenName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEphemeralCockpitTokenSecretKeySet("data.scaleway_secret_version.secret_key"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.name", "data", secret.Base64Encoded([]byte(tokenName))),
					acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_cockpit_token.main"),
				),
			},
		},
	})
}

func testAccCheckEphemeralCockpitTokenSecretKeySet(dataSourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}

		encoded := rs.Primary.Attributes["data"]
		if encoded == "" {
			return fmt.Errorf("secret version data is empty for %s", dataSourceName)
		}

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return fmt.Errorf("failed to decode secret version data: %w", err)
		}

		if len(decoded) == 0 {
			return errors.New("decoded cockpit token secret_key is empty")
		}

		return nil
	}
}
