package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_container_namespace", &resource.Sweeper{
		Name: "scaleway_container_namespace",
		F:    testSweepContainerNamespace,
	})
}

func testSweepContainerNamespace(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		containerAPI := container.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the container namespaces in (%s)", region)
		listNamespaces, err := containerAPI.ListNamespaces(
			&container.ListNamespacesRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing namespaces in (%s) in sweeper: %s", region, err)
		}

		for _, ns := range listNamespaces.Namespaces {
			_, err := containerAPI.DeleteNamespace(&container.DeleteNamespaceRequest{
				NamespaceID: ns.ID,
				Region:      region,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting namespace in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayContainerNamespace_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerNamespaceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "test-cr-ns-01"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "test-cr-ns-01"
						description = "test container namespace 01"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "description", "test container namespace 01"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "test-cr-ns-01"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "test-cr-ns-01"
						environment_variables = {
							"test" = "test"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "test-cr-ns-01"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "environment_variables.test", "test"),

					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttrSet("scaleway_container_namespace.main", "name"),
					resource.TestCheckResourceAttrSet("scaleway_container_namespace.main", "registry_endpoint"),
					resource.TestCheckResourceAttrSet("scaleway_container_namespace.main", "registry_namespace_id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "tf-env-test"
						environment_variables = {
							"test" = "test"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "environment_variables.test", "test"),

					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "tf-env-test"
						environment_variables = {
							"foo" = "bar"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "environment_variables.foo", "bar"),

					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayContainerNamespaceExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return nil
		}

		_, err = api.GetNamespace(&container.GetNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayContainerNamespaceDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_namespace" {
				continue
			}

			api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteNamespace(&container.DeleteNamespaceRequest{
				NamespaceID: id,
				Region:      region,
			})

			if err == nil {
				return fmt.Errorf("container namespace (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
