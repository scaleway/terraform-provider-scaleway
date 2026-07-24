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

var DestroyWaitTimeout = 3 * time.Minute

func TestAccAnnotationsKeyResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "tf_test_annotations_key"
						description = "Test annotation key"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "name", "tf_test_annotations_key"),
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "description", "Test annotation key"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_key.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_key.main", "organization_id"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_key.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAnnotationsKeyResource_WithOrganizationID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_annotations_key" "main" {
						name            = "tf_test_annotations_key_org"
						description     = "Test annotation key with organization"
						organization_id = "%s"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "name", "tf_test_annotations_key_org"),
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "description", "Test annotation key with organization"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_key.main", "id"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_key.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAnnotationsKeyResource_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "tf_test_annotations_key_update"
						description = "Initial description"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "name", "tf_test_annotations_key_update"),
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "description", "Initial description"),
				),
			},
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "tf_test_annotations_key_update"
						description = "Updated description"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "name", "tf_test_annotations_key_update"),
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "description", "Updated description"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_key.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAnnotationsKeyResource_Minimal(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name = "tf_test_annotations_key_minimal"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_key.main", "name", "tf_test_annotations_key_minimal"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_key.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_key.main", "organization_id"),
				),
			},
			{
				ResourceName:      "scaleway_annotations_key.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func IsAnnotationsKeyDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ctx := context.Background()

		return retry.RetryContext(ctx, DestroyWaitTimeout, func() *retry.RetryError {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "scaleway_annotations_key" {
					continue
				}

				api := annotationsSDK.NewAPI(tt.Meta.ScwClient())
				_, err := api.GetKey(&annotationsSDK.GetKeyRequest{
					KeyID: rs.Primary.ID,
				})

				switch {
				case err == nil:
					return retry.RetryableError(fmt.Errorf("annotation key (%s) still exists", rs.Primary.ID))
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
