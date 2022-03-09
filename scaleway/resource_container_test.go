package scaleway

import (
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/stretchr/testify/require"
	"net/http"
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
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
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
						name = "test-for-container-as-a-service-public"
						description = "test registry namespace for container as a service"
						is_public = true
					}

					resource scaleway_container_namespace main {
						name = "test-cr-ns"
						description = "test container namespace namespace"
					}

					resource scaleway_container main {
						name = "my-container-01"
						description = "test container"
						namespace_id = scaleway_container_namespace.main.id
						registry_image = scaleway_registry_namespace.main.endpoint
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-01"),
					resource.TestCheckResourceAttr("scaleway_container.main", "description", "test container"),
				),
			},
		},
	})
}

func TestAccScalewayContainer_WithIMG(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	addImageToRegistry := func(tt *TestTools, n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("not found: %s", n)
			}
			api, region, id, err := registryAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return nil
			}

			ns, err := api.GetNamespace(&registry.GetNamespaceRequest{
				NamespaceID: id,
				Region:      region,
			})
			if err != nil {
				return err
			}

			pool, err := dockertest.NewPool(ns.Endpoint)
			require.NoError(t, err, "could not connect to Docker")

			resource, err := pool.Run("docker-gs-ping", "latest", []string{})
			require.NoError(t, err, "could not start container")

			t.Cleanup(func() {
				require.NoError(t, pool.Purge(resource), "failed to remove container")
			})

			var resp *http.Response

			err = pool.Retry(func() error {
				resp, err = http.Get(fmt.Sprint("http://localhost:", resource.GetPort("8080/tcp"), "/"))
				if err != nil {
					t.Log("container not ready, waiting...")
					return err
				}
				return nil
			})
			require.NoError(t, err, "HTTP error")
			defer resp.Body.Close()

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_registry_namespace main {
						name = "test-for-container-as-a-service"
						description = "test registry namespace for container as a service"
						is_public = true
					}

					resource scaleway_container_namespace main {
						name = "test-cr-ns"
						description = "test container"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					addImageToRegistry(tt, "scaleway_registry_namespace.main"),
				),
			},
			{
				Config: `
					resource scaleway_registry_namespace main {
						name = "test-for-container-as-a-service"
						description = "test registry namespace for container as a service"
						is_public = true
					}

					resource scaleway_container_namespace main {
						name = "test-cr-ns"
						description = "test container"
					}

					resource scaleway_container main {
						name = "my-container-02"
						description = "environment variables test"
						namespace_id = scaleway_container_namespace.main.id
						registry_image = "${scaleway_registry_namespace.main.endpoint}/docker-gs-ping"
						port = 9090
						cpu_limit = 140
						memory_limit = 256
						min_scale = 3
						max_scale = 5
						timeout = 600
						max_concurrency = 80
						privacy = private
						protocol = h2c
						redeploy = true

						environment_variables = {
							"foo" = "var"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerExists(tt, "scaleway_container.main"),
					testCheckResourceAttrUUID("scaleway_container_namespace.main", "id"),
					testCheckResourceAttrUUID("scaleway_container.main", "id"),
					resource.TestCheckResourceAttr("scaleway_container.main", "name", "my-container-02"),
					resource.TestCheckResourceAttr("scaleway_container.main", "port", "9090"),
					resource.TestCheckResourceAttr("scaleway_container.main", "cpu_limit", "140"),
					resource.TestCheckResourceAttr("scaleway_container.main", "memory_limit", "256"),
					resource.TestCheckResourceAttr("scaleway_container.main", "min_scale", "3"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_scale", "5"),
					resource.TestCheckResourceAttr("scaleway_container.main", "timeout", "600"),
					resource.TestCheckResourceAttr("scaleway_container.main", "max_concurrency", "80"),
					resource.TestCheckResourceAttr("scaleway_container.main", "redeploy", "true"),
					resource.TestCheckResourceAttr("scaleway_container.main", "privacy", container.ContainerPrivacyPrivate.String()),
					resource.TestCheckResourceAttr("scaleway_container.main", "protocol", container.ContainerProtocolH2c.String()),
				),
			},
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
