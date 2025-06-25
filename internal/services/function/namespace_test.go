package function_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	functionSDK "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/function"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

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
						tags = ["tag1", "tag2"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					acctest.CheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "tags.1", "tag2"),
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
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "tags.#", "0"),
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
					passwordMatchHash("scaleway_function_namespace.main", "secret_environment_variables.test_secret", "test_secret"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "tags.#", "0"),

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
                                                tags = ["tag1", "tag2"]
                                        }
                               `,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "test-cr-ns-01"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.test", "test"),
					passwordMatchHash("scaleway_function_namespace.main", "secret_environment_variables.test_secret", "test_secret"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "tags.1", "tag2"),

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
						secret_environment_variables = {
							"test_secret" = "test_secret"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.test", "test"),
					passwordMatchHash("scaleway_function_namespace.main", "secret_environment_variables.test_secret", "test_secret"),
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
						secret_environment_variables = {
							"test_secret" = "updated_secret"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "environment_variables.foo", "bar"),
					passwordMatchHash("scaleway_function_namespace.main", "secret_environment_variables.test_secret", "updated_secret"),
					acctest.CheckResourceAttrUUID("scaleway_function_namespace.main", "id"),
				),
			},
		},
	})
}

func TestAccFunctionNamespace_VPCIntegration(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	namespaceID := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckFunctionNamespaceDestroy(tt),
			testAccCheckFunctionDestroy(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network main {}
			
					resource scaleway_function_namespace main {}
			
					resource scaleway_function main {
						namespace_id = scaleway_function_namespace.main.id
						sandbox = "v1"
						privacy = "private"
						runtime = "go123"
						handler = "Handle"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "activate_vpc_integration", "false"),
					acctest.CheckResourceIDPersisted("scaleway_function_namespace.main", &namespaceID),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network main {}
			
					resource scaleway_function_namespace main {}
			
					resource scaleway_function main {
						namespace_id = scaleway_function_namespace.main.id
						privacy = "private"
						sandbox = "v1"
						runtime = "go123"
						handler = "Handle"
						private_network_id = scaleway_vpc_private_network.main.id
					}
				`,
				ExpectError: regexp.MustCompile("Application can't be attached to private network, vpc integration must be activated on its parent namespace"),
			},
			{
				Config: `
					resource scaleway_vpc_private_network main {}

					resource scaleway_function_namespace main {
						activate_vpc_integration = true
					}

					resource scaleway_function main {
						namespace_id = scaleway_function_namespace.main.id
						sandbox = "v1"
						privacy = "private"
						runtime = "go123"
						handler = "Handle"
						private_network_id = scaleway_vpc_private_network.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionNamespaceExists(tt, "scaleway_function_namespace.main"),
					testAccCheckFunctionExists(tt, "scaleway_function.main"),
					resource.TestCheckResourceAttr("scaleway_function_namespace.main", "activate_vpc_integration", "true"),
					resource.TestCheckResourceAttrPair("scaleway_function.main", "private_network_id", "scaleway_vpc_private_network.main", "id"),
					acctest.CheckResourceIDChanged("scaleway_function_namespace.main", &namespaceID),
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
