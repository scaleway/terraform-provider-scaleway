package cockpit_test

import (
	"fmt"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccTokenEphemeralResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccTokenEphemeralResource_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	tokenName := "tf-tests-cpt-tok-eph-basic"
	dataPath := tfjsonpath.New("data")

	factories := maps.Clone(tt.ProviderFactories)
	factories["echo"] = echoprovider.NewProviderServer()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: fmt.Sprintf(`
					ephemeral "scaleway_cockpit_token" "main" {
						name = "%[1]s"
						scopes {
							query_metrics = true
							write_logs    = false
						}
					}

					provider "echo" {
						data = ephemeral.scaleway_cockpit_token.main
					}

					resource "echo" "token" {}
				`, tokenName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.token", dataPath.AtMapKey("name"), knownvalue.StringExact(tokenName)),
					statecheck.ExpectKnownValue("echo.token", dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.token", dataPath.AtMapKey("project_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.token", dataPath.AtMapKey("region"), knownvalue.NotNull()),
				},
				Check: acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_cockpit_token.main"),
			},
		},
	})
}

func TestAccTokenEphemeralResource_AllScopesDisabled(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccTokenEphemeralResource_AllScopesDisabled because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	tokenName := "tf-tests-cpt-tok-eph-all-off"
	dataPath := tfjsonpath.New("data")
	scopesPath := dataPath.AtMapKey("scopes").AtSliceIndex(0)

	factories := maps.Clone(tt.ProviderFactories)
	factories["echo"] = echoprovider.NewProviderServer()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: fmt.Sprintf(`
					ephemeral "scaleway_cockpit_token" "main" {
						name = "%[1]s"
						scopes {
							query_metrics       = false
							write_metrics       = false
							setup_metrics_rules = false
							query_logs          = false
							write_logs          = false
							setup_logs_rules    = false
							setup_alerts        = false
							query_traces        = false
							write_traces        = false
						}
					}

					provider "echo" {
						data = ephemeral.scaleway_cockpit_token.main
					}

					resource "echo" "token" {}
				`, tokenName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.token", dataPath.AtMapKey("name"), knownvalue.StringExact(tokenName)),
					statecheck.ExpectKnownValue("echo.token", dataPath.AtMapKey("secret_key"), knownvalue.NotNull()),
					// Explicit false must not be overridden by write_metrics / write_logs defaults.
					statecheck.ExpectKnownValue("echo.token", scopesPath.AtMapKey("write_metrics"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("echo.token", scopesPath.AtMapKey("write_logs"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("echo.token", scopesPath.AtMapKey("query_metrics"), knownvalue.Bool(false)),
				},
				Check: acctest.CheckEphemeralResourceNotInState("ephemeral.scaleway_cockpit_token.main"),
			},
		},
	})
}
