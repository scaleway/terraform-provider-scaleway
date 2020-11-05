package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceRegistryNamespace_Basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayRegistryNamespaceBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_registry_namespace_beta" "test" {
						name = "test-cr"
					}
					
					data "scaleway_registry_namespace_beta" "test" {
						name = scaleway_registry_namespace_beta.test.name
					}
					
					data "scaleway_registry_namespace_beta" "test2" {
						namespace_id = scaleway_registry_namespace_beta.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRegistryNamespaceBetaExists("scaleway_registry_namespace_beta.test"),

					resource.TestCheckResourceAttr("scaleway_registry_namespace_beta.test", "name", "test-cr"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test", "id"),
					resource.TestCheckResourceAttr("data.scaleway_registry_namespace_beta.test", "is_public", "false"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test", "endpoint"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test", "namespace_id"),

					resource.TestCheckResourceAttr("data.scaleway_registry_namespace_beta.test2", "name", "test-cr"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test2", "id"),
					resource.TestCheckResourceAttr("data.scaleway_registry_namespace_beta.test2", "is_public", "false"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test2", "endpoint"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test2", "namespace_id"),
				),
			},
			{
				Config: `
					resource "scaleway_registry_namespace_beta" "test" {
						name = "test-cr"
						is_public = "true"
					}
					
					data "scaleway_registry_namespace_beta" "test" {
						name = scaleway_registry_namespace_beta.test.name
					}
					
					data "scaleway_registry_namespace_beta" "test2" {
						namespace_id = scaleway_registry_namespace_beta.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRegistryNamespaceBetaExists("scaleway_registry_namespace_beta.test"),

					resource.TestCheckResourceAttr("scaleway_registry_namespace_beta.test", "name", "test-cr"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test", "id"),
					resource.TestCheckResourceAttr("data.scaleway_registry_namespace_beta.test", "is_public", "true"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test", "endpoint"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test", "namespace_id"),

					resource.TestCheckResourceAttr("data.scaleway_registry_namespace_beta.test2", "name", "test-cr"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test2", "id"),
					resource.TestCheckResourceAttr("data.scaleway_registry_namespace_beta.test2", "is_public", "true"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test2", "endpoint"),
					resource.TestCheckResourceAttrSet("data.scaleway_registry_namespace_beta.test2", "namespace_id"),
				),
			},
			{
				ResourceName:      "data.scaleway_registry_namespace_beta.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "data.scaleway_registry_namespace_beta.test2",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
