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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name        = "tf-test-kms-key-unprotected"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  description = "Test key"
				  tags        = ["tf", "test"]
				  unprotected = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-unprotected"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "region", "fr-par"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "usage", "symmetric_encryption"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "algorithm", "aes_256_gcm"),
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
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name        = "tf-test-kms-key-update"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  description = "Test key"
				  tags        = ["tf", "test"]
				  unprotected = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-update"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "description", "Test key"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "algorithm", "aes_256_gcm"),
				),
			},
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name        = "tf-test-kms-key-updated"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  description = "Test key updated"
				  tags        = ["tf", "updated"]
				  unprotected = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-updated"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "description", "Test key updated"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "tags.1", "updated"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "algorithm", "aes_256_gcm"),
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

func TestAccKeyManagerKey_WithRotationPolicy(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name        = "tf-test-kms-key-rotation"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  description = "Test key with rotation policy"
				  unprotected = true
				  
				  rotation_policy {
				    rotation_period = "876000h"
					next_rotation_at = "2027-01-01T00:00:00Z"
				  }
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-rotation"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "usage", "symmetric_encryption"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "algorithm", "aes_256_gcm"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "description", "Test key with rotation policy"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "rotation_policy.0.rotation_period", "876000h0m0s"),
				),
			},
		},
	})
}

func TestAccKeyManagerKey_RotationPolicyDrift(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	config := `
	resource "scaleway_key_manager_key" "main" {
	  name        = "tf-test-kms-key-rotation-drift"
	  region      = "fr-par"
	  usage       = "symmetric_encryption"
	  algorithm   = "aes_256_gcm"
	  unprotected = true

	  rotation_policy {
	    rotation_period = "720h"
	  }
	}
	`

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "rotation_policy.0.rotation_period", "720h0m0s"),
					resource.TestCheckResourceAttrSet("scaleway_key_manager_key.main", "rotation_policy.0.next_rotation_at"),
				),
			},
			{
				// next_rotation_at is not set in the config: re-planning must
				// not produce a diff, otherwise every apply postpones the
				// scheduled rotation.
				Config:   config,
				PlanOnly: true,
			},
			{
				ResourceName:      "scaleway_key_manager_key.main",
				ImportState:       true,
				ImportStateVerify: true,
				// unprotected is not returned by the API (only the computed
				// protected attribute is), so it cannot survive an import.
				ImportStateVerifyIgnore: []string{"unprotected"},
			},
			{
				// Changing only rotation_period must let the API recompute the
				// schedule instead of re-sending the value carried over from
				// state (the cassette body matcher fails on replay otherwise).
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name        = "tf-test-kms-key-rotation-drift"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true

				  rotation_policy {
				    rotation_period = "8760h"
				  }
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "rotation_policy.0.rotation_period", "8760h0m0s"),
					resource.TestCheckResourceAttrSet("scaleway_key_manager_key.main", "rotation_policy.0.next_rotation_at"),
				),
			},
			{
				// An explicitly configured next_rotation_at must be sent to
				// the API and round-trip unchanged.
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name        = "tf-test-kms-key-rotation-drift"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true

				  rotation_policy {
				    rotation_period  = "8760h"
				    next_rotation_at = "2027-03-01T00:00:00Z"
				  }
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "rotation_policy.0.rotation_period", "8760h0m0s"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "rotation_policy.0.next_rotation_at", "2027-03-01T00:00:00Z"),
				),
			},
		},
	})
}

func TestAccKeyManagerKey_WithCustomAlgorithm(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "rsa_4096" {
				  name        = "tf-test-kms-key-rsa4096"
				  region      = "fr-par"
				  usage       = "asymmetric_encryption"
				  algorithm   = "rsa_oaep_4096_sha256"
				  description = "Test key with RSA-4096 algorithm"
				  unprotected = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.rsa_4096", "name", "tf-test-kms-key-rsa4096"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.rsa_4096", "usage", "asymmetric_encryption"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.rsa_4096", "algorithm", "rsa_oaep_4096_sha256"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.rsa_4096", "description", "Test key with RSA-4096 algorithm"),
				),
			},
		},
	})
}

func TestAccKeyManagerKey_DefaultAlgorithm(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "default_alg" {
				  name        = "tf-test-kms-key-default-alg"
				  region      = "fr-par"
				  usage       = "asymmetric_encryption"
				  algorithm   = "rsa_oaep_3072_sha256"
				  description = "Test key with RSA-3072 algorithm"
				  unprotected = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.default_alg", "name", "tf-test-kms-key-default-alg"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.default_alg", "usage", "asymmetric_encryption"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.default_alg", "algorithm", "rsa_oaep_3072_sha256"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.default_alg", "description", "Test key with RSA-3072 algorithm"),
				),
			},
		},
	})
}
