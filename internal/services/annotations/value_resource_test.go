package annotations_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	annotationstestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/annotations/testfuncs"
)

func TestAccAnnotationsValueResource_Basic(t *testing.T) {
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
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_annotations_key" "main" {
						name            = "tf_test_annotations_key_value"
						description     = "Test annotation key for value"
						organization_id = "%[1]s"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "tf_test_annotations_value"
						description = "Test annotation value"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "name", "tf_test_annotations_value"),
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "description", "Test annotation value"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_value.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_annotations_value.main", "key_id"),
					resource.TestCheckResourceAttrWith("scaleway_annotations_value.main", "srn", acctest.CheckSRN("annotations", "values")),
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

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			annotationstestfuncs.IsAnnotationsKeyDestroyed(tt),
			annotationstestfuncs.IsAnnotationsValueDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_annotations_key" "main" {
						name            = "tf_test_annotations_key_value_update"
						description     = "Test annotation key for value update"
						organization_id = "%[1]s"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "tf_test_annotations_value_update"
						description = "Initial description"
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "name", "tf_test_annotations_value_update"),
					resource.TestCheckResourceAttr("scaleway_annotations_value.main", "description", "Initial description"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_annotations_key" "main" {
						name            = "tf_test_annotations_key_value_update"
						description     = "Test annotation key for value update"
						organization_id = "%[1]s"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "tf_test_annotations_value_update"
						description = "Updated description"
					}
				`, orgID),
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

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			annotationstestfuncs.IsAnnotationsKeyDestroyed(tt),
			annotationstestfuncs.IsAnnotationsValueDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_annotations_key" "main" {
						name            = "tf_test_annotations_key_value_minimal"
						organization_id = "%[1]s"
					}

					resource "scaleway_annotations_value" "main" {
						key_id = scaleway_annotations_key.main.id
						name   = "tf_test_annotations_value_minimal"
					}
				`, orgID),
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
