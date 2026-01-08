package keymanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceVerify_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccDataSourceVerify_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	digest, err := createDigest("this is a message to sign")
	if err != nil {
		t.Errorf("failed to create digest: %v", err.Error())
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "main" {
					name        = "tf-test-verify-key"
					region      = "fr-par"
					usage       = "asymmetric_signing"
					algorithm   = "rsa_pss_2048_sha256"
					unprotected = true
				}

				ephemeral "scaleway_key_manager_sign" "test_verify" {
					key_id		= scaleway_key_manager_key.main.id
					digest		= "%[1]s"
					region		= "fr-par"
				}

				resource "scaleway_secret" "main" {
					name        = "test-verify-secret"
				}

				resource "scaleway_secret_version" "v1" {
					description = "version1"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_sign.test_verify.signature
				}

				data "scaleway_secret_version" "data_v1" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.v1]
				}

				data "scaleway_key_manager_verify" "main" {
				  key_id = scaleway_key_manager_key.main.id
				  region = "fr-par"
				  digest = "%[1]s"
				  signature = data.scaleway_secret_version.data_v1.data
				}
				`, digest),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_key_manager_verify.main", "valid", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceVerify_Invalid(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccDataSourceVerify_Invalid because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	digest, err := createDigest("this is a message to sign")
	if err != nil {
		t.Errorf("failed to create digest: %v", err.Error())
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "main" {
					name        = "tf-test-verify-key-invalid"
					region      = "fr-par"
					usage       = "asymmetric_signing"
					algorithm   = "rsa_pss_2048_sha256"
					unprotected = true
				}

				resource "scaleway_key_manager_key" "other" {
					name        = "tf-test-verify-other-key"
					region      = "fr-par"
					usage       = "asymmetric_signing"
					algorithm   = "rsa_pss_2048_sha256"
					unprotected = true
					depends_on  = [scaleway_key_manager_key.main]
				}

				ephemeral "scaleway_key_manager_sign" "test_verify" {
					key_id		= scaleway_key_manager_key.main.id
					digest		= "%[1]s"
					region		= "fr-par"
				}

				resource "scaleway_secret" "main" {
					name        = "test-verify-secret-invalid"
				}

				resource "scaleway_secret_version" "v1" {
					description = "version1"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_sign.test_verify.signature
				}

				data "scaleway_secret_version" "data_v1" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.v1]
				}

				data "scaleway_key_manager_verify" "main" {
				  key_id = scaleway_key_manager_key.other.id
				  region = "fr-par"
				  digest = "%[1]s"
				  signature = data.scaleway_secret_version.data_v1.data
				}
				`, digest),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_key_manager_verify.main", "valid", "false"),
				),
			},
		},
	})
}
