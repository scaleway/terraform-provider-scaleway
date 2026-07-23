package annotations_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceAnnotationsValue_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             IsAnnotationsValueDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_annotations_key" "main" {
						name        = "test_key_value_ds"
						description = "Test annotation key for value data source"
					}

					resource "scaleway_annotations_value" "main" {
						key_id      = scaleway_annotations_key.main.id
						name        = "test_value_ds"
						description = "Test annotation value for data source"
					}

					data "scaleway_annotations_value" "main" {
						value_id = scaleway_annotations_value.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_value.main", "value_id", "scaleway_annotations_value.main", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_value.main", "key_id", "scaleway_annotations_value.main", "key_id"),
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_value.main", "name", "scaleway_annotations_value.main", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_annotations_value.main", "description", "scaleway_annotations_value.main", "description"),
				),
			},
		},
	})
}
