package annotations_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceAnnotationsBinding_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsBindingDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "test_key_binding_ds"
						description = "Test annotation key for binding data source"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "test_value_binding_ds"
						description = "Test annotation value for binding data source"
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-binding-ds"
						region      = "fr-par"
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						unprotected = true
					}

					resource "scaleway_annotations_binding" "main" {
						srn      = scaleway_key_manager_key.main.srn
						value_id = scaleway_annotations_value.main.id
					}

					data "scaleway_annotations_binding" "main" {
						id = scaleway_annotations_binding.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_binding.main", "id", "scaleway_annotations_binding.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_binding.main", "srn", "scaleway_annotations_binding.main", "srn"),
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_binding.main", "value_id", "scaleway_annotations_binding.main", "value_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_binding.main", "key_id", "scaleway_annotations_binding.main", "key_id"),
				),
			},
		},
	})
}
