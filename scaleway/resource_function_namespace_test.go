package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_function_namespace", &resource.Sweeper{
		Name: "scaleway_function_namespace",
		F:    testSweepFunctionNamespace,
	})
}

func testSweepFunctionNamespace(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		functionAPI := function.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the function namespaces in (%s)", region)
		listNamespaces, err := functionAPI.ListNamespaces(
			&function.ListNamespacesRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing namespaces in (%s) in sweeper: %s", region, err)
		}

		for _, ns := range listNamespaces.Namespaces {
			_, err := functionAPI.DeleteNamespace(&function.DeleteNamespaceRequest{
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

func TestAccScalewayFunctionNamespace_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionNamespaceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "test-cr-ns-01"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					testCheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "test-cr-ns-01"
						description = "test function namespace 01"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "description", "test function namespace 01"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "test-cr-ns-01"),
					testCheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "test-cr-ns-01"
						environment_variables = {
							"test" = "test"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "description", "test function namespace 01"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "test-cr-ns-01"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.test", "test"),

					testCheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
		},
	})
}

func TestAccScalewayFunctionNamespace_EnvironmentVariables(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionNamespaceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "tf-env-test"
						environment_variables = {
							"test" = "test"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.test", "test"),

					testCheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "tf-env-test"
						environment_variables = {
							"foo" = "bar"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.foo", "bar"),

					testCheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayFunctionNamespaceExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := functionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return nil
		}

		_, err = api.GetNamespace(&function.GetNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayFunctionNamespaceDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_function_namespace" {
				continue
			}

			api, region, id, err := functionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteNamespace(&function.DeleteNamespaceRequest{
				NamespaceID: id,
				Region:      region,
			})

			if err == nil {
				return fmt.Errorf("function namespace (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
