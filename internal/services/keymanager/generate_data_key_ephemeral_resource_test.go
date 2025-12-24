package keymanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

func TestAccGenerateDataKeyEphemeralResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccGenerateDataKeyEphemeralResource_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
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

				resource "scaleway_secret" "main" {
					name        = "test-data-key-secret"
				}

				resource "scaleway_secret_version" "ciphertext" {
					description = "version1"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_generate_data_key.main.ciphertext
				}

				resource "scaleway_secret_version" "plaintext" {
					description = "version2"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_generate_data_key.main.plaintext
					data_wo_version = 2
					depends_on	= [scaleway_secret_version.ciphertext]
				}

				data "scaleway_secret_version" "ciphertext" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.ciphertext]
				}

				data "scaleway_secret_version" "plaintext" {
				  secret_id = scaleway_secret.main.id
				  revision  = "2"
				  depends_on = [scaleway_secret_version.plaintext]
				}

				ephemeral "scaleway_key_manager_decrypt" "test_decrypt" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  ciphertext  = ephemeral.scaleway_key_manager_generate_data_key.main.ciphertext
				  region     = "fr-par"
				}

				resource "scaleway_secret_version" "decrypted" {
					description = "version3"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_decrypt.test_decrypt.plaintext
					data_wo_version = 3
					depends_on	= [scaleway_secret_version.plaintext]
				}

				data "scaleway_secret_version" "decrypted" {
				  secret_id = scaleway_secret.main.id
				  revision  = "3"
				  depends_on = [scaleway_secret_version.decrypted]
				}

				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_secret_version.ciphertext", "data"),
					resource.TestCheckResourceAttrSet("data.scaleway_secret_version.plaintext", "data"),
					// Check that a generated data key plaintext is equivalent to calling decrypt with its ciphertext
					resource.TestCheckResourceAttrPair("data.scaleway_secret_version.decrypted", "data", "data.scaleway_secret_version.plaintext", "data"),
				),
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

	emptyPlaintext := "plaintext was empty"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "test_key" {
				  name        = "tf-test-generate-data-key"
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

				resource "scaleway_secret" "main" {
					name        = "test-data-key-secret-no-plaintext"
				}

				resource "scaleway_secret_version" "ciphertext" {
					description = "version1"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_generate_data_key.main.ciphertext
				}

				resource "scaleway_secret_version" "plaintext" {
					description = "version2"
					secret_id   = scaleway_secret.main.id
  					data_wo     = ephemeral.scaleway_key_manager_generate_data_key.main.plaintext != "" ? ephemeral.scaleway_key_manager_generate_data_key.main.plaintext : "%s"
					data_wo_version = 2
					depends_on	= [scaleway_secret_version.ciphertext]
				}

				data "scaleway_secret_version" "ciphertext" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.ciphertext]
				}

				data "scaleway_secret_version" "plaintext" {
				  secret_id = scaleway_secret.main.id
				  revision  = "2"
				  depends_on = [scaleway_secret_version.plaintext]
				}
				`, emptyPlaintext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_secret_version.ciphertext", "data"),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.plaintext", "data", secret.Base64Encoded([]byte(emptyPlaintext))),
				),
			},
		},
	})
}
