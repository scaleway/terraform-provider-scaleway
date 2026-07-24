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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

func TestAccAnnotationsValueResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsValueDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "tf_test_annotations_key_value"
						description = "Test annotation key for value"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "tf_test_annotations_value"
						description = "Test annotation value"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "name", "tf_test_annotations_value"),
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "description", "Test annotation value"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_value.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_value.main", "key_id"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_value.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAnnotationsValueResource_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsValueDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "tf_test_annotations_key_value_update"
						description = "Test annotation key for value update"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "tf_test_annotations_value_update"
						description = "Initial description"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "name", "tf_test_annotations_value_update"),
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "description", "Initial description"),
				),
			},
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "tf_test_annotations_key_value_update"
						description = "Test annotation key for value update"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "tf_test_annotations_value_update"
						description = "Updated description"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "name", "tf_test_annotations_value_update"),
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "description", "Updated description"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_value.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAnnotationsValueResource_Minimal(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsValueDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "tf_test_annotations_key_value_minimal"
					}

					resource "scaleway_annotations_value" "main" {
						key_id = scaleway_annotations_key.main.id
						name   = "tf_test_annotations_value_minimal"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "name", "tf_test_annotations_value_minimal"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_value.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_value.main", "key_id"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_value.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func IsAnnotationsValueDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_annotations_value" {
					continue
				}

				api := annotationsSDK.NewAPI(tt.Meta.ScwClient())
				_, err := api.GetValue(&annotationsSDK.GetValueRequest{
					ValueID: rs.Primary.ID,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("annotation value (%s) still exists", rs.Primary.ID))
				case httperrors.Is404(err):
					continue
				default:
					return retry.NonRetryableError(err)
				}
			}

			return nil
		})
	}
}
