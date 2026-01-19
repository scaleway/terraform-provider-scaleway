package keymanager_test

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	secrettestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret/testfuncs"
)

func TestAccSignEphemeralResource_Basic(t *testing.T) {
	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping TestAccSignEphemeralResource_Basic because testing Ephemeral Resources is not yet supported on OpenTofu")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	digest, err := createDigest("this is a message to sign")
	if err != nil {
		t.Errorf("failed to create digest: %v", err.Error())
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			IsKeyManagerKeyDestroyed(tt),
			secrettestfuncs.CheckSecretDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "test_key" {
					name        = "tf-test-encrypt-key"
					region      = "fr-par"
					usage       = "asymmetric_signing"
					algorithm   = "rsa_pss_2048_sha256"
					unprotected = true
				}

				ephemeral "scaleway_key_manager_sign" "test_sign" {
					key_id		= scaleway_key_manager_key.test_key.id
					digest		= "%s"
					region		= "fr-par"
				}

				resource "scaleway_secret" "main" {
					name        = "test-sign-secret"
				}

				resource "scaleway_secret_version" "v1" {
					description = "version1"
					secret_id   = scaleway_secret.main.id
					data_wo     = ephemeral.scaleway_key_manager_sign.test_sign.signature
				}

				data "scaleway_secret_version" "data_v1" {
				  secret_id = scaleway_secret.main.id
				  revision  = "1"
				  depends_on = [scaleway_secret_version.v1]
				}
				`, digest),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_secret_version.data_v1", "data"),
				),
			},
		},
	})
}

func createDigest(message string) (string, error) {
	plaintext := []byte(message)

	digest := sha256.New()
	if _, err := digest.Write(plaintext); err != nil {
		return "", fmt.Errorf("failed to create digest: %w", err)
	}

	digestB64 := base64.StdEncoding.EncodeToString(digest.Sum(nil))

	return digestB64, nil
}
