package scwconfig_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
	"github.com/stretchr/testify/require"
)

func TestAccEphemeralConfig_ActiveProfile(t *testing.T) {
	// Skip if we are running coverage tests on CI
	if accessKey := os.Getenv("SCW_ACCESS_KEY"); accessKey == ciAccessKey {
		t.Skip("Skipping TestAccEphemeralConfig_ActiveProfile")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	ctx := t.Context()

	dataPath := tfjsonpath.New("data")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: func() map[string]func() (tfprotov6.ProviderServer, error) {
			_ = os.Unsetenv("SCW_PROFILE")

			t.Setenv("SCW_CONFIG_PATH", "./testfixture/test_config.yaml")

			metaDefault, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			return map[string]func() (tfprotov6.ProviderServer, error){
				"default": func() (tfprotov6.ProviderServer, error) {
					providers, errProvider := provider.NewProviderList(ctx, &provider.Config{Meta: metaDefault})
					if errProvider != nil {
						return nil, errProvider
					}

					muxServer, errMux := tf6muxserver.NewMuxServer(ctx, providers...)
					if errMux != nil {
						return nil, errMux
					}

					return muxServer.ProviderServer(), nil
				},
				"echo": echoprovider.NewProviderServer(),
			}
		}(),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: `
					ephemeral "scaleway_config" "main" {
						provider = "default"
					}

					provider "echo" {
						data = ephemeral.scaleway_config.main
					}

					resource "echo" "test_config" {}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("access_key"), knownvalue.StringExact("SCWXXXXXXXXXXXXXXXXX")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("access_key_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("secret_key"), knownvalue.StringExact("01234567-abcd-effe-dcba-012345678910")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("secret_key_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("project_id"), knownvalue.StringExact("11111111-2222-3333-4444-555555555555")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("project_id_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("organization_id"), knownvalue.StringExact("11111111-2222-3333-4444-555555555555")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("organization_id_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("region"), knownvalue.StringExact("nl-ams")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("region_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("zone"), knownvalue.StringExact("nl-ams-1")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("zone_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
				},
			},
		},
	})
}

func TestAccEphemeralConfig_OtherProfile(t *testing.T) {
	// Skip if we are running coverage tests on CI
	if accessKey := os.Getenv("SCW_ACCESS_KEY"); accessKey == ciAccessKey {
		t.Skip("Skipping TestAccEphemeralConfig_OtherProfile")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	ctx := t.Context()

	dataPath := tfjsonpath.New("data")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: func() map[string]func() (tfprotov6.ProviderServer, error) {
			_ = os.Unsetenv("SCW_PROFILE")

			t.Setenv("SCW_CONFIG_PATH", "./testfixture/test_config.yaml")
			t.Setenv("SCW_PROFILE", "other")

			metaOther, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			return map[string]func() (tfprotov6.ProviderServer, error){
				"other": func() (tfprotov6.ProviderServer, error) {
					providers, errProvider := provider.NewProviderList(ctx, &provider.Config{Meta: metaOther})
					if errProvider != nil {
						return nil, errProvider
					}

					muxServer, errMux := tf6muxserver.NewMuxServer(ctx, providers...)
					if errMux != nil {
						return nil, errMux
					}

					return muxServer.ProviderServer(), nil
				},
				"echo": echoprovider.NewProviderServer(),
			}
		}(),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: `
					ephemeral "scaleway_config" "main" {
						provider = "other"
					}

					provider "echo" {
						data = ephemeral.scaleway_config.main
					}

					resource "echo" "test_config" {}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("access_key"), knownvalue.StringExact("SCWYYYYYYYYYYYYYYYYY")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("access_key_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("secret_key"), knownvalue.StringExact("99999999-9999-9999-9999-999999999999")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("secret_key_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("project_id"), knownvalue.StringExact("99999999-9999-9999-9999-999999999999")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("project_id_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("organization_id"), knownvalue.StringExact("99999999-9999-9999-9999-999999999999")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("organization_id_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("region"), knownvalue.StringExact("fr-par")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("region_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("zone"), knownvalue.StringExact("fr-par-1")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("zone_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
				},
			},
		},
	})
}

func TestAccEphemeralConfig_MixedProfile(t *testing.T) {
	// Skip if we are running coverage tests on CI
	if accessKey := os.Getenv("SCW_ACCESS_KEY"); accessKey == ciAccessKey {
		t.Skip("Skipping TestAccEphemeralConfig_MixedProfile")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	ctx := t.Context()

	dataPath := tfjsonpath.New("data")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: func() map[string]func() (tfprotov6.ProviderServer, error) {
			_ = os.Unsetenv("SCW_PROFILE")

			t.Setenv("SCW_CONFIG_PATH", "./testfixture/test_config.yaml")
			t.Setenv("SCW_PROFILE", "incomplete")
			t.Setenv("SCW_DEFAULT_PROJECT_ID", "77777777-7777-7777-7777-777777777777")

			metaMixed, err := meta.NewMeta(ctx, &meta.Config{
				TerraformVersion: "terraform-tests",
				HTTPClient:       tt.Meta.HTTPClient(),
			})
			require.NoError(t, err)

			return map[string]func() (tfprotov6.ProviderServer, error){
				"mixed": func() (tfprotov6.ProviderServer, error) {
					providers, errProvider := provider.NewProviderList(ctx, &provider.Config{Meta: metaMixed})
					if errProvider != nil {
						return nil, errProvider
					}

					muxServer, errMux := tf6muxserver.NewMuxServer(ctx, providers...)
					if errMux != nil {
						return nil, errMux
					}

					return muxServer.ProviderServer(), nil
				},
				"echo": echoprovider.NewProviderServer(),
			}
		}(),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004
				Config: `
					ephemeral "scaleway_config" "main" {
						provider = "mixed"
					}

					provider "echo" {
						data = ephemeral.scaleway_config.main
					}

					resource "echo" "test_config" {}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("access_key"), knownvalue.StringExact("SCW11111111111111111")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("access_key_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("secret_key"), knownvalue.StringExact("11111111-1111-1111-1111-111111111111")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("secret_key_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("project_id"), knownvalue.StringExact("77777777-7777-7777-7777-777777777777")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("project_id_source"), knownvalue.StringExact(meta.CredentialsSourceEnvironment)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("organization_id"), knownvalue.StringExact("11111111-2222-3333-4444-555555555555")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("organization_id_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("region"), knownvalue.StringExact("nl-ams")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("region_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("zone"), knownvalue.StringExact("pl-waw-1")),
					statecheck.ExpectKnownValue("echo.test_config", dataPath.AtMapKey("zone_source"), knownvalue.StringExact(meta.CredentialsSourceActiveProfile)),
				},
			},
		},
	})
}
