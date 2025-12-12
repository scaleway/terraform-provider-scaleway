package keymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceKey_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name        = "tf-test-kms-key-unprotected-a"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  description = "Test key"
				  tags        = ["tf", "test"]
				  unprotected = true
				}

				data "scaleway_key_manager_key" "by_id" {
				  key_id = scaleway_key_manager_key.main.id
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.by_id", "name", "tf-test-kms-key-unprotected-a"),
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.by_id", "region", "fr-par"),
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.by_id", "usage", "symmetric_encryption"),
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.by_id", "algorithm", "aes_256_gcm"),
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.by_id", "description", "Test key"),
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.by_id", "tags.0", "tf"),
					resource.TestCheckResourceAttr("data.scaleway_key_manager_key.by_id", "tags.1", "test"),
				),
			},
		},
	})
}
