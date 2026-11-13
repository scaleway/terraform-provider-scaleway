package keymanager_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDecryptEphemeralResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccDecryptEphemeralResource_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	plainTextData := "this is some secret data"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "main" {
				  name        = "tf-test-decrypt-key"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true
				}

				ephemeral "scaleway_key_manager_encrypt" "test_encrypt" {
				  key_id     = scaleway_key_manager_key.main.id
				  plaintext  = "%s"
				  region     = "fr-par"
				}

				ephemeral "scaleway_key_manager_decrypt" "test_decrypt" {
				  key_id     = scaleway_key_manager_key.main.id
				  ciphertext = ephemeral.scaleway_key_manager_encrypt.test_encrypt.ciphertext
				  region     = "fr-par"
				  depends_on = [ephemeral.scaleway_key_manager_encrypt.test_encrypt]
				}

				provider "echo" {
				  data = ephemeral.scaleway_key_manager_decrypt.test_decrypt
				}

				resource "echo" "test_decrypt" {}
				`, plainTextData),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_decrypt", dataPath.AtMapKey("plaintext"), knownvalue.StringExact(plainTextData)),
				},
			},
		},
	})
}

func TestAccDecryptEphemeralResource_WithAssociatedData(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccDecryptEphemeralResource_WithAssociatedData because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	plainTextData := "this is some secret data"
	associatedData := "some associated data"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "test_key" {
				  name        = "tf-test-decrypt-key-ad"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true
				}

				ephemeral "scaleway_key_manager_encrypt" "test_encrypt" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  plaintext  = "%[1]s"
				  region     = "fr-par"
				  associated_data = {
					value = "%[2]s"
				  }
				}

				ephemeral "scaleway_key_manager_decrypt" "test_decrypt" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  ciphertext = ephemeral.scaleway_key_manager_encrypt.test_encrypt.ciphertext
				  region     = "fr-par"
				  associated_data = {
					value = "%[2]s"
				  }
				}

				provider "echo" {
				  data = ephemeral.scaleway_key_manager_decrypt.test_decrypt
				}

				resource "echo" "test_decrypt" {}
				`, plainTextData, associatedData),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_decrypt", dataPath.AtMapKey("plaintext"), knownvalue.StringExact(plainTextData)),
					statecheck.ExpectKnownValue("echo.test_decrypt", dataPath.AtMapKey("associated_data").AtMapKey("value"), knownvalue.StringExact(associatedData)),
				},
			},
		},
	})
}

func TestAccDecryptEphemeralResource_InvalidAssociatedData(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccDecryptEphemeralResource_InvalidAssociatedData because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	plainTextData := "this is secret"
	associatedData := "some associated data"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "test_key" {
				  name        = "tf-test-decrypt-key-invalid"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true
				}

				ephemeral "scaleway_key_manager_encrypt" "test_encrypt" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  plaintext  = "%s"
				  region     = "fr-par"
				  associated_data = {
					value = "%s"
				  }
				}

				ephemeral "scaleway_key_manager_decrypt" "test_decrypt" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  ciphertext = ephemeral.scaleway_key_manager_encrypt.test_encrypt.ciphertext
				  region     = "fr-par"
				  associated_data = {
					value = "qwerty"
				  }
				}
				`, plainTextData, associatedData),
				ExpectError: regexp.MustCompile("cipher: message authentication failed"),
			},
		},
	})
}
