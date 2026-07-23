package annotations_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccAnnotationsBindingResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsBindingDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "tf_test_annotations_key_binding"
						description = "Test annotation key for binding"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "tf_test_annotations_value_binding"
						description = "Test annotation value for binding"
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-binding"
						region      = "fr-par"
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						description = "Test key for binding"
						tags        = ["tf", "test"]
						unprotected = true
					}

					resource "scaleway_annotations_binding" "main" {
						srn        = scaleway_key_manager_key.main.srn
						value_id   = scaleway_annotations_value.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_annotations_binding.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_annotations_binding.main", "srn", "scaleway_key_manager_key.main", "srn"),
					resource.TestCheckResourceAttrPair("scaleway_annotations_binding.main", "value_id", "scaleway_annotations_value.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_annotations_binding.main", "key_id", "scaleway_annotations_key.main", "id"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_binding.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAnnotationsBindingResource_Minimal(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsBindingDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name = "tf_test_annotations_key_binding_minimal"
					}

					resource "scaleway_annotations_value" "main" {
						key_id = scaleway_annotations_key.main.id
						name   = "tf_test_annotations_value_binding_minimal"
					}

					resource "scaleway_key_manager_key" "main" {
						name        = "tf-test-kms-key-binding-minimal"
						region      = "fr-par"
						usage       = "symmetric_encryption"
						algorithm   = "aes_256_gcm"
						unprotected = true
					}

					resource "scaleway_annotations_binding" "main" {
						srn      = scaleway_key_manager_key.main.srn
						value_id = scaleway_annotations_value.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_annotations_binding.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_annotations_binding.main", "srn", "scaleway_key_manager_key.main", "srn"),
					resource.TestCheckResourceAttrPair("scaleway_annotations_binding.main", "value_id", "scaleway_annotations_value.main", "id"),
					resource.TestCheckResourceAttrPair("scaleway_annotations_binding.main", "key_id", "scaleway_annotations_key.main", "id"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_binding.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func IsAnnotationsBindingDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_annotations_binding" {
					continue
				}

				api := annotationsSDK.NewAPI(tt.Meta.ScwClient())

				bindings, err := api.ListBindings(&annotationsSDK.ListBindingsRequest{})
				if err != nil {
					return retry.NonRetryableError(err)
				}

				for _, binding := range bindings.Bindings {
					if binding.ID == rs.Primary.ID {
						return retry.RetryableError(fmt.Errorf("annotation binding (%s) still exists", rs.Primary.ID))
					}
				}

				continue
			}

			return nil
		})
	}
}
