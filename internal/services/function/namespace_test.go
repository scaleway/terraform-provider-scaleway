package function_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	functionSDK "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/function"
)

func init() {
	resource.AddTestSweepers("scaleway_function_namespace", &resource.Sweeper{
		Name: "scaleway_function_namespace",
		F:    testSweepFunctionNamespace,
	})
}

func testSweepFunctionNamespace(_ string) error {
	return acctest.SweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		functionAPI := functionSDK.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the function namespaces in (%s)", region)
		listNamespaces, err := functionAPI.ListNamespaces(
			&functionSDK.ListNamespacesRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing namespaces in (%s) in sweeper: %s", region, err)
		}

		for _, ns := range listNamespaces.Namespaces {
			_, err := functionAPI.DeleteNamespace(&functionSDK.DeleteNamespaceRequest{
				NamespaceID: ns.ID,
				Region:      region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting namespace in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccFunctionNamespace_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionNamespaceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "test-cr-ns-01"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					acctest.CheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
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
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "description", "test function namespace 01"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "test-cr-ns-01"),
					acctest.CheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_function_namespace main {
						name = "test-cr-ns-01"
						environment_variables = {
							"test" = "test"
						}
						secret_environment_variables = {
							"test_secret" = "test_secret"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "test-cr-ns-01"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.test", "test"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "secret_environment_variables.test_secret", "test_secret"),

					acctest.CheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
		},
	})
}

func TestAccFunctionNamespace_NoName(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionNamespaceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
				),
			},
		},
	})
}

func TestAccFunctionNamespace_EnvironmentVariables(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionNamespaceDestroy(tt),
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
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.test", "test"),

					acctest.CheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
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
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.foo", "bar"),

					acctest.CheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
		},
	})
}

func testAccCheckFunctionNamespaceExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := function.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetNamespace(&functionSDK.GetNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckFunctionNamespaceDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_function_namespace" {
				continue
			}

			api, region, id, err := function.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteNamespace(&functionSDK.DeleteNamespaceRequest{
				NamespaceID: id,
				Region:      region,
			})

			if err == nil {
				return fmt.Errorf("function namespace (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
