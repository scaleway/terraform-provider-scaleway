package keymanager_test

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
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

	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsKeyManagerKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				// lintignore:AT004 // the "echo" provider block is required to capture ephemeral data
				Config: fmt.Sprintf(`
				resource "scaleway_key_manager_key" "test_key" {
					name        = "tf-test-sign-key"
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

				provider "echo" {
					data = ephemeral.scaleway_key_manager_sign.test_sign
				}

				resource "echo" "test_sign" {}
				`, digest),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("echo.test_sign", dataPath.AtMapKey("signature"), knownvalue.NotNull()),
				},
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
