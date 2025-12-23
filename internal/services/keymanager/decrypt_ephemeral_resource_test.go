package keymanager_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

func TestAccDecryptEphemeralResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	plainTextData := "this is some secret data"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
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
				}

				resource "scaleway_secret" "main" {
					name        = "test-decrypt-secret"
				}

				resource "scaleway_secret_version" "v1" {
					description = "test decrypted"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_decrypt.test_decrypt.plaintext
				}

				data "scaleway_secret_version" "v1" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.v1]
				}
				`, plainTextData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_secret_version.v1", "data", secret.Base64Encoded([]byte(plainTextData))),
				),
			},
		},
	})
}

func TestAccDecryptEphemeralResource_WithAssociatedData(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	plainTextData := "this is some secret data"
	associatedData := "some associated data"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "test_key" {
				  name        = "tf-test-decrypt-key"
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

				resource "scaleway_secret" "main" {
					name        = "test-decrypt-secret"
				}

				resource "scaleway_secret_version" "data" {
					description = "test decrypted data"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_decrypt.test_decrypt.plaintext
				}

				resource "scaleway_secret_version" "associated_data" {
					description 	= "test decrypted associated data"
					secret_id   	= scaleway_secret.main.id
					data_wo     	= ephemeral.scaleway_key_manager_decrypt.test_decrypt.associated_data.value
					data_wo_version = 2
					depends_on		= [scaleway_secret_version.data]
				}

				data "scaleway_secret_version" "data" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.data]
				}

				data "scaleway_secret_version" "associated_data" {
				  secret_id = scaleway_secret.main.id
				  revision  = "2"
				  depends_on = [scaleway_secret_version.associated_data]
				}
				`, plainTextData, associatedData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_secret_version.data", "data", secret.Base64Encoded([]byte(plainTextData))),
					resource.TestCheckResourceAttr("data.scaleway_secret_version.associated_data", "data", secret.Base64Encoded([]byte(associatedData))),
				),
			},
		},
	})
}

func TestAccDecryptEphemeralResource_ErrorWrongAssociatedData(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	plainTextData := "this is secret"
	associatedData := "some associated data"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "test_key" {
				  name        = "tf-test-decrypt-key"
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
