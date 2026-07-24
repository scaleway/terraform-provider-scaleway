package annotations_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceAnnotationsKey_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	orgID, orgIDExists := tt.Meta.ScwClient().GetDefaultOrganizationID()
	if !orgIDExists {
		t.Skip("No default organization ID found, skipping test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsKeyDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_annotations_key" "main" {
						organization_id = "%[1]s"
						name            = "test_key_ds"
						description     = "Test annotation key for data source"
					}

					data "scaleway_annotations_key" "main" {
						key_id = scaleway_annotations_key.main.id
					}
				`, orgID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_key.main", "key_id", "scaleway_annotations_key.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_key.main", "name", "scaleway_annotations_key.main", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_key.main", "description", "scaleway_annotations_key.main", "description"),
				),
			},
		},
	})
}
