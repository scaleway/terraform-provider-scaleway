package keymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccEncryptEphemeralResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccEncryptEphemeralResource_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
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
				  name        = "tf-test-encrypt-key"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true
				}

				ephemeral "scaleway_key_manager_encrypt" "test_encrypt" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  plaintext  = "test plaintext data"
				  region     = "fr-par"
				}

				provider "echo" {
				  data = ephemeral.scaleway_key_manager_encrypt.test_encrypt
				}

				resource "echo" "test_encrypt" {}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_encrypt", dataPath.AtMapKey("plaintext"), knownvalue.StringExact("test plaintext data")),
					statecheck.ExpectKnownValue("echo.test_encrypt", dataPath.AtMapKey("ciphertext"), knownvalue.NotNull()),
				},
			},
		},
	})
}
