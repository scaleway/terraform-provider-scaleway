package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceContainerNamespace_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerNamespaceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_container_namespace" "main" {
						name = "test-cr-data"
					}
					
					data "scaleway_container_namespace" "by_name" {
						name = scaleway_container_namespace.main.name
					}
					
					data "scaleway_container_namespace" "by_id" {
						namespace_id = scaleway_container_namespace.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "test-cr-data"),
					resource.TestCheckResourceAttrSet("data.scaleway_container_namespace.by_name", "id"),

					resource.TestCheckResourceAttr("data.scaleway_container_namespace.by_id", "name", "test-cr-data"),
					resource.TestCheckResourceAttrSet("data.scaleway_container_namespace.by_id", "id"),
				),
			},
		},
	})
}
