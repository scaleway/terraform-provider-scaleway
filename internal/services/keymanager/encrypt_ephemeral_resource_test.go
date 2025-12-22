package keymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccEncryptEphemeralResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccEncryptEphemeralResource_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
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

				resource "scaleway_secret" "main" {
					name        = "test-encrypted-secret"
				}

				resource "scaleway_secret_version" "v1" {
					description = "version1"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_encrypt.test_encrypt.ciphertext
					depends_on	= [ephemeral.scaleway_key_manager_encrypt.test_encrypt]
				}

				data "scaleway_secret_version" "data_v1" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.v1]
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_secret_version.data_v1", "data"),
				),
			},
		},
	})
}
