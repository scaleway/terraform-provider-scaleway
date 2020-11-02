package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_registry_namespace_beta", &resource.Sweeper{
		Name: "scaleway_registry_namespace_beta",
		F:    testSweepRegistryNamespace,
	})
}

func testSweepRegistryNamespace(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar, scw.RegionNlAms}, func(scwClient *scw.Client, region scw.Region) error {
		registryAPI := registry.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the registry namespaces in (%s)", region)
		listNamespaces, err := registryAPI.ListNamespaces(&registry.ListNamespacesRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing namespaces in (%s) in sweeper: %s", region, err)
		}

		for _, ns := range listNamespaces.Namespaces {
			_, err := registryAPI.DeleteNamespace(&registry.DeleteNamespaceRequest{
				NamespaceID: ns.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting namespace in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayRegistryNamespace_Basic(t *testing.T) {
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
