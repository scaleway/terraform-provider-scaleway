package keymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccActionRotateKey_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccActionRotateKey_Basic because actions are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					locals {
						region = "fr-par"
					}
				
					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-rotation-action"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key"
						tags        = ["tf", "test"]
						unprotected = true
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_key_manager_key_rotate_action.main]
							}
						}
					}

					action "scaleway_key_manager_key_rotate_action" "main" {
						config {
							key_id = scaleway_key_manager_key.main.id
							region = local.region
						}
					}
				`,
			},
			{
				Config: `
					locals {
						region = "fr-par"
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-rotation-action"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key"
						tags        = ["tf", "test"]
						unprotected = true
						lifecycle {
							action_trigger {
								events  = [after_create]
								actions = [action.scaleway_key_manager_key_rotate_action.main]
							}
						}
					}

					action "scaleway_key_manager_key_rotate_action" "main" {
						config {
							key_id = scaleway_key_manager_key.main.id
							region = local.region
						}
					}

					data "scaleway_key_manager_key" "main" {
						key_id = scaleway_key_manager_key.main.id
						depends_on = [scaleway_key_manager_key.main]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.main", "name", "tf-test-kms-key-rotation-action"),
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.main", "rotation_count", "2"),
				),
			},
		},
	})
}
