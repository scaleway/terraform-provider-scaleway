package keymanager_test

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func createKeyMaterial() string {
	return base64.StdEncoding.EncodeToString([]byte("a987654321fedcba987654321fedcba"))
}

func createSalt() string {
	return base64.StdEncoding.EncodeToString([]byte("saltvalue16bytes"))
}

func TestAccKeyMaterialResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_Basic because write-only fields are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					locals {
						region = "fr-par"
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-material"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key with external key material"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id       = scaleway_key_manager_key.main.id
						region       = local.region
						key_material = "%s"
					}
				`, createKeyMaterial()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-material"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "origin", "external"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "state", "enabled"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key_material.main", "key_state", "enabled"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key_material.main", "origin", "external"),
				),
			},
		},
	})
}

func TestAccKeyMaterialResource_WithSalt(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_WithSalt because write-only fields are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					locals {
						region = "fr-par"
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-material-salt"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key with salt"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id                  = scaleway_key_manager_key.main.id
						region                  = local.region
						key_material_wo         = "%s"
						key_material_wo_version = 1
						salt_wo                 = "%s"
						salt_wo_version         = 1
					}
				`, createKeyMaterial(), createSalt()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-material-salt"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "origin", "external"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "state", "enabled"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key_material.main", "key_state", "enabled"),
				),
			},
		},
	})
}

func TestAccKeyMaterialResource_WriteOnly(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_WriteOnly because write-only fields are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					locals {
						region = "fr-par"
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-material-wo"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key with write-only key material"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id                 = scaleway_key_manager_key.main.id
						region                 = local.region
						key_material_wo        = "%s"
						key_material_wo_version = 1
					}
				`, createKeyMaterial()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-material-wo"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "origin", "external"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "state", "enabled"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key_material.main", "key_state", "enabled"),
				),
			},
		},
	})
}

func TestAccKeyMaterialResource_WithRegionalID(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_WithRegionalID because write-only fields are not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-regional"
						region      = "fr-par"
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key with regional ID"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id       = scaleway_key_manager_key.main.id
						key_material = "%s"
					}
				`, createKeyMaterial()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-regional"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "origin", "external"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "state", "enabled"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key_material.main", "key_state", "enabled"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key_material.main", "region", "fr-par"),
				),
			},
		},
	})
}

func TestAccKeyMaterialResource_RawBytes(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_RawBytes because write-only fields are not yet supported on OpenTofu")
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
						name        = "tf-test-kms-key-raw-bytes"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key with raw bytes key material"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id       = scaleway_key_manager_key.main.id
						region       = local.region
						# Using a simple raw string (not base64-encoded)
						key_material = "my-secret-key-material-32bytes!!"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-raw-bytes"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "origin", "external"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "state", "enabled"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key_material.main", "key_state", "enabled"),
				),
			},
		},
	})
}

func TestAccKeyMaterialResource_RawBytesWithSalt(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_RawBytesWithSalt because write-only fields are not yet supported on OpenTofu")
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
						name        = "tf-test-kms-key-raw-salt"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key with raw salt"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id       = scaleway_key_manager_key.main.id
						region       = local.region
						key_material = "my-secret-key-material-32bytes!!"
						salt         = "my-salt-16bytes!"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "name", "tf-test-kms-key-raw-salt"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "origin", "external"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key.main", "state", "enabled"),
					resource.TestCheckResourceAttr("scaleway_key_manager_key_material.main", "key_state", "enabled"),
				),
			},
		},
	})
}

func TestAccKeyMaterialResource_ValidationError(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_ValidationError because write-only fields are not yet supported on OpenTofu")
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
						name        = "tf-test-kms-key-conflict"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key for conflict test"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id        = scaleway_key_manager_key.main.id
						region        = local.region
						key_material  = "test-key-material"
						key_material_wo = "test-key-material-wo"
					}
				`,
				ExpectError: regexp.MustCompile("key_material and key_material_wo cannot both be specified"),
			},
		},
	})
}

func TestAccKeyMaterialResource_NoKeyMaterial(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_NoKeyMaterial because write-only fields are not yet supported on OpenTofu")
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
						name        = "tf-test-kms-key-no-material"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key for missing material test"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id = scaleway_key_manager_key.main.id
						region = local.region
					}
				`,
				ExpectError: regexp.MustCompile("key_material and key_material_wo cannot both be specified"),
			},
		},
	})
}

func TestAccKeyMaterialResource_SaltConflict(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_SaltConflict because write-only fields are not yet supported on OpenTofu")
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
						name        = "tf-test-kms-key-salt-conflict"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key for salt conflict test"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id       = scaleway_key_manager_key.main.id
						region       = local.region
						key_material = "test-key-material"
						salt         = "test-salt"
						salt_wo      = "test-salt-wo"
					}
				`,
				ExpectError: regexp.MustCompile("salt and salt_wo cannot both be specified"),
			},
		},
	})
}

func TestAccKeyMaterialResource_MissingSaltWoVersion(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_MissingSaltWoVersion because write-only fields are not yet supported on OpenTofu")
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
						name        = "tf-test-kms-key-missing-version"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key for missing version test"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id       = scaleway_key_manager_key.main.id
						region       = local.region
						key_material = "test-key-material"
						salt_wo      = "test-salt-wo"
					}
				`,
				ExpectError: regexp.MustCompile("salt_wo_version is required"),
			},
		},
	})
}

func TestAccKeyMaterialResource_MissingKeyMaterialWoVersion(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccKeyMaterialResource_MissingKeyMaterialWoVersion because write-only fields are not yet supported on OpenTofu")
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
						name        = "tf-test-kms-key-missing-wo-version"
						region      = local.region
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key for missing wo version test"
						tags        = ["tf", "test"]
						origin      = "external"
						unprotected = true
					}

					resource "scaleway_key_manager_key_material" "main" {
						key_id        = scaleway_key_manager_key.main.id
						region        = local.region
						key_material_wo = "test-key-material-wo"
					}
				`,
				ExpectError: regexp.MustCompile("key_material_wo_version is required"),
			},
		},
	})
}
