package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
)

func TestRegistryNamespaceBeta(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayRegistryNamespaceBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_registry_namespace_beta cr01 {
						name = "test-cr"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRegistryNamespaceBetaExists("scaleway_registry_namespace_beta.cr01"),
					resource.TestCheckResourceAttr("scaleway_registry_namespace_beta.cr01", "name", "test-cr"),
					testCheckResourceAttrUUID("scaleway_registry_namespace_beta.cr01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_registry_namespace_beta cr01 {
						name = "test-cr"
						description = "test registry namespace"
						is_public = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRegistryNamespaceBetaExists("scaleway_registry_namespace_beta.cr01"),
					resource.TestCheckResourceAttr("scaleway_registry_namespace_beta.cr01", "description", "test registry namespace"),
					resource.TestCheckResourceAttr("scaleway_registry_namespace_beta.cr01", "is_public", "true"),
					testCheckResourceAttrUUID("scaleway_registry_namespace_beta.cr01", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayRegistryNamespaceBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := registryAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return nil
		}

		_, err = api.GetNamespace(&registry.GetNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayRegistryNamespaceBetaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_registry_namespace_beta" {
			continue
		}

		api, region, id, err := registryAPIWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.DeleteNamespace(&registry.DeleteNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})

		if err == nil {
			return fmt.Errorf("namespace (%s) still exists", rs.Primary.ID)
		}

		if !is404Error(err) {
			return err
		}
	}

	return nil
}
