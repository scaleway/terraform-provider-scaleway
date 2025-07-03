package keymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccKeyManagerKey_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "main" {
				  name         = "tf-test-kms-key-unprotected"
				  region       = "fr-par"
				  usage        = "symmetric_encryption"
				  description  = "Test key"
				  tags         = ["tf", "test"]
				  unprotected  = true
				}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-unprotected"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "region", "fr-par"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "usage", "symmetric_encryption"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "description", "Test key"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "tags.0", "tf"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "tags.1", "test"),
				),ad
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
