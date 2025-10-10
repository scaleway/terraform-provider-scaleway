package container_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	containerSDK "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	registrySDK "github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/container"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/registry"
	vpcchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc/testfuncs"
)

const containerNamespaceResource = "scaleway_container_namespace"

func TestAccNamespace_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isNamespaceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "test-cr-ns-01"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
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
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "description", "test container namespace 01"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "test-cr-ns-01"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
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
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "test-cr-ns-01"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "environment_variables.test", "test"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.test_secret", "test_secret"),

					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
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
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "description", ""),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "test-cr-ns-01"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "environment_variables.test", "test"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.test_secret", "test_secret"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.1", "tag2"),

					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttrSet("scaleway_container_namespace.main", "name"),
					resource.TestCheckResourceAttrSet("scaleway_container_namespace.main", "registry_endpoint"),
					resource.TestCheckResourceAttrSet("scaleway_container_namespace.main", "registry_namespace_id"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.#", "0"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
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
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "environment_variables.test", "test"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.test_secret", "test_secret"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.#", "0"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "tf-env-test"
						environment_variables = {
							"foo" = "bar"
						}
						secret_environment_variables = {
							"foo_secret" = "bar_secret"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "tf-env-test"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "environment_variables.foo", "bar"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.foo_secret", "bar_secret"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.#", "0"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "tf-tags-test"
						tags = ["tag1", "tag2"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "tf-tags-test"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "tags.1", "tag2"),

					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
		},
	})
}

func TestAccNamespace_SecretManagement(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isNamespaceDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "test-secret-ns"
						secret_environment_variables = {
							"SECRET_1" = "value1"
							"SECRET_2" = "value2"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.SECRET_1", "value1"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.SECRET_2", "value2"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "test-secret-ns"
						secret_environment_variables = {
							"SECRET_1" = "value1"
							"SECRET_2" = "updated_value2"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.SECRET_1", "value1"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.SECRET_2", "updated_value2"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
						name = "test-secret-ns"
						secret_environment_variables = {
							"SECRET_KEY_1" = "value1"
							"SECRET_2" = "updated_value2"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.SECRET_KEY_1", "value1"),
					resource.TestCheckNoResourceAttr("scaleway_container_namespace.main", "secret_environment_variables.SECRET_1"),
					passwordMatchHash("scaleway_container_namespace.main", "secret_environment_variables.SECRET_2", "updated_value2"),
				),
			},
		},
	})
}

func TestAccNamespace_DestroyRegistry(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isNamespaceDestroyed(tt),
			isRegistryDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
						region = "pl-waw"
						name = "test-cr-ns-01"
						destroy_registry = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					acctest.CheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
		},
	})
}

func TestAccNamespace_VPCIntegration(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	namespaceID := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			isNamespaceDestroyed(tt),
			isContainerDestroyed(tt),
			vpcchecks.CheckPrivateNetworkDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc main {
						name = "TestAccNamespace_VPCIntegration"
					}

					resource scaleway_vpc_private_network main {
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						sandbox = "v1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "activate_vpc_integration", "true"),
					acctest.CheckResourceIDPersisted("scaleway_container_namespace.main", &namespaceID),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "TestAccNamespace_VPCIntegration"
					}

					resource scaleway_vpc_private_network main {
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_container_namespace main {
						activate_vpc_integration = true
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						private_network_id = scaleway_vpc_private_network.main.id
						sandbox = "v1"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isNamespacePresent(tt, "scaleway_container_namespace.main"),
					isContainerPresent(tt, "scaleway_container.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "activate_vpc_integration", "true"),
					resource.TestCheckResourceAttrPair("scaleway_container.main", "private_network_id", "scaleway_vpc_private_network.main", "id"),
					acctest.CheckResourceIDPersisted("scaleway_container_namespace.main", &namespaceID),
				),
			},
		},
	})
}

func isNamespacePresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := container.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetNamespace(&containerSDK.GetNamespaceRequest{
			NamespaceID: id,
			Region:      region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isNamespaceDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != containerNamespaceResource {
				continue
			}

			api, region, id, err := container.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteNamespace(&containerSDK.DeleteNamespaceRequest{
				NamespaceID: id,
				Region:      region,
			})
			if err == nil {
				return fmt.Errorf("container namespace (%s) still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}

func isRegistryDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != containerNamespaceResource {
				continue
			}

			api, region, _, err := registry.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteNamespace(&registrySDK.DeleteNamespaceRequest{
				NamespaceID: rs.Primary.Attributes["registry_namespace_id"],
				Region:      region,
			})
			if err == nil {
				return fmt.Errorf("registry namespace (%s) still exists", rs.Primary.Attributes["registry_namespace_id"])
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
