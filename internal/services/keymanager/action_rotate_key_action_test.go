package keymanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	kmSDK "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/keymanager"
)

func isKeyRotated(tt *acctest.TestTools, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		api, region, keyID, err := keymanager.NewKeyManagerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		key, err := api.GetKey(&kmSDK.GetKeyRequest{
			Region: region,
			KeyID:  keyID,
		}, scw.WithContext(context.Background()))
		if err != nil || key == nil {
			return fmt.Errorf("failed to get key: %w", err)
		}

		if key.RotationCount != 1 {
			return fmt.Errorf("key %s rotation count is %d, expected 1", rs.Primary.ID, key.RotationCount)
		}

		return nil
	}
}

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
					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-rotation-action"
						region      = "fr-par"
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
							region = scaleway_key_manager_key.main.region
						}
					}
				`,
				Check: isKeyRotated(tt, "scaleway_key_manager_key.main"),
			},
		},
	})
}
