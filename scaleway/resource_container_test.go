package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
)

// We must add registry for the sweeper
func TestAccScalewayContainer_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "name"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "registry_image"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "domain_name"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "max_concurrency"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "domain_name"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "protocol"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "cpu_limit"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "timeout"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "memory_limit"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "max_scale"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "min_scale"),
					resource.TestCheckResourceAttrSet("scaleway_container.main", "privacy"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {
					}

					resource scaleway_container main {
						name = "my-container-tf"
						namespace_id = scaleway_container_namespace.main.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-tf"),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "8080"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "70"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "128"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "0"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "20"),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "300"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_concurrency", "50"),
					resource.TestCheckResourceAttr("scaleway_container.main", "redeploy", "false"),
					resource.TestCheckResourceAttr("scaleway_container.main", "privacy", container.ContainerPrivacyPublic.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", container.ContainerProtocolHTTP1.String()),
				),
			},
			{
				Config: `
					resource scaleway_registry_namespace main {
						name = "test-for-cotainer-as-a-service"
						description = "test registry namespace for container as a service"
						is_public = true
					}

					resource scaleway_container_namespace main {
						name = "test-cr-ns-01"
						description = "test container namespace 01"
					}

					resource scaleway_container main {
						name = "my-container-tf"
						namespace_id = scaleway_container_namespace.main.id
						registry_image = scaleway_registry_namespace.main.endpoint
					}
						`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerNamespaceExists(tt, "scaleway_container_namespace.main"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "description", "test container namespace 01"),
					resource.TestCheckResourceAttr("scaleway_container_namespace.main", "name", "test-cr-ns-01"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
				),
			},
			/*			{
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
						},*/
		},
	})
}

func testAccCheckScalewayContainerExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource container not found: %s", n)
		}

		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return nil
		}

		_, err = api.GetContainer(&container.GetContainerRequest{
			ContainerID: id,
			Region:      region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayContainerDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_namespace" {
				continue
			}

			api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteContainer(&container.DeleteContainerRequest{
				ContainerID: id,
				Region:      region,
			})

			if err == nil {
				return fmt.Errorf("container (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
