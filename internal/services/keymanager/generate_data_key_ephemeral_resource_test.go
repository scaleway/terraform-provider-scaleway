package keymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccGenerateDataKeyEphemeralResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccGenerateDataKeyEphemeralResource_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: `
				resource "scaleway_key_manager_key" "test_key" {
				  name        = "tf-test-generate-data-key"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true
				}

				ephemeral "scaleway_key_manager_generate_data_key" "main" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  region     = "fr-par"
				}

				ephemeral "scaleway_key_manager_decrypt" "test_decrypt" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  ciphertext  = ephemeral.scaleway_key_manager_generate_data_key.main.ciphertext
				  region     = "fr-par"
				}

				provider "echo" {
				  alias = "generate_data_key"
				  data  = ephemeral.scaleway_key_manager_generate_data_key.main
				}

				provider "echo" {
				  alias = "decrypt"
				  data  = ephemeral.scaleway_key_manager_decrypt.test_decrypt
				}

				resource "echo" "test_generate_data_key" {
				  provider = echo.generate_data_key
				}

				resource "echo" "test_decrypt" {
				  provider = echo.decrypt
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_generate_data_key", dataPath.AtMapKey("ciphertext"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.test_generate_data_key", dataPath.AtMapKey("plaintext"), knownvalue.NotNull()),
					// Check that a generated data key plaintext is equivalent to calling decrypt with its ciphertext
					statecheck.CompareValuePairs(
						"echo.test_decrypt", dataPath.AtMapKey("plaintext"),
						"echo.test_generate_data_key", dataPath.AtMapKey("plaintext"),
						compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccGenerateDataKeyEphemeralResource_WithoutPlaintext(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccGenerateDataKeyEphemeralResource_WithoutPlaintext because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: `
				resource "scaleway_key_manager_key" "test_key" {
				  name        = "tf-test-generate-no-plaintext"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true
				}

				ephemeral "scaleway_key_manager_generate_data_key" "main" {
				  key_id    		= scaleway_key_manager_key.test_key.id
				  region     		= "fr-par"
				  without_plaintext = true
				}

				provider "echo" {
				  data = ephemeral.scaleway_key_manager_generate_data_key.main
				}

				resource "echo" "test_generate_data_key" {}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_generate_data_key", dataPath.AtMapKey("ciphertext"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("echo.test_generate_data_key", dataPath.AtMapKey("plaintext"), knownvalue.StringExact("")),
				},
			},
		},
	})
}
