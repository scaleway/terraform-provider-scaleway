package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceRegistryNamespace_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRegistryNamespaceBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_registry_namespace" "test" {
						name = "test-cr"
					}
					
					data "scaleway_registry_namespace" "test" {
						name = scaleway_registry_namespace.test.name
					}
					
					data "scaleway_registry_namespace" "test2" {
						namespace_id = scaleway_registry_namespace.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRegistryNamespaceExists(tt, "scaleway_registry_namespace.test"),

					resource.TestCheckResourceAttr("scaleway_registry_namespace.test", "name", "test-cr"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test", "id"),
					resource.TestCheckResourceAttr("data.scaleway_registry_namespace.test", "is_public", "false"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test", "endpoint"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test", "namespace_id"),

					resource.TestCheckResourceAttr("data.scaleway_registry_namespace.test2", "name", "test-cr"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test2", "id"),
					resource.TestCheckResourceAttr("data.scaleway_registry_namespace.test2", "is_public", "false"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test2", "endpoint"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test2", "namespace_id"),
				),
			},
			{
				Config: `
					resource "scaleway_registry_namespace" "test" {
						name = "test-cr"
						is_public = "true"
					}
					
					data "scaleway_registry_namespace" "test" {
						name = scaleway_registry_namespace.test.name
					}
					
					data "scaleway_registry_namespace" "test2" {
						namespace_id = scaleway_registry_namespace.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRegistryNamespaceExists(tt, "scaleway_registry_namespace.test"),

					resource.TestCheckResourceAttr("scaleway_registry_namespace.test", "name", "test-cr"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test", "id"),
					resource.TestCheckResourceAttr("data.scaleway_registry_namespace.test", "is_public", "true"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test", "endpoint"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test", "namespace_id"),

					resource.TestCheckResourceAttr("data.scaleway_registry_namespace.test2", "name", "test-cr"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test2", "id"),
					resource.TestCheckResourceAttr("data.scaleway_registry_namespace.test2", "is_public", "true"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test2", "endpoint"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace.test2", "namespace_id"),
				),
			},
		},
	})
}
