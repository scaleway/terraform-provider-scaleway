package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
)

func TestContainerRegistry(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayContainerRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_registry cr01 {
						name = "test-cr"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerRegistryExists("scaleway_container_registry.cr01"),
					resource.TestCheckResourceAttr("scaleway_container_registry.cr01", "name", "test-cr"),
					testCheckResourceAttrUUID("scaleway_container_registry.cr01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_registry cr01 {
						name = "test-cr"
						description = "test container repository"
						is_public = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerRegistryExists("scaleway_container_registry.cr01"),
					testCheckResourceAttrUUID("scaleway_container_registry.cr01", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayContainerRegistryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := containerRegistryWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
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

func testAccCheckScalewayContainerRegistryDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_container_registry" {
			continue
		}

		api, region, id, err := containerRegistryWithRegionAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.DeleteNamespace(&registry.DeleteNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})

		if err == nil {
			return fmt.Errorf("Namespace (%s) still exists", rs.Primary.ID)
		}

		if !is404Error(err) {
			return err
		}
	}

	return nil
}
