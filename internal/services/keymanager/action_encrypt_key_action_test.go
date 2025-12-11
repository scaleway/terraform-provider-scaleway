package keymanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// TODO: test with associated data
// TODO test with symmetric_encryption, asymmetric_encryption, and error if else

func TestAccActionEncrypt_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRotateKey_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	plaintext := "this is a test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					locals {
						region = "fr-par"
					}

					ephemeral "scaleway_key_manager_key_encrypt" "my_data" {
						plaintext = %s
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-encrypt-action"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key"
						tags        = ["tf", "test"]
						unprotected = true
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_key_manager_key_encrypt_action.main]
							}
						}
					}

					action "scaleway_key_manager_key_encrypt_action" "main" {
						config {
							key_id = scaleway_key_manager_key.main.id
							region = local.region
							plaintext_wo = scaleway_key_manager_key_encrypt.my_data.plaintext
						}
					}

					output "action_result" {
						value = action.scaleway_key_manager_key_encrypt_action.main.output_attribute
					}
				`, plaintext),
			},
			{
				Config: fmt.Sprintf(`
					locals {
						region = "fr-par"
					}

					ephemeral "scaleway_key_manager_key_encrypt" "my_data" {
						plaintext = %s
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-encrypt-action"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key"
						tags        = ["tf", "test"]
						unprotected = true
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_key_manager_key_encrypt_action.main]
							}
						}
					}

					action "scaleway_key_manager_key_encrypt_action" "main" {
						config {
							key_id = scaleway_key_manager_key.main.id
							region = local.region
							plaintext_wo = scaleway_key_manager_key_encrypt.my_data.plaintext
						}
					}

					output "action_result" {
						value = action.scaleway_key_manager_key_encrypt_action.main.output_attribute
					}

					data "scaleway_key_manager_key" "main" {
						key_id = scaleway_key_manager_key.main.id
						depends_on = [scaleway_key_manager_key.main]
					}
				`, plaintext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.main", "name", "tf-test-kms-key-rotation-action"),
					resource.TestCheckResourceAttrSet(my_action_output, "ciphertext"),
				),
			},
		},
	})
}
