package annotations_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	annotationstestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/annotations/testfuncs"
	keymanagertestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/keymanager/testfuncs"
)

func TestAccAnnotationsBindingResource_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			annotationstestfuncs.IsAnnotationsKeyDestroyed(tt),
			annotationstestfuncs.IsAnnotationsValueDestroyed(tt),
			annotationstestfuncs.IsAnnotationsBindingDestroyed(tt),
			keymanagertestfuncs.IsKeyManagerKeyDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_annotations_key" "main" {
						name            = "tf_test_annotations_key_binding"
						description     = "Test annotation key for binding"
						organization_id = "%[1]s"
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
						target_srn = scaleway_key_manager_key.main.srn
						value_id   = scaleway_annotations_value.main.id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_annotations_binding.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_binding.main", "srn"),
					resource.TestCheckResourceAttrPair("scaleway_annotations_binding.main", "target_srn", "scaleway_key_manager_key.main", "srn"),
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
