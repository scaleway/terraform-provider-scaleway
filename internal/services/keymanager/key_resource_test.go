package keymanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/keymanager"
)

func TestAccKeyManagerKey_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name         = "tf-test-kms-key-unprotected-a"
				  region       = "fr-par"
				  usage        = "symmetric_encryption"
				  description  = "Test key"
				  tags         = ["tf", "test"]
				  unprotected  = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-unprotected-a"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "region", "fr-par"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "usage", "symmetric_encryption"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "description", "Test key"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "tags.0", "tf"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "tags.1", "test"),
				),
			},
		},
	})
}

func TestAccKeyManagerKey_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name         = "tf-test-kms-key-update"
				  region       = "fr-par"
				  usage        = "symmetric_encryption"
				  description  = "Test key"
				  tags         = ["tf", "test"]
				  unprotected  = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-update"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "description", "Test key"),
				),
			},
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name         = "tf-test-kms-key-updated"
				  region       = "fr-par"
				  usage        = "symmetric_encryption"
				  description  = "Test key updated"
				  tags         = ["tf", "updated"]
				  unprotected  = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-updated"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "description", "Test key updated"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "tags.1", "updated"),
				),
			},
		},
	})
}

func IsKeyManagerKeyDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_key_manager_key" {
				continue
			}

			client, region, keyID, err := keymanager.NewKeyManagerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			key, err := client.GetKey(&key_manager.GetKeyRequest{
				Region: region,
				KeyID:  keyID,
			})
			if err == nil {
				if key.DeletionRequestedAt != nil {
					continue
				}

				return fmt.Errorf("Key (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
