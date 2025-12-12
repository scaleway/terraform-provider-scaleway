package keymanager_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/keymanager"
	"github.com/stretchr/testify/assert"
)

func TestAccEncryptEphemeralResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "scaleway_key_manager_key" "test_key" {
				  name        = "tf-test-encrypt-key"
				  region      = "fr-par"
				  usage       = "symmetric_encryption"
				  algorithm   = "aes_256_gcm"
				  unprotected = true
				}

				ephemeral "scaleway_key_manager_encrypt" "test_encrypt" {
				  key_id     = scaleway_key_manager_key.test_key.id
				  plaintext  = "test plaintext data"
				  region     = "fr-par"
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"output.ciphertext",
						tfjsonpath.New("value"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestEncryptEphemeralResource_Open(t *testing.T) {
	r := keymanager.NewEncryptEphemeralResource()

	data := keymanager.EncryptEphemeralResourceModel{
		KeyID:     types.StringValue("fr-par/11111111-1111-1111-1111-111111111111"),
		Plaintext: types.StringValue("test plaintext data"),
		Region:    types.StringValue("fr-par"),
	}

	req := ephemeral.OpenRequest{Config: tfsdk.Config{Raw: &data}}
	var resp ephemeral.OpenResponse

	r.Open(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.Result.Get("ciphertext").StringValue() != "")
}
